package orderbook

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNewManager(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.GetOrderBookCount() != 0 {
		t.Errorf("Initial orderbook count = %d, want 0", manager.GetOrderBookCount())
	}
}

func TestManagerGetOrCreateOrderBook(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	ob := manager.GetOrCreateOrderBook(123, "token-1")
	if ob == nil {
		t.Fatal("GetOrCreateOrderBook returned nil")
	}
	if ob.MarketID() != 123 {
		t.Errorf("MarketID = %d, want 123", ob.MarketID())
	}
	if ob.TokenID() != "token-1" {
		t.Errorf("TokenID = %s, want token-1", ob.TokenID())
	}

	// Get same orderbook again
	ob2 := manager.GetOrCreateOrderBook(123, "token-1")
	if ob != ob2 {
		t.Error("Should return same orderbook instance")
	}

	if manager.GetOrderBookCount() != 1 {
		t.Errorf("Orderbook count = %d, want 1", manager.GetOrderBookCount())
	}
}

func TestManagerGetOrderBook(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent orderbook
	ob := manager.GetOrderBook("non-existent")
	if ob != nil {
		t.Error("GetOrderBook for non-existent should return nil")
	}

	// Create and get
	manager.GetOrCreateOrderBook(123, "token-1")
	ob = manager.GetOrderBook("token-1")
	if ob == nil {
		t.Fatal("GetOrderBook returned nil after creation")
	}
}

func TestManagerApplySnapshot(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}

	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	ob := manager.GetOrderBook("token-1")
	if ob == nil {
		t.Fatal("Orderbook should exist after snapshot")
	}
	if ob.BidCount() != 1 {
		t.Errorf("BidCount = %d, want 1", ob.BidCount())
	}
	if ob.AskCount() != 1 {
		t.Errorf("AskCount = %d, want 1", ob.AskCount())
	}
}

func TestManagerGetBBO(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent token
	bbo := manager.GetBBO("non-existent")
	if bbo.BestBid != nil || bbo.BestAsk != nil {
		t.Error("BBO for non-existent token should have nil values")
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	bbo = manager.GetBBO("token-1")
	if bbo.BestBid == nil {
		t.Error("BestBid should not be nil")
	}
	if !bbo.BestBid.Price.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("BestBid.Price = %s, want 0.50", bbo.BestBid.Price)
	}
}

func TestManagerGetDepth(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent token
	depth := manager.GetDepth("non-existent", 5)
	if len(depth.Bids) != 0 || len(depth.Asks) != 0 {
		t.Error("Depth for non-existent token should be empty")
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	depth = manager.GetDepth("token-1", 5)
	if len(depth.Bids) != 2 {
		t.Errorf("Bids count = %d, want 2", len(depth.Bids))
	}
	if len(depth.Asks) != 1 {
		t.Errorf("Asks count = %d, want 1", len(depth.Asks))
	}
}

func TestManagerGetMidPrice(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	midPrice := manager.GetMidPrice("non-existent")
	if !midPrice.IsZero() {
		t.Error("MidPrice for non-existent should be zero")
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	midPrice = manager.GetMidPrice("token-1")
	expected := decimal.NewFromFloat(0.55)
	if !midPrice.Equal(expected) {
		t.Errorf("MidPrice = %s, want %s", midPrice, expected)
	}
}

func TestManagerGetSpread(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	spread := manager.GetSpread("non-existent")
	if !spread.IsZero() {
		t.Error("Spread for non-existent should be zero")
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	spread = manager.GetSpread("token-1")
	expected := decimal.NewFromFloat(0.05)
	if !spread.Equal(expected) {
		t.Errorf("Spread = %s, want %s", spread, expected)
	}
}

func TestManagerGetSpreadBps(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	spreadBps := manager.GetSpreadBps("non-existent")
	if !spreadBps.IsZero() {
		t.Error("SpreadBps for non-existent should be zero")
	}

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	spreadBps = manager.GetSpreadBps("token-1")
	if spreadBps.IsZero() {
		t.Error("SpreadBps should not be zero")
	}
}

func TestManagerScanBidsAbove(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	result := manager.ScanBidsAbove("non-existent", decimal.NewFromFloat(0.50))
	if len(result) != 0 {
		t.Error("ScanBidsAbove for non-existent should be empty")
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	manager.ApplySnapshot(123, "token-1", bids, nil, 1)

	result = manager.ScanBidsAbove("token-1", decimal.NewFromFloat(0.50))
	if len(result) != 1 {
		t.Errorf("ScanBidsAbove result count = %d, want 1", len(result))
	}
}

func TestManagerScanAsksBelow(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	result := manager.ScanAsksBelow("non-existent", decimal.NewFromFloat(0.55))
	if len(result) != 0 {
		t.Error("ScanAsksBelow for non-existent should be empty")
	}

	// With data
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}
	manager.ApplySnapshot(123, "token-1", nil, asks, 1)

	result = manager.ScanAsksBelow("token-1", decimal.NewFromFloat(0.55))
	if len(result) != 1 {
		t.Errorf("ScanAsksBelow result count = %d, want 1", len(result))
	}
}

func TestManagerGetAllBidsAsks(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	bids := manager.GetAllBids("non-existent")
	asks := manager.GetAllAsks("non-existent")
	if len(bids) != 0 || len(asks) != 0 {
		t.Error("GetAllBids/Asks for non-existent should be empty")
	}

	// With data
	bidLevels := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	askLevels := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bidLevels, askLevels, 1)

	bids = manager.GetAllBids("token-1")
	asks = manager.GetAllAsks("token-1")
	if len(bids) != 1 {
		t.Errorf("GetAllBids count = %d, want 1", len(bids))
	}
	if len(asks) != 1 {
		t.Errorf("GetAllAsks count = %d, want 1", len(asks))
	}
}

func TestManagerClearOrderBook(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	manager.ApplySnapshot(123, "token-1", bids, nil, 1)

	manager.ClearOrderBook("token-1")

	ob := manager.GetOrderBook("token-1")
	if !ob.IsEmpty() {
		t.Error("Orderbook should be empty after clear")
	}
}

func TestManagerRemoveOrderBook(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	manager.GetOrCreateOrderBook(123, "token-1")
	if manager.GetOrderBookCount() != 1 {
		t.Fatal("Orderbook should exist")
	}

	manager.RemoveOrderBook("token-1")
	if manager.GetOrderBookCount() != 0 {
		t.Errorf("Orderbook count after remove = %d, want 0", manager.GetOrderBookCount())
	}
	if manager.GetOrderBook("token-1") != nil {
		t.Error("Orderbook should not exist after remove")
	}
}

func TestManagerGetTokenIDs(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	manager.GetOrCreateOrderBook(123, "token-1")
	manager.GetOrCreateOrderBook(456, "token-2")

	ids := manager.GetTokenIDs()
	if len(ids) != 2 {
		t.Errorf("TokenIDs count = %d, want 2", len(ids))
	}
}

func TestManagerGetLastPrice(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	price := manager.GetLastPrice("non-existent")
	if !price.IsZero() {
		t.Error("LastPrice for non-existent should be zero")
	}
}

func TestManagerGetMarketSummary(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Non-existent
	summary := manager.GetMarketSummary("non-existent")
	if summary.TokenID != "non-existent" {
		t.Error("Summary should have correct token ID")
	}
	if !summary.BestBid.IsZero() {
		t.Error("BestBid should be zero for non-existent")
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	manager.ApplySnapshot(123, "token-1", bids, asks, 1)

	summary = manager.GetMarketSummary("token-1")
	if !summary.BestBid.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("BestBid = %s, want 0.50", summary.BestBid)
	}
	if !summary.BestAsk.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("BestAsk = %s, want 0.55", summary.BestAsk)
	}
	if summary.BidCount != 1 {
		t.Errorf("BidCount = %d, want 1", summary.BidCount)
	}
	if summary.AskCount != 1 {
		t.Errorf("AskCount = %d, want 1", summary.AskCount)
	}
}

func TestManagerGetAllMarketSummaries(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	manager.ApplySnapshot(123, "token-1", bids, nil, 1)
	manager.ApplySnapshot(456, "token-2", bids, nil, 1)

	summaries := manager.GetAllMarketSummaries()
	if len(summaries) != 2 {
		t.Errorf("Summaries count = %d, want 2", len(summaries))
	}
}

func TestManagerHealthCheck(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Not connected
	if manager.HealthCheck() {
		t.Error("HealthCheck should return false when not connected")
	}
}

func TestManagerGetConnectionState(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	state := manager.GetConnectionState()
	if state != StateDisconnected {
		t.Errorf("ConnectionState = %v, want StateDisconnected", state)
	}
}

func TestManagerString(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	str := manager.String()
	if str == "" {
		t.Error("String() should not return empty")
	}
}

func TestManagerStartStop(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := manager.Start(ctx)
	if err != nil {
		t.Errorf("Start() error: %v", err)
	}

	// Give some time for goroutine to start
	time.Sleep(10 * time.Millisecond)

	manager.Stop()
}

func TestManagerEvents(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	events := manager.Events()
	if events == nil {
		t.Error("Events() should not return nil")
	}
}

func TestManagerHandleDepthUpdate(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Simulate depth update event
	diffMsg := &DepthDiffMessage{
		Channel:  ChannelDepthDiff,
		MarketID: 123,
		TokenID:  "token-1",
		Bids: []PriceLevelDiff{
			{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		},
		Asks:     []PriceLevelDiff{},
		Sequence: 1,
	}

	event := &Event{
		Type:     EventDepthUpdate,
		MarketID: 123,
		TokenID:  "token-1",
		Data:     diffMsg,
	}

	manager.handleEvent(event)

	ob := manager.GetOrderBook("token-1")
	if ob == nil {
		t.Fatal("Orderbook should be created from depth update")
	}
	if ob.BidCount() != 1 {
		t.Errorf("BidCount = %d, want 1", ob.BidCount())
	}
}

func TestManagerHandlePriceUpdate(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	priceMsg := &LastPriceMessage{
		Channel:   ChannelLastPrice,
		MarketID:  123,
		TokenID:   "token-1",
		Price:     decimal.NewFromFloat(0.55),
		Timestamp: 1704067200000,
	}

	event := &Event{
		Type:     EventPriceUpdate,
		MarketID: 123,
		TokenID:  "token-1",
		Data:     priceMsg,
	}

	manager.handleEvent(event)

	price := manager.GetLastPrice("token-1")
	if !price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("LastPrice = %s, want 0.55", price)
	}
}

func TestMarketSummary(t *testing.T) {
	summary := &MarketSummary{
		TokenID:   "token-1",
		BestBid:   decimal.NewFromFloat(0.50),
		BestAsk:   decimal.NewFromFloat(0.55),
		MidPrice:  decimal.NewFromFloat(0.525),
		Spread:    decimal.NewFromFloat(0.05),
		SpreadBps: decimal.NewFromFloat(952.38),
		BidCount:  5,
		AskCount:  3,
		LastPrice: decimal.NewFromFloat(0.52),
	}

	if summary.TokenID != "token-1" {
		t.Errorf("TokenID = %s, want token-1", summary.TokenID)
	}
	if summary.BidCount != 5 {
		t.Errorf("BidCount = %d, want 5", summary.BidCount)
	}
}
