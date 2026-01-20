package orderbook

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/binary-jerry/opinion-sdk/common"
	"github.com/shopspring/decimal"
)

// SnapshotFetcher defines the interface for fetching orderbook snapshots
type SnapshotFetcher interface {
	GetOrderbookSnapshot(ctx context.Context, tokenID string) (bids, asks []PriceLevel, err error)
}

// SDK provides a high-level interface for managing orderbooks via WebSocket
// Uses callback pattern (like Polymarket SDK) - simplified event handling
type SDK struct {
	config          *Config
	wsClient        *WSClient
	manager         *Manager
	snapshotFetcher SnapshotFetcher
	rateLimiter     *common.RateLimiter

	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	started          bool
	validationCancel context.CancelFunc
}

// NewSDK creates a new orderbook SDK instance
func NewSDK(config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}

	wsClient := NewWSClient(config)
	manager := NewManager(wsClient)

	// Create rate limiter: 10 requests/second, burst of 5
	rateLimiter := common.NewRateLimiter(10, 5)

	sdk := &SDK{
		config:      config,
		wsClient:    wsClient,
		manager:     manager,
		rateLimiter: rateLimiter,
	}

	// Set up reconnect handler on manager
	// This replaces the old handleReconnectEvents goroutine
	manager.SetReconnectHandler(sdk.refreshAllSnapshots)

	return sdk
}

// SetSnapshotFetcher sets the snapshot fetcher
func (s *SDK) SetSnapshotFetcher(fetcher SnapshotFetcher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshotFetcher = fetcher
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

	s.mu.Lock()
	s.started = true
	s.mu.Unlock()

	return nil
}

// refreshAllSnapshots refreshes snapshots for all registered market token pairs
// Called by Manager on reconnect via callback
// 注意：orderbooks 已在断开连接时被 clearAllOrderBooks 清空
func (s *SDK) refreshAllSnapshots() {
	s.mu.RLock()
	fetcher := s.snapshotFetcher
	s.mu.RUnlock()

	if fetcher == nil {
		log.Printf("[OrderbookSDK] Cannot refresh snapshots: no snapshot fetcher set")
		return
	}

	pairs := s.manager.GetAllMarketTokenPairs()
	if len(pairs) == 0 {
		return
	}

	log.Printf("[OrderbookSDK] Refreshing snapshots for %d markets after reconnection", len(pairs))

	for _, pair := range pairs {
		// Refresh snapshots asynchronously
		go func(p *MarketTokenPair) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Fetch YES snapshot with rate limiting
			if err := s.rateLimiter.Wait(ctx); err != nil {
				log.Printf("[OrderbookSDK] Rate limit wait failed for YES token %s: %v", p.YesTokenID, err)
				return
			}
			yesBids, yesAsks, err := fetcher.GetOrderbookSnapshot(ctx, p.YesTokenID)
			if err != nil {
				log.Printf("[OrderbookSDK] Failed to refresh YES snapshot for %s: %v", p.YesTokenID, err)
			} else {
				s.manager.ApplySnapshot(p.MarketID, p.YesTokenID, yesBids, yesAsks, time.Now().UnixMilli())
				log.Printf("[OrderbookSDK] Refreshed YES snapshot for %s", p.YesTokenID)
			}

			// Fetch NO snapshot with rate limiting
			if err := s.rateLimiter.Wait(ctx); err != nil {
				log.Printf("[OrderbookSDK] Rate limit wait failed for NO token %s: %v", p.NoTokenID, err)
				return
			}
			noBids, noAsks, err := fetcher.GetOrderbookSnapshot(ctx, p.NoTokenID)
			if err != nil {
				log.Printf("[OrderbookSDK] Failed to refresh NO snapshot for %s: %v", p.NoTokenID, err)
			} else {
				s.manager.ApplySnapshot(p.MarketID, p.NoTokenID, noBids, noAsks, time.Now().UnixMilli())
				log.Printf("[OrderbookSDK] Refreshed NO snapshot for %s", p.NoTokenID)
			}
		}(pair)
	}
}

// Stop disconnects and stops processing events
func (s *SDK) Stop() error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false

	// Stop periodic validation if running
	if s.validationCancel != nil {
		s.validationCancel()
		s.validationCancel = nil
	}
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

// SubscribeMarketWithSnapshot subscribes to a market with automatic snapshot initialization
// 流程：订阅 WS -> 获取 snapshot -> 应用 snapshot
// 在 snapshot 应用前收到的 WS 消息会被忽略（因为 !initialized）
// 定期校验机制可以修正短暂的数据不一致
func (s *SDK) SubscribeMarketWithSnapshot(ctx context.Context, marketID int64, yesTokenID, noTokenID string) error {
	s.mu.RLock()
	fetcher := s.snapshotFetcher
	s.mu.RUnlock()

	if fetcher == nil {
		return fmt.Errorf("snapshot fetcher not set, call SetSnapshotFetcher first")
	}

	// 1. Register token pair for mirror sync
	s.manager.RegisterMarketTokenPair(&MarketTokenPair{
		MarketID:   marketID,
		YesTokenID: yesTokenID,
		NoTokenID:  noTokenID,
	})

	// 2. Subscribe to WebSocket (消息会因 !initialized 被忽略)
	if err := s.wsClient.SubscribeDepth(marketID); err != nil {
		return fmt.Errorf("failed to subscribe to depth updates: %w", err)
	}

	// 3. Fetch snapshots in parallel with rate limiting
	var wg sync.WaitGroup
	var yesErr, noErr error
	var yesBids, yesAsks, noBids, noAsks []PriceLevel

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := s.rateLimiter.Wait(ctx); err != nil {
			yesErr = err
			return
		}
		yesBids, yesAsks, yesErr = fetcher.GetOrderbookSnapshot(ctx, yesTokenID)
	}()

	go func() {
		defer wg.Done()
		if err := s.rateLimiter.Wait(ctx); err != nil {
			noErr = err
			return
		}
		noBids, noAsks, noErr = fetcher.GetOrderbookSnapshot(ctx, noTokenID)
	}()

	wg.Wait()

	// 4. Apply snapshots (之后的 WS 消息会被正常处理)
	timestamp := time.Now().UnixMilli()

	if yesErr != nil {
		log.Printf("[OrderbookSDK] Warning: failed to get YES snapshot for token %s: %v", yesTokenID, yesErr)
	} else {
		s.manager.ApplySnapshot(marketID, yesTokenID, yesBids, yesAsks, timestamp)
		log.Printf("[OrderbookSDK] Applied YES snapshot for token %s: %d bids, %d asks", yesTokenID, len(yesBids), len(yesAsks))
	}

	if noErr != nil {
		log.Printf("[OrderbookSDK] Warning: failed to get NO snapshot for token %s: %v", noTokenID, noErr)
	} else {
		s.manager.ApplySnapshot(marketID, noTokenID, noBids, noAsks, timestamp)
		log.Printf("[OrderbookSDK] Applied NO snapshot for token %s: %d bids, %d asks", noTokenID, len(noBids), len(noAsks))
	}

	return nil
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

// Updates returns the update notification channel (like Polymarket SDK)
func (s *SDK) Updates() <-chan OrderBookUpdate {
	return s.manager.Updates()
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
func (s *SDK) ApplySnapshot(marketID int64, tokenID string, bids, asks []PriceLevel, timestamp int64) {
	s.manager.ApplySnapshot(marketID, tokenID, bids, asks, timestamp)
}

// IsOrderBookInitialized returns whether the orderbook for a token has received a snapshot
func (s *SDK) IsOrderBookInitialized(tokenID string) bool {
	return s.manager.IsOrderBookInitialized(tokenID)
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

// String returns a string representation of the SDK state
func (s *SDK) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("SDK{started=%v, state=%s, orderbooks=%d}",
		s.started, s.wsClient.State(), s.manager.GetOrderBookCount())
}

// GetHealthStatus returns detailed health status for all orderbooks
func (s *SDK) GetHealthStatus() *HealthStatus {
	return s.manager.GetHealthStatus()
}

// StartPeriodicValidation starts periodic orderbook validation
func (s *SDK) StartPeriodicValidation(interval time.Duration) {
	s.mu.Lock()
	if s.validationCancel != nil {
		s.mu.Unlock()
		return // Already running
	}

	if !s.started || s.ctx == nil {
		s.mu.Unlock()
		log.Printf("[OrderbookSDK] Warning: StartPeriodicValidation called before Start(), ignoring")
		return
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.validationCancel = cancel
	s.mu.Unlock()

	go s.runPeriodicValidation(ctx, interval)
	log.Printf("[OrderbookSDK] Started periodic validation with interval %v", interval)
}

// StopPeriodicValidation stops the periodic validation
func (s *SDK) StopPeriodicValidation() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.validationCancel != nil {
		s.validationCancel()
		s.validationCancel = nil
		log.Printf("[OrderbookSDK] Stopped periodic validation")
	}
}

// runPeriodicValidation runs the validation loop
func (s *SDK) runPeriodicValidation(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.validateAllOrderBooks()
		}
	}
}

// validateAllOrderBooks validates and corrects all registered orderbooks
func (s *SDK) validateAllOrderBooks() {
	s.mu.RLock()
	fetcher := s.snapshotFetcher
	s.mu.RUnlock()

	if fetcher == nil {
		return
	}

	pairs := s.manager.GetAllMarketTokenPairs()
	if len(pairs) == 0 {
		return
	}

	for _, pair := range pairs {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Validate YES token
		if result := s.ValidateOrderBook(ctx, pair.MarketID, pair.YesTokenID); result.Error != nil {
			log.Printf("[OrderbookSDK] Validation error for YES %s: %v", pair.YesTokenID, result.Error)
		} else if !result.Valid {
			log.Printf("[OrderbookSDK] Corrected YES orderbook %s: bids diff=%d, asks diff=%d",
				pair.YesTokenID, result.BidDifference, result.AskDifference)
		}

		// Validate NO token
		if result := s.ValidateOrderBook(ctx, pair.MarketID, pair.NoTokenID); result.Error != nil {
			log.Printf("[OrderbookSDK] Validation error for NO %s: %v", pair.NoTokenID, result.Error)
		} else if !result.Valid {
			log.Printf("[OrderbookSDK] Corrected NO orderbook %s: bids diff=%d, asks diff=%d",
				pair.NoTokenID, result.BidDifference, result.AskDifference)
		}

		cancel()
	}
}

// ValidateOrderBook validates a single orderbook against the API snapshot
func (s *SDK) ValidateOrderBook(ctx context.Context, marketID int64, tokenID string) *ValidationResult {
	result := &ValidationResult{TokenID: tokenID}

	s.mu.RLock()
	fetcher := s.snapshotFetcher
	s.mu.RUnlock()

	if fetcher == nil {
		result.Error = fmt.Errorf("snapshot fetcher not set")
		return result
	}

	// Rate limit the API call
	if err := s.rateLimiter.Wait(ctx); err != nil {
		result.Error = fmt.Errorf("rate limit: %w", err)
		return result
	}

	// Fetch fresh snapshot from API
	apiBids, apiAsks, err := fetcher.GetOrderbookSnapshot(ctx, tokenID)
	if err != nil {
		result.Error = fmt.Errorf("fetch snapshot: %w", err)
		return result
	}

	// Get local orderbook state
	ob := s.manager.GetOrderBook(tokenID)
	if ob == nil || !ob.IsInitialized() {
		// Not initialized, apply snapshot
		s.manager.ApplySnapshot(marketID, tokenID, apiBids, apiAsks, time.Now().UnixMilli())
		result.Corrected = true
		return result
	}

	// Compare local with API
	localBids := ob.GetAllBids()
	localAsks := ob.GetAllAsks()

	result.BidDifference = s.countDifferences(localBids, apiBids)
	result.AskDifference = s.countDifferences(localAsks, apiAsks)

	// If differences found, apply the fresh snapshot
	if result.BidDifference > 0 || result.AskDifference > 0 {
		s.manager.ApplySnapshot(marketID, tokenID, apiBids, apiAsks, time.Now().UnixMilli())
		result.Corrected = true
	} else {
		result.Valid = true
	}

	return result
}

// countDifferences counts the number of price level differences
func (s *SDK) countDifferences(local []OrderSummary, api []PriceLevel) int {
	differences := 0

	// Create map from local orders
	localMap := make(map[string]decimal.Decimal)
	for _, order := range local {
		localMap[order.Price.String()] = order.Size
	}

	// Create map from API orders
	apiMap := make(map[string]decimal.Decimal)
	for _, level := range api {
		apiMap[level.Price.String()] = level.Size
	}

	// Count differences
	for price, localSize := range localMap {
		if apiSize, exists := apiMap[price]; !exists || !apiSize.Equal(localSize) {
			differences++
		}
	}

	// Check for levels in API that aren't in local
	for price := range apiMap {
		if _, exists := localMap[price]; !exists {
			differences++
		}
	}

	return differences
}
