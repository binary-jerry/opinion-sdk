package orderbook

import (
	"context"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
)

// SDK provides a high-level interface for managing orderbooks via WebSocket
type SDK struct {
	config   *Config
	wsClient *WSClient
	manager  *Manager

	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	started     bool
}

// NewSDK creates a new orderbook SDK instance
func NewSDK(config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}

	wsClient := NewWSClient(config)
	manager := NewManager(wsClient)

	return &SDK{
		config:   config,
		wsClient: wsClient,
		manager:  manager,
	}
}

// Start connects to the WebSocket and begins processing events
func (s *SDK) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	// Connect WebSocket
	if err := s.wsClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Create internal context
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Start manager
	if err := s.manager.Start(s.ctx); err != nil {
		s.wsClient.Disconnect()
		return fmt.Errorf("failed to start manager: %w", err)
	}

	s.mu.Lock()
	s.started = true
	s.mu.Unlock()

	return nil
}

// Stop disconnects and stops processing events
func (s *SDK) Stop() error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false
	s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	s.manager.Stop()
	return s.wsClient.Close()
}

// Subscribe subscribes to orderbook updates for a market
func (s *SDK) Subscribe(marketID int64) error {
	return s.wsClient.SubscribeDepth(marketID)
}

// SubscribePrice subscribes to price updates for a market
func (s *SDK) SubscribePrice(marketID int64) error {
	return s.wsClient.SubscribePrice(marketID)
}

// SubscribeTrade subscribes to trade updates for a market
func (s *SDK) SubscribeTrade(marketID int64) error {
	return s.wsClient.SubscribeTrade(marketID)
}

// SubscribeAll subscribes to all channels for a market
func (s *SDK) SubscribeAll(marketID int64) error {
	if err := s.wsClient.SubscribeDepth(marketID); err != nil {
		return err
	}
	if err := s.wsClient.SubscribePrice(marketID); err != nil {
		return err
	}
	return s.wsClient.SubscribeTrade(marketID)
}

// Unsubscribe unsubscribes from orderbook updates for a market
func (s *SDK) Unsubscribe(marketID int64) error {
	return s.wsClient.Unsubscribe(ChannelDepthDiff, marketID)
}

// UnsubscribePrice unsubscribes from price updates for a market
func (s *SDK) UnsubscribePrice(marketID int64) error {
	return s.wsClient.Unsubscribe(ChannelLastPrice, marketID)
}

// UnsubscribeTrade unsubscribes from trade updates for a market
func (s *SDK) UnsubscribeTrade(marketID int64) error {
	return s.wsClient.Unsubscribe(ChannelLastTrade, marketID)
}

// UnsubscribeAll unsubscribes from all channels for a market
func (s *SDK) UnsubscribeAll(marketID int64) error {
	s.wsClient.Unsubscribe(ChannelDepthDiff, marketID)
	s.wsClient.Unsubscribe(ChannelLastPrice, marketID)
	s.wsClient.Unsubscribe(ChannelLastTrade, marketID)
	return nil
}

// SubscribeOrderUpdates subscribes to user order updates
func (s *SDK) SubscribeOrderUpdates() error {
	return s.wsClient.SubscribeOrderUpdates()
}

// SubscribeTradeRecords subscribes to user trade records
func (s *SDK) SubscribeTradeRecords() error {
	return s.wsClient.SubscribeTradeRecords()
}

// Events returns the event channel
func (s *SDK) Events() <-chan *Event {
	return s.manager.Events()
}

// GetBBO returns the best bid and offer for a token
func (s *SDK) GetBBO(tokenID string) *BBO {
	return s.manager.GetBBO(tokenID)
}

// GetDepth returns orderbook depth for a token
func (s *SDK) GetDepth(tokenID string, levels int) *Depth {
	return s.manager.GetDepth(tokenID, levels)
}

// GetLastPrice returns the last known price for a token
func (s *SDK) GetLastPrice(tokenID string) decimal.Decimal {
	return s.manager.GetLastPrice(tokenID)
}

// GetMidPrice returns the mid price for a token
func (s *SDK) GetMidPrice(tokenID string) decimal.Decimal {
	return s.manager.GetMidPrice(tokenID)
}

// GetSpread returns the spread for a token
func (s *SDK) GetSpread(tokenID string) decimal.Decimal {
	return s.manager.GetSpread(tokenID)
}

// GetSpreadBps returns the spread in basis points for a token
func (s *SDK) GetSpreadBps(tokenID string) decimal.Decimal {
	return s.manager.GetSpreadBps(tokenID)
}

// ScanBidsAbove returns all bids at or above the given price
func (s *SDK) ScanBidsAbove(tokenID string, price decimal.Decimal) []OrderSummary {
	return s.manager.ScanBidsAbove(tokenID, price)
}

// ScanAsksBelow returns all asks at or below the given price
func (s *SDK) ScanAsksBelow(tokenID string, price decimal.Decimal) []OrderSummary {
	return s.manager.ScanAsksBelow(tokenID, price)
}

// GetAllBids returns all bids for a token
func (s *SDK) GetAllBids(tokenID string) []OrderSummary {
	return s.manager.GetAllBids(tokenID)
}

// GetAllAsks returns all asks for a token
func (s *SDK) GetAllAsks(tokenID string) []OrderSummary {
	return s.manager.GetAllAsks(tokenID)
}

// GetOrderBook returns the full orderbook for a token
func (s *SDK) GetOrderBook(tokenID string) *OrderBook {
	return s.manager.GetOrderBook(tokenID)
}

// GetMarketSummary returns a summary for a token
func (s *SDK) GetMarketSummary(tokenID string) *MarketSummary {
	return s.manager.GetMarketSummary(tokenID)
}

// GetAllMarketSummaries returns summaries for all subscribed tokens
func (s *SDK) GetAllMarketSummaries() []*MarketSummary {
	return s.manager.GetAllMarketSummaries()
}

// GetConnectionState returns the current WebSocket connection state
func (s *SDK) GetConnectionState() ConnectionState {
	return s.wsClient.State()
}

// IsConnected returns true if connected to WebSocket
func (s *SDK) IsConnected() bool {
	return s.wsClient.State() == StateConnected
}

// GetSubscriptions returns active subscriptions
func (s *SDK) GetSubscriptions() []*Subscription {
	return s.wsClient.GetSubscriptions()
}

// GetTokenIDs returns all token IDs with orderbooks
func (s *SDK) GetTokenIDs() []string {
	return s.manager.GetTokenIDs()
}

// ClearOrderBook clears the orderbook for a token
func (s *SDK) ClearOrderBook(tokenID string) {
	s.manager.ClearOrderBook(tokenID)
}

// ApplySnapshot applies a full orderbook snapshot
func (s *SDK) ApplySnapshot(marketID int64, tokenID string, bids, asks []PriceLevel, sequence int64) {
	s.manager.ApplySnapshot(marketID, tokenID, bids, asks, sequence)
}

// HealthCheck returns true if the SDK is healthy
func (s *SDK) HealthCheck() bool {
	s.mu.RLock()
	started := s.started
	s.mu.RUnlock()

	return started && s.manager.HealthCheck()
}

// Config returns the SDK configuration
func (s *SDK) Config() *Config {
	return s.config
}

// GetWSClient returns the underlying WebSocket client (for advanced usage)
func (s *SDK) GetWSClient() *WSClient {
	return s.wsClient
}

// GetManager returns the underlying orderbook manager (for advanced usage)
func (s *SDK) GetManager() *Manager {
	return s.manager
}

// WaitForConnection waits for the WebSocket connection to be established
func (s *SDK) WaitForConnection(ctx context.Context) error {
	events := s.wsClient.Events()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-events:
			if event.Type == EventConnected {
				return nil
			}
			if event.Type == EventError {
				return event.Error
			}
		}
	}
}

// String returns a string representation of the SDK state
func (s *SDK) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("SDK{started=%v, state=%s, orderbooks=%d}",
		s.started, s.wsClient.State(), s.manager.GetOrderBookCount())
}
