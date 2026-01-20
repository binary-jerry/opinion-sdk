package orderbook

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// OrderBookUpdate represents an orderbook update notification (like Polymarket SDK)
// Only contains metadata, not the full orderbook data
type OrderBookUpdate struct {
	TokenID   string
	EventType string // "depth_update", "snapshot"
	Timestamp int64
}

// Manager manages orderbooks for multiple markets/tokens
// Uses callback pattern (like Polymarket SDK) - no eventChan
type Manager struct {
	mu               sync.RWMutex
	orderbooks       map[string]*OrderBook      // tokenID -> orderbook
	prices           map[string]decimal.Decimal // tokenID -> last price
	marketTokenPairs map[int64]*MarketTokenPair // marketID -> token pair (for mirror sync)
	tokenToMarket    map[string]int64           // tokenID -> marketID (reverse lookup)
	wsClient         *WSClient

	// Update notification channel (like Polymarket SDK)
	// Only sends notifications, not full events
	updateChan chan OrderBookUpdate

	// Snapshot refresh callback (called on reconnect)
	onReconnect func()

	closeChan chan struct{}
	closeOnce sync.Once
}

// NewManager creates a new orderbook manager
func NewManager(wsClient *WSClient) *Manager {
	m := &Manager{
		orderbooks:       make(map[string]*OrderBook),
		prices:           make(map[string]decimal.Decimal),
		marketTokenPairs: make(map[int64]*MarketTokenPair),
		tokenToMarket:    make(map[string]int64),
		wsClient:         wsClient,
		updateChan:       make(chan OrderBookUpdate, 100),
		closeChan:        make(chan struct{}),
	}

	// Set up callbacks on WSClient (like Polymarket SDK)
	wsClient.SetMessageHandler(m.handleMessage)
	wsClient.SetStateChangeHandler(m.handleStateChange)

	return m
}

// SetReconnectHandler sets the callback for reconnection events
// This allows SDK to refresh snapshots on reconnect
func (m *Manager) SetReconnectHandler(handler func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onReconnect = handler
}

// Updates returns the update notification channel (like Polymarket SDK)
func (m *Manager) Updates() <-chan OrderBookUpdate {
	return m.updateChan
}

// Stop stops the manager
func (m *Manager) Stop() {
	m.closeOnce.Do(func() {
		close(m.closeChan)
		close(m.updateChan)
	})
}

// handleStateChange handles WebSocket state changes (callback from WSClient)
func (m *Manager) handleStateChange(state ConnectionState) {
	switch state {
	case StateConnected:
		// Connection established, trigger reconnect handler if set
		m.mu.RLock()
		handler := m.onReconnect
		m.mu.RUnlock()

		if handler != nil {
			// Run in goroutine to avoid blocking
			go handler()
		}

	case StateDisconnected, StateReconnecting:
		// Connection lost - clear all orderbooks to ensure data consistency
		m.clearAllOrderBooks()
	}
}

// handleMessage handles WebSocket messages (callback from WSClient)
func (m *Manager) handleMessage(data []byte) {
	// First try to parse as heartbeat response: {"code": 200, "message": "HEARTBEAT"}
	var heartbeatResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &heartbeatResp); err == nil {
		if heartbeatResp.Code == 200 && heartbeatResp.Message == "HEARTBEAT" {
			m.wsClient.UpdateHeartbeatResponse()
			return
		}
	}

	// Parse as regular message
	var msg struct {
		MsgType string `json:"msgType"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[Manager] failed to parse message: %v", err)
		return
	}

	switch msg.MsgType {
	case ChannelDepthDiff:
		var diff SingleDepthDiffMessage
		if err := json.Unmarshal(data, &diff); err == nil {
			m.handleSingleDepthDiff(&diff)
		}
	case ChannelLastPrice:
		var price LastPriceMessage
		if err := json.Unmarshal(data, &price); err == nil {
			m.handlePriceUpdate(&price)
		}
	}
}

// handleSingleDepthDiff processes a single price level update with mirror sync
// 如果 orderbook 未初始化（未收到 snapshot），消息会被忽略
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

	currentTokenID := msg.TokenID
	var mirrorTokenID string

	if pair != nil {
		if currentTokenID == pair.YesTokenID {
			mirrorTokenID = pair.NoTokenID
		} else if currentTokenID == pair.NoTokenID {
			mirrorTokenID = pair.YesTokenID
		}
	}

	if currentTokenID == "" {
		return
	}

	// Get orderbook for current token (must be initialized)
	ob := m.orderbooks[currentTokenID]
	if ob == nil || !ob.IsInitialized() {
		return // Ignore messages before snapshot
	}

	timestamp := time.Now().UnixMilli()

	// Apply update to current token's orderbook
	m.applyPriceUpdate(currentTokenID, msg.Side, price, size)

	// Send update notification
	m.sendUpdate(OrderBookUpdate{
		TokenID:   currentTokenID,
		EventType: "depth_update",
		Timestamp: timestamp,
	})

	// Mirror sync: update the other token's orderbook
	if mirrorTokenID != "" {
		mirrorOb := m.orderbooks[mirrorTokenID]
		if mirrorOb == nil || !mirrorOb.IsInitialized() {
			return // Mirror token not initialized
		}

		mirrorPrice := decimal.NewFromInt(1).Sub(price)
		mirrorSide := m.getMirrorSide(msg.Side)

		m.applyPriceUpdate(mirrorTokenID, mirrorSide, mirrorPrice, size)

		// Send update notification for mirror token
		m.sendUpdate(OrderBookUpdate{
			TokenID:   mirrorTokenID,
			EventType: "depth_update",
			Timestamp: timestamp,
		})
	}
}

// applyPriceUpdate applies a single price level update to an orderbook
func (m *Manager) applyPriceUpdate(tokenID, side string, price, size decimal.Decimal) {
	ob, exists := m.orderbooks[tokenID]
	if !exists || !ob.IsInitialized() {
		return
	}

	timestamp := time.Now().UnixMilli()
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

func (m *Manager) handlePriceUpdate(msg *LastPriceMessage) {
	m.mu.Lock()
	m.prices[msg.TokenID] = msg.Price
	m.mu.Unlock()
}

// sendUpdate sends an update notification (non-blocking)
func (m *Manager) sendUpdate(update OrderBookUpdate) {
	select {
	case m.updateChan <- update:
	default:
		// Channel full, drop oldest and retry
		select {
		case <-m.updateChan:
		default:
		}
		select {
		case m.updateChan <- update:
		default:
		}
	}
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
func (m *Manager) ApplySnapshot(marketID int64, tokenID string, bids, asks []PriceLevel, timestamp int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderbooks[tokenID]
	if !exists {
		ob = NewOrderBook(marketID, tokenID)
		m.orderbooks[tokenID] = ob
	}

	ob.ApplySnapshot(bids, asks, 0, timestamp)

	// Send snapshot notification
	m.sendUpdate(OrderBookUpdate{
		TokenID:   tokenID,
		EventType: "snapshot",
		Timestamp: timestamp,
	})
}

// HealthCheck returns true if the manager is healthy
func (m *Manager) HealthCheck() bool {
	if m.wsClient.State() != StateConnected {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.orderbooks) == 0 {
		return true
	}

	now := time.Now().UnixMilli()
	staleThreshold := int64(5 * 60 * 1000) // 5 minutes

	for _, ob := range m.orderbooks {
		if !ob.IsInitialized() {
			continue
		}
		if now-ob.Timestamp() > staleThreshold {
			return false
		}
		return true
	}

	return false
}

// GetHealthStatus returns detailed health status for all orderbooks
func (m *Manager) GetHealthStatus() *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UnixMilli()
	staleThreshold := int64(5 * 60 * 1000)

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
	TokenID   string
	BestBid   decimal.Decimal
	BestAsk   decimal.Decimal
	MidPrice  decimal.Decimal
	Spread    decimal.Decimal
	SpreadBps decimal.Decimal
	BidCount  int
	AskCount  int
	LastPrice decimal.Decimal
	UpdatedAt time.Time
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

// clearAllOrderBooks clears all orderbooks (called on disconnect)
func (m *Manager) clearAllOrderBooks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ob := range m.orderbooks {
		ob.Clear()
	}
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

// RegisterMarketTokenPair registers a YES/NO token pair for a market
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
