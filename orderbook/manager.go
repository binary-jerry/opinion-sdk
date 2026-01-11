package orderbook

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Manager manages orderbooks for multiple markets/tokens
type Manager struct {
	mu               sync.RWMutex
	orderbooks       map[string]*OrderBook           // tokenID -> orderbook
	prices           map[string]decimal.Decimal      // tokenID -> last price
	pendingDiffs     map[string][]*SingleDepthDiffMessage // tokenID -> pending diffs before snapshot
	marketTokenPairs map[int64]*MarketTokenPair      // marketID -> token pair (for mirror sync)
	tokenToMarket    map[string]int64                // tokenID -> marketID (reverse lookup)
	wsClient         *WSClient
	eventChan        chan *Event
	doneChan         chan struct{}
}

// NewManager creates a new orderbook manager
func NewManager(wsClient *WSClient) *Manager {
	return &Manager{
		orderbooks:       make(map[string]*OrderBook),
		prices:           make(map[string]decimal.Decimal),
		pendingDiffs:     make(map[string][]*SingleDepthDiffMessage),
		marketTokenPairs: make(map[int64]*MarketTokenPair),
		tokenToMarket:    make(map[string]int64),
		wsClient:         wsClient,
		eventChan:        make(chan *Event, 100),
		doneChan:         make(chan struct{}),
	}
}

// Start begins processing events from the WebSocket client
func (m *Manager) Start(ctx context.Context) error {
	go m.processEvents(ctx)
	return nil
}

// Stop stops the manager
func (m *Manager) Stop() {
	close(m.doneChan)
}

// Events returns the event channel for external consumers
func (m *Manager) Events() <-chan *Event {
	return m.eventChan
}

func (m *Manager) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.doneChan:
			return
		case event := <-m.wsClient.Events():
			m.handleEvent(event)
			// Forward event to external consumers
			select {
			case m.eventChan <- event:
			default:
			}
		}
	}
}

func (m *Manager) handleEvent(event *Event) {
	switch event.Type {
	case EventDepthUpdate:
		m.handleDepthUpdate(event)
	case EventPriceUpdate:
		m.handlePriceUpdate(event)
	case EventConnected:
		// Connection established
	case EventDisconnected:
		// Connection lost - clear all orderbooks to ensure data consistency on reconnect
		m.clearAllOrderBooks()
	case EventReconnecting:
		// Attempting to reconnect - clear orderbooks to avoid stale data
		m.clearAllOrderBooks()
	}
}

func (m *Manager) handleDepthUpdate(event *Event) {
	// Handle SingleDepthDiffMessage (actual WebSocket format)
	if singleDiff, ok := event.Data.(*SingleDepthDiffMessage); ok {
		m.handleSingleDepthDiff(singleDiff)
		return
	}

	// Handle legacy DepthDiffMessage format (backward compatibility)
	diff, ok := event.Data.(*DepthDiffMessage)
	if !ok {
		return
	}

	// Validate tokenID to avoid creating invalid orderbooks
	if diff.TokenID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderbooks[diff.TokenID]
	if !exists {
		ob = NewOrderBook(diff.MarketID, diff.TokenID)
		m.orderbooks[diff.TokenID] = ob
	}

	// If orderbook is not initialized, skip (no buffering for legacy format)
	if !ob.IsInitialized() {
		return
	}

	// Apply the diff with current timestamp
	timestamp := time.Now().UnixMilli()
	ob.ApplyDiff(diff.Bids, diff.Asks, diff.Sequence, timestamp)
}

// handleSingleDepthDiff processes a single price level update with mirror sync
func (m *Manager) handleSingleDepthDiff(msg *SingleDepthDiffMessage) {
	price, err := decimal.NewFromString(msg.Price)
	if err != nil {
		return
	}
	size, err := decimal.NewFromString(msg.Size)
	if err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the market token pair for mirror sync
	pair := m.marketTokenPairs[msg.MarketID]

	// Determine current token and mirror token
	var currentTokenID, mirrorTokenID string
	if pair != nil {
		if msg.OutcomeSide == OutcomeSideYES {
			currentTokenID = pair.YesTokenID
			mirrorTokenID = pair.NoTokenID
		} else {
			currentTokenID = pair.NoTokenID
			mirrorTokenID = pair.YesTokenID
		}
	} else {
		// No pair registered, just use the tokenID from message
		currentTokenID = msg.TokenID
	}

	// Validate tokenID to avoid creating invalid orderbooks with empty key
	if currentTokenID == "" {
		return
	}

	// Get or create orderbook for current token
	ob, exists := m.orderbooks[currentTokenID]
	if !exists {
		ob = NewOrderBook(msg.MarketID, currentTokenID)
		m.orderbooks[currentTokenID] = ob
	}

	// If orderbook is not initialized, buffer the diff
	if !ob.IsInitialized() {
		pending := m.pendingDiffs[currentTokenID]
		if len(pending) < 1000 {
			m.pendingDiffs[currentTokenID] = append(pending, msg)
		}
		return
	}

	// Apply update to current token's orderbook
	m.applyPriceUpdate(currentTokenID, msg.Side, price, size)

	// Mirror sync: update the other token's orderbook
	if mirrorTokenID != "" {
		mirrorOb, exists := m.orderbooks[mirrorTokenID]
		if exists && mirrorOb.IsInitialized() {
			// Mirror price: NO = 1 - YES
			mirrorPrice := decimal.NewFromInt(1).Sub(price)
			// Mirror side: asks <-> bids
			mirrorSide := m.getMirrorSide(msg.Side)
			m.applyPriceUpdate(mirrorTokenID, mirrorSide, mirrorPrice, size)
		}
	}
}

// applyPriceUpdate applies a single price level update to an orderbook
func (m *Manager) applyPriceUpdate(tokenID, side string, price, size decimal.Decimal) {
	ob, exists := m.orderbooks[tokenID]
	if !exists || !ob.IsInitialized() {
		return
	}

	timestamp := time.Now().UnixMilli()

	// Create a single-element diff
	diff := PriceLevelDiff{Price: price, Size: size}

	if side == "bids" {
		ob.ApplyDiff([]PriceLevelDiff{diff}, nil, 0, timestamp)
	} else {
		ob.ApplyDiff(nil, []PriceLevelDiff{diff}, 0, timestamp)
	}
}

// getMirrorSide returns the opposite side
func (m *Manager) getMirrorSide(side string) string {
	if side == "asks" {
		return "bids"
	}
	return "asks"
}

func (m *Manager) handlePriceUpdate(event *Event) {
	msg, ok := event.Data.(*LastPriceMessage)
	if !ok {
		return
	}

	m.mu.Lock()
	m.prices[msg.TokenID] = msg.Price
	m.mu.Unlock()
}

// GetOrderBook returns the orderbook for a token
func (m *Manager) GetOrderBook(tokenID string) *OrderBook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.orderbooks[tokenID]
}

// GetOrCreateOrderBook returns an existing orderbook or creates a new one
func (m *Manager) GetOrCreateOrderBook(marketID int64, tokenID string) *OrderBook {
	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderbooks[tokenID]
	if !exists {
		ob = NewOrderBook(marketID, tokenID)
		m.orderbooks[tokenID] = ob
	}
	return ob
}

// GetBBO returns the best bid and offer for a token
func (m *Manager) GetBBO(tokenID string) *BBO {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return &BBO{}
	}
	return ob.GetBBO()
}

// GetDepth returns orderbook depth for a token
func (m *Manager) GetDepth(tokenID string, levels int) *Depth {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return &Depth{
			Bids: []OrderSummary{},
			Asks: []OrderSummary{},
		}
	}
	return ob.GetDepth(levels)
}

// GetLastPrice returns the last known price for a token
func (m *Manager) GetLastPrice(tokenID string) decimal.Decimal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.prices[tokenID]
}

// GetMidPrice returns the mid price for a token
func (m *Manager) GetMidPrice(tokenID string) decimal.Decimal {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return decimal.Zero
	}
	return ob.GetMidPrice()
}

// GetSpread returns the spread for a token
func (m *Manager) GetSpread(tokenID string) decimal.Decimal {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return decimal.Zero
	}
	return ob.GetSpread()
}

// GetSpreadBps returns the spread in basis points for a token
func (m *Manager) GetSpreadBps(tokenID string) decimal.Decimal {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return decimal.Zero
	}
	return ob.GetSpreadBps()
}

// ScanBidsAbove returns all bids at or above the given price for a token
func (m *Manager) ScanBidsAbove(tokenID string, price decimal.Decimal) []OrderSummary {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return []OrderSummary{}
	}
	return ob.ScanBidsAbove(price)
}

// ScanAsksBelow returns all asks at or below the given price for a token
func (m *Manager) ScanAsksBelow(tokenID string, price decimal.Decimal) []OrderSummary {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return []OrderSummary{}
	}
	return ob.ScanAsksBelow(price)
}

// GetAllBids returns all bids for a token sorted by price
func (m *Manager) GetAllBids(tokenID string) []OrderSummary {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return []OrderSummary{}
	}
	return ob.GetAllBids()
}

// GetAllAsks returns all asks for a token sorted by price
func (m *Manager) GetAllAsks(tokenID string) []OrderSummary {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	m.mu.RUnlock()

	if ob == nil {
		return []OrderSummary{}
	}
	return ob.GetAllAsks()
}

// ClearOrderBook clears the orderbook for a token
func (m *Manager) ClearOrderBook(tokenID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ob, exists := m.orderbooks[tokenID]; exists {
		ob.Clear()
	}
}

// RemoveOrderBook removes an orderbook for a token
func (m *Manager) RemoveOrderBook(tokenID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.orderbooks, tokenID)
}

// GetOrderBookCount returns the number of orderbooks being managed
func (m *Manager) GetOrderBookCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.orderbooks)
}

// GetTokenIDs returns all token IDs with orderbooks
func (m *Manager) GetTokenIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.orderbooks))
	for id := range m.orderbooks {
		ids = append(ids, id)
	}
	return ids
}

// ApplySnapshot applies a full orderbook snapshot for a token
// Since there's no sequence in the API, we discard all pending diffs after applying snapshot
func (m *Manager) ApplySnapshot(marketID int64, tokenID string, bids, asks []PriceLevel, timestamp int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderbooks[tokenID]
	if !exists {
		ob = NewOrderBook(marketID, tokenID)
		m.orderbooks[tokenID] = ob
	}

	// Apply snapshot with sequence=0 (not used)
	ob.ApplySnapshot(bids, asks, 0, timestamp)

	// Discard all pending diffs - since there's no sequence, we can't determine order
	// The orderbook will converge to correct state with subsequent updates
	delete(m.pendingDiffs, tokenID)
}

// HealthCheck returns true if the manager is healthy
// Checks: WebSocket connected, at least one initialized orderbook, no stale data
func (m *Manager) HealthCheck() bool {
	if m.wsClient.State() != StateConnected {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// If no orderbooks registered yet, consider healthy
	if len(m.orderbooks) == 0 {
		return true
	}

	now := time.Now().UnixMilli()
	staleThreshold := int64(5 * 60 * 1000) // 5 minutes

	for _, ob := range m.orderbooks {
		if !ob.IsInitialized() {
			continue
		}
		// At least one initialized orderbook, check if it's stale
		if now-ob.Timestamp() > staleThreshold {
			return false
		}
		return true // At least one healthy orderbook
	}

	// No initialized orderbooks
	return false
}

// GetHealthStatus returns detailed health status for all orderbooks
func (m *Manager) GetHealthStatus() *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UnixMilli()
	staleThreshold := int64(5 * 60 * 1000) // 5 minutes

	status := &HealthStatus{
		Connected:      m.wsClient.State() == StateConnected,
		OrderBookCount: len(m.orderbooks),
		OrderBooks:     make([]OrderBookHealth, 0, len(m.orderbooks)),
	}

	for tokenID, ob := range m.orderbooks {
		health := OrderBookHealth{
			TokenID:     tokenID,
			Initialized: ob.IsInitialized(),
			LastUpdate:  ob.Timestamp(),
			BidCount:    ob.BidCount(),
			AskCount:    ob.AskCount(),
		}

		if ob.Timestamp() > 0 {
			health.StaleSeconds = (now - ob.Timestamp()) / 1000
			health.IsStale = now-ob.Timestamp() > staleThreshold
		}

		if !health.Initialized {
			status.UninitializedCount++
		} else if health.IsStale {
			status.StaleCount++
		} else {
			status.HealthyCount++
		}

		status.OrderBooks = append(status.OrderBooks, health)
	}

	return status
}

// GetConnectionState returns the WebSocket connection state
func (m *Manager) GetConnectionState() ConnectionState {
	return m.wsClient.State()
}

// MarketSummary contains summary information for a market
type MarketSummary struct {
	TokenID    string
	BestBid    decimal.Decimal
	BestAsk    decimal.Decimal
	MidPrice   decimal.Decimal
	Spread     decimal.Decimal
	SpreadBps  decimal.Decimal
	BidCount   int
	AskCount   int
	LastPrice  decimal.Decimal
	UpdatedAt  time.Time
}

// GetMarketSummary returns a summary for a token
func (m *Manager) GetMarketSummary(tokenID string) *MarketSummary {
	m.mu.RLock()
	ob := m.orderbooks[tokenID]
	lastPrice := m.prices[tokenID]
	m.mu.RUnlock()

	summary := &MarketSummary{
		TokenID:   tokenID,
		LastPrice: lastPrice,
		UpdatedAt: time.Now(),
	}

	if ob == nil {
		return summary
	}

	bbo := ob.GetBBO()
	if bbo.BestBid != nil {
		summary.BestBid = bbo.BestBid.Price
	}
	if bbo.BestAsk != nil {
		summary.BestAsk = bbo.BestAsk.Price
	}

	summary.MidPrice = ob.GetMidPrice()
	summary.Spread = ob.GetSpread()
	summary.SpreadBps = ob.GetSpreadBps()
	summary.BidCount = ob.BidCount()
	summary.AskCount = ob.AskCount()

	return summary
}

// GetAllMarketSummaries returns summaries for all tokens
func (m *Manager) GetAllMarketSummaries() []*MarketSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]*MarketSummary, 0, len(m.orderbooks))
	for tokenID, ob := range m.orderbooks {
		summary := &MarketSummary{
			TokenID:   tokenID,
			LastPrice: m.prices[tokenID],
			UpdatedAt: time.Now(),
		}

		bbo := ob.GetBBO()
		if bbo.BestBid != nil {
			summary.BestBid = bbo.BestBid.Price
		}
		if bbo.BestAsk != nil {
			summary.BestAsk = bbo.BestAsk.Price
		}

		summary.MidPrice = ob.GetMidPrice()
		summary.Spread = ob.GetSpread()
		summary.SpreadBps = ob.GetSpreadBps()
		summary.BidCount = ob.BidCount()
		summary.AskCount = ob.AskCount()

		summaries = append(summaries, summary)
	}

	return summaries
}

// String returns a string representation of the manager state
func (m *Manager) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("Manager{orderbooks=%d, state=%s}",
		len(m.orderbooks), m.wsClient.State())
}

// clearAllOrderBooks clears all orderbooks and pending diffs
// Called on disconnect/reconnect to ensure data consistency
func (m *Manager) clearAllOrderBooks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ob := range m.orderbooks {
		ob.Clear()
	}
	// Clear all pending diffs
	m.pendingDiffs = make(map[string][]*SingleDepthDiffMessage)
}

// IsOrderBookInitialized returns whether the orderbook for a token is initialized
func (m *Manager) IsOrderBookInitialized(tokenID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ob, exists := m.orderbooks[tokenID]
	if !exists {
		return false
	}
	return ob.IsInitialized()
}

// GetPendingDiffCount returns the number of pending diffs for a token
func (m *Manager) GetPendingDiffCount(tokenID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pendingDiffs[tokenID])
}

// RegisterMarketTokenPair registers a YES/NO token pair for a market
// This enables mirror sync between YES and NO orderbooks
func (m *Manager) RegisterMarketTokenPair(pair *MarketTokenPair) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.marketTokenPairs[pair.MarketID] = pair
	m.tokenToMarket[pair.YesTokenID] = pair.MarketID
	m.tokenToMarket[pair.NoTokenID] = pair.MarketID
}

// GetMarketTokenPair returns the token pair for a market
func (m *Manager) GetMarketTokenPair(marketID int64) *MarketTokenPair {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.marketTokenPairs[marketID]
}

// GetAllMarketTokenPairs returns all registered token pairs
func (m *Manager) GetAllMarketTokenPairs() []*MarketTokenPair {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pairs := make([]*MarketTokenPair, 0, len(m.marketTokenPairs))
	for _, pair := range m.marketTokenPairs {
		pairs = append(pairs, pair)
	}
	return pairs
}

// MarkUninitialized marks an orderbook as uninitialized
// Used when reconnecting to trigger snapshot refresh
func (m *Manager) MarkUninitialized(tokenID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ob, exists := m.orderbooks[tokenID]; exists {
		ob.Clear()
	}
}

// GetMirrorTokenID returns the mirror token ID for a given token
func (m *Manager) GetMirrorTokenID(tokenID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	marketID, exists := m.tokenToMarket[tokenID]
	if !exists {
		return ""
	}

	pair := m.marketTokenPairs[marketID]
	if pair == nil {
		return ""
	}

	if tokenID == pair.YesTokenID {
		return pair.NoTokenID
	}
	return pair.YesTokenID
}
