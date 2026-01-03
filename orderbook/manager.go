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
	mu         sync.RWMutex
	orderbooks map[string]*OrderBook // tokenID -> orderbook
	prices     map[string]decimal.Decimal // tokenID -> last price
	wsClient   *WSClient
	eventChan  chan *Event
	doneChan   chan struct{}
}

// NewManager creates a new orderbook manager
func NewManager(wsClient *WSClient) *Manager {
	return &Manager{
		orderbooks: make(map[string]*OrderBook),
		prices:     make(map[string]decimal.Decimal),
		wsClient:   wsClient,
		eventChan:  make(chan *Event, 100),
		doneChan:   make(chan struct{}),
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
		// Connection lost - orderbooks may be stale
	case EventReconnecting:
		// Attempting to reconnect
	}
}

func (m *Manager) handleDepthUpdate(event *Event) {
	diff, ok := event.Data.(*DepthDiffMessage)
	if !ok {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderbooks[diff.TokenID]
	if !exists {
		ob = NewOrderBook(diff.MarketID, diff.TokenID)
		m.orderbooks[diff.TokenID] = ob
	}

	// Convert PriceLevelDiff to proper format
	ob.ApplyDiff(diff.Bids, diff.Asks, diff.Sequence)
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
func (m *Manager) ApplySnapshot(marketID int64, tokenID string, bids, asks []PriceLevel, sequence int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderbooks[tokenID]
	if !exists {
		ob = NewOrderBook(marketID, tokenID)
		m.orderbooks[tokenID] = ob
	}

	ob.ApplySnapshot(bids, asks, sequence)
}

// HealthCheck returns true if the manager is healthy
func (m *Manager) HealthCheck() bool {
	return m.wsClient.State() == StateConnected
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
