package orderbook

import (
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

func TestManagerStop(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Stop should be safe to call
	manager.Stop()
}

func TestManagerUpdates(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	updates := manager.Updates()
	if updates == nil {
		t.Error("Updates() should not return nil")
	}
}

func TestManagerHandleDepthUpdate(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// First initialize the orderbook (required for depth updates to be applied)
	manager.ApplySnapshot(123, "token-1", nil, nil, time.Now().UnixMilli())

	// Simulate depth update via handleMessage (msgType format)
	jsonMsg := `{"msgType":"market.depth.diff","marketId":123,"tokenId":"token-1","side":"bids","price":"0.50","size":"100","outcomeSide":1}`
	manager.handleMessage([]byte(jsonMsg))

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

	// Simulate price update via handleMessage (msgType format)
	jsonMsg := `{"msgType":"market.last.price","marketId":123,"tokenId":"token-1","price":"0.55","timestamp":1704067200000}`
	manager.handleMessage([]byte(jsonMsg))

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

// ===== New tests for mirror sync and token pair features =====

const (
	TestMarketID   = int64(3892)
	TestYesTokenID = "80625377815318636821010811703646073832053347721900305398950501252099028045180"
	TestNoTokenID  = "52086497305048575202366950269787213150477134543220443931147225499380707602347"
)

func TestRegisterMarketTokenPair(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	pair := &MarketTokenPair{
		MarketID:   TestMarketID,
		YesTokenID: TestYesTokenID,
		NoTokenID:  TestNoTokenID,
	}

	manager.RegisterMarketTokenPair(pair)

	// Verify pair is registered
	gotPair := manager.GetMarketTokenPair(TestMarketID)
	if gotPair == nil {
		t.Fatal("expected pair to be registered")
	}
	if gotPair.YesTokenID != TestYesTokenID {
		t.Errorf("expected YesTokenID %s, got %s", TestYesTokenID, gotPair.YesTokenID)
	}
	if gotPair.NoTokenID != TestNoTokenID {
		t.Errorf("expected NoTokenID %s, got %s", TestNoTokenID, gotPair.NoTokenID)
	}
}

func TestGetMirrorTokenID(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	pair := &MarketTokenPair{
		MarketID:   TestMarketID,
		YesTokenID: TestYesTokenID,
		NoTokenID:  TestNoTokenID,
	}
	manager.RegisterMarketTokenPair(pair)

	// YES token should return NO token as mirror
	mirrorID := manager.GetMirrorTokenID(TestYesTokenID)
	if mirrorID != TestNoTokenID {
		t.Errorf("expected mirror of YES to be NO token, got %s", mirrorID)
	}

	// NO token should return YES token as mirror
	mirrorID = manager.GetMirrorTokenID(TestNoTokenID)
	if mirrorID != TestYesTokenID {
		t.Errorf("expected mirror of NO to be YES token, got %s", mirrorID)
	}

	// Unknown token should return empty
	mirrorID = manager.GetMirrorTokenID("unknown-token")
	if mirrorID != "" {
		t.Errorf("expected empty mirror for unknown token, got %s", mirrorID)
	}
}

func TestMirrorPriceCalculation(t *testing.T) {
	tests := []struct {
		yesPrice        string
		expectedNoPrice string
	}{
		{"0.50", "0.50"},
		{"0.75", "0.25"},
		{"0.10", "0.90"},
		{"0.99", "0.01"},
		{"0.01", "0.99"},
		{"1.00", "0.00"},
		{"0.00", "1.00"},
	}

	for _, tc := range tests {
		yesPrice, _ := decimal.NewFromString(tc.yesPrice)
		expectedNoPrice, _ := decimal.NewFromString(tc.expectedNoPrice)

		// Mirror price = 1 - original price
		mirrorPrice := decimal.NewFromInt(1).Sub(yesPrice)

		if !mirrorPrice.Equal(expectedNoPrice) {
			t.Errorf("mirror of %s = %s, expected %s", tc.yesPrice, mirrorPrice.String(), tc.expectedNoPrice)
		}
	}
}

func TestGetAllMarketTokenPairs(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Register multiple pairs
	pair1 := &MarketTokenPair{
		MarketID:   3892,
		YesTokenID: "yes-token-1",
		NoTokenID:  "no-token-1",
	}
	pair2 := &MarketTokenPair{
		MarketID:   3893,
		YesTokenID: "yes-token-2",
		NoTokenID:  "no-token-2",
	}

	manager.RegisterMarketTokenPair(pair1)
	manager.RegisterMarketTokenPair(pair2)

	// Get all pairs
	pairs := manager.GetAllMarketTokenPairs()

	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// Verify pairs exist (order may vary)
	found3892 := false
	found3893 := false
	for _, p := range pairs {
		if p.MarketID == 3892 {
			found3892 = true
		}
		if p.MarketID == 3893 {
			found3893 = true
		}
	}

	if !found3892 || !found3893 {
		t.Error("not all registered pairs were returned")
	}
}

func TestMarkUninitialized(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Create and initialize orderbook
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	manager.ApplySnapshot(TestMarketID, TestYesTokenID, bids, nil, time.Now().UnixMilli())

	// Verify initialized
	ob := manager.GetOrderBook(TestYesTokenID)
	if !ob.IsInitialized() {
		t.Error("orderbook should be initialized")
	}

	// Mark as uninitialized (for reconnect scenario)
	manager.MarkUninitialized(TestYesTokenID)

	// Verify uninitialized (Clear makes it uninitialized)
	if ob.IsInitialized() {
		t.Error("orderbook should be uninitialized after MarkUninitialized")
	}
}

func TestSingleDepthDiffParsing(t *testing.T) {
	// Test parsing of actual WebSocket message format
	msg := &SingleDepthDiffMessage{
		MsgType:     "market.depth.diff",
		MarketID:    3892,
		TokenID:     TestYesTokenID,
		OutcomeSide: 1,
		Side:        "bids",
		Price:       "0.485",
		Size:        "1000.5",
	}

	// Verify field parsing
	if msg.MsgType != ChannelDepthDiff {
		t.Errorf("expected msgType %s, got %s", ChannelDepthDiff, msg.MsgType)
	}
	if msg.OutcomeSide != OutcomeSideYES {
		t.Errorf("expected outcomeSide %d, got %d", OutcomeSideYES, msg.OutcomeSide)
	}

	// Parse price and size
	price, err := decimal.NewFromString(msg.Price)
	if err != nil {
		t.Fatalf("failed to parse price: %v", err)
	}
	if !price.Equal(decimal.NewFromFloat(0.485)) {
		t.Errorf("expected price 0.485, got %s", price.String())
	}

	size, err := decimal.NewFromString(msg.Size)
	if err != nil {
		t.Fatalf("failed to parse size: %v", err)
	}
	if !size.Equal(decimal.NewFromFloat(1000.5)) {
		t.Errorf("expected size 1000.5, got %s", size.String())
	}
}

func TestDirectOrderBookApplyDiff(t *testing.T) {
	// Test that ApplyDiff works correctly on OrderBook directly
	ob := NewOrderBook(123, "token")
	ob.ApplySnapshot(nil, nil, 0, time.Now().UnixMilli())

	if !ob.IsInitialized() {
		t.Fatal("orderbook should be initialized after ApplySnapshot")
	}

	// Apply a single ask diff
	diff := PriceLevelDiff{
		Price: decimal.NewFromFloat(0.60),
		Size:  decimal.NewFromInt(500),
	}
	ob.ApplyDiff(nil, []PriceLevelDiff{diff}, 0, time.Now().UnixMilli())

	asks := ob.GetAllAsks()
	if len(asks) != 1 {
		t.Fatalf("expected 1 ask after ApplyDiff, got %d", len(asks))
	}
}

func TestHandleSingleDepthDiffWithMirrorSync(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Register token pair
	pair := &MarketTokenPair{
		MarketID:   TestMarketID,
		YesTokenID: TestYesTokenID,
		NoTokenID:  TestNoTokenID,
	}
	manager.RegisterMarketTokenPair(pair)

	// Initialize both orderbooks with empty data
	manager.ApplySnapshot(TestMarketID, TestYesTokenID, nil, nil, time.Now().UnixMilli())
	manager.ApplySnapshot(TestMarketID, TestNoTokenID, nil, nil, time.Now().UnixMilli())

	// Verify orderbooks are initialized
	yesOb := manager.GetOrderBook(TestYesTokenID)
	if yesOb == nil {
		t.Fatal("YES orderbook should exist")
	}
	if !yesOb.IsInitialized() {
		t.Fatal("YES orderbook should be initialized")
	}

	// Simulate YES token ask update at 0.60
	// Expected mirror: NO bids at 0.40
	jsonMsg := `{"msgType":"market.depth.diff","marketId":3892,"tokenId":"` + TestYesTokenID + `","outcomeSide":1,"side":"asks","price":"0.60","size":"500"}`
	manager.handleMessage([]byte(jsonMsg))

	// Verify YES orderbook has ask at 0.60
	yesAsks := yesOb.GetAllAsks()
	if len(yesAsks) != 1 {
		t.Fatalf("expected 1 YES ask, got %d", len(yesAsks))
	}
	if !yesAsks[0].Price.Equal(decimal.NewFromFloat(0.60)) {
		t.Errorf("expected YES ask price 0.60, got %s", yesAsks[0].Price.String())
	}

	// Verify NO orderbook has mirrored bid at 0.40
	noOb := manager.GetOrderBook(TestNoTokenID)
	noBids := noOb.GetAllBids()
	if len(noBids) != 1 {
		t.Fatalf("expected 1 NO bid (mirrored), got %d", len(noBids))
	}
	if !noBids[0].Price.Equal(decimal.NewFromFloat(0.40)) {
		t.Errorf("expected NO bid price 0.40 (mirrored), got %s", noBids[0].Price.String())
	}
	if !noBids[0].Size.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected NO bid size 500, got %s", noBids[0].Size.String())
	}
}

func TestHandleSingleDepthDiffBidToAskMirror(t *testing.T) {
	wsClient := NewWSClient(nil)
	manager := NewManager(wsClient)

	// Register token pair
	pair := &MarketTokenPair{
		MarketID:   TestMarketID,
		YesTokenID: TestYesTokenID,
		NoTokenID:  TestNoTokenID,
	}
	manager.RegisterMarketTokenPair(pair)

	// Initialize both orderbooks
	manager.ApplySnapshot(TestMarketID, TestYesTokenID, nil, nil, time.Now().UnixMilli())
	manager.ApplySnapshot(TestMarketID, TestNoTokenID, nil, nil, time.Now().UnixMilli())

	// Simulate YES token bid update at 0.45
	// Expected mirror: NO asks at 0.55
	jsonMsg := `{"msgType":"market.depth.diff","marketId":3892,"tokenId":"` + TestYesTokenID + `","outcomeSide":1,"side":"bids","price":"0.45","size":"300"}`
	manager.handleMessage([]byte(jsonMsg))

	// Verify YES orderbook has bid at 0.45
	yesOb := manager.GetOrderBook(TestYesTokenID)
	yesBids := yesOb.GetAllBids()
	if len(yesBids) != 1 {
		t.Fatalf("expected 1 YES bid, got %d", len(yesBids))
	}
	if !yesBids[0].Price.Equal(decimal.NewFromFloat(0.45)) {
		t.Errorf("expected YES bid price 0.45, got %s", yesBids[0].Price.String())
	}

	// Verify NO orderbook has mirrored ask at 0.55
	noOb := manager.GetOrderBook(TestNoTokenID)
	noAsks := noOb.GetAllAsks()
	if len(noAsks) != 1 {
		t.Fatalf("expected 1 NO ask (mirrored), got %d", len(noAsks))
	}
	if !noAsks[0].Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("expected NO ask price 0.55 (mirrored), got %s", noAsks[0].Price.String())
	}
}
