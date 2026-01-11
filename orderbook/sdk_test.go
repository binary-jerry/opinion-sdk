package orderbook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

func TestNewSDK(t *testing.T) {
	sdk := NewSDK(nil)
	if sdk == nil {
		t.Fatal("NewSDK returned nil")
	}
	if sdk.config == nil {
		t.Error("SDK config should not be nil")
	}
	if sdk.wsClient == nil {
		t.Error("SDK wsClient should not be nil")
	}
	if sdk.manager == nil {
		t.Error("SDK manager should not be nil")
	}
}

func TestNewSDKWithConfig(t *testing.T) {
	config := &Config{
		APIKey:            "test-key",
		WSEndpoint:        "wss://custom.endpoint.com",
		HeartbeatInterval: 45,
		BufferSize:        200,
	}
	sdk := NewSDK(config)

	if sdk.config.APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", sdk.config.APIKey)
	}
	if sdk.config.WSEndpoint != "wss://custom.endpoint.com" {
		t.Errorf("WSEndpoint = %s, want wss://custom.endpoint.com", sdk.config.WSEndpoint)
	}
}

func TestSDKStartStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{
		WSEndpoint:        wsURL,
		HeartbeatInterval: 300,
		BufferSize:        10,
	}
	sdk := NewSDK(config)

	ctx := context.Background()
	err := sdk.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if !sdk.IsConnected() {
		t.Error("SDK should be connected after Start")
	}

	err = sdk.Stop()
	if err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	if sdk.IsConnected() {
		t.Error("SDK should not be connected after Stop")
	}
}

func TestSDKStartAlreadyStarted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	// Second start should be no-op
	err := sdk.Start(ctx)
	if err != nil {
		t.Errorf("Second Start() should not error: %v", err)
	}
}

func TestSDKStopNotStarted(t *testing.T) {
	sdk := NewSDK(nil)

	err := sdk.Stop()
	if err != nil {
		t.Errorf("Stop() on non-started SDK should not error: %v", err)
	}
}

func TestSDKSubscribe(t *testing.T) {
	receivedMessages := make(chan []byte, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			receivedMessages <- msg
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300, BufferSize: 10}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	// Test Subscribe
	err := sdk.Subscribe(123)
	if err != nil {
		t.Errorf("Subscribe() error: %v", err)
	}

	// Verify message received
	select {
	case <-receivedMessages:
		// OK
	case <-time.After(time.Second):
		t.Error("Timeout waiting for subscribe message")
	}
}

func TestSDKSubscribeAll(t *testing.T) {
	messageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
			messageCount++
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	err := sdk.SubscribeAll(123)
	if err != nil {
		t.Errorf("SubscribeAll() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Should have sent 3 subscribe messages (depth, price, trade)
	if messageCount < 3 {
		t.Errorf("Message count = %d, want at least 3", messageCount)
	}
}

func TestSDKUnsubscribe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	// Subscribe first
	sdk.Subscribe(123)
	time.Sleep(50 * time.Millisecond)

	// Unsubscribe
	err := sdk.Unsubscribe(123)
	if err != nil {
		t.Errorf("Unsubscribe() error: %v", err)
	}
}

func TestSDKUnsubscribeAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	err := sdk.UnsubscribeAll(123)
	if err != nil {
		t.Errorf("UnsubscribeAll() error: %v", err)
	}
}

func TestSDKEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300, BufferSize: 10}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	events := sdk.Events()
	if events == nil {
		t.Error("Events() should not return nil")
	}
}

func TestSDKGetBBO(t *testing.T) {
	sdk := NewSDK(nil)

	// Apply some data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, asks, 1)

	bbo := sdk.GetBBO("token-1")
	if bbo.BestBid == nil {
		t.Error("BestBid should not be nil")
	}
	if !bbo.BestBid.Price.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("BestBid.Price = %s, want 0.50", bbo.BestBid.Price)
	}
}

func TestSDKGetDepth(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, nil, 1)

	depth := sdk.GetDepth("token-1", 5)
	if len(depth.Bids) != 2 {
		t.Errorf("Bids count = %d, want 2", len(depth.Bids))
	}
}

func TestSDKGetLastPrice(t *testing.T) {
	sdk := NewSDK(nil)

	price := sdk.GetLastPrice("token-1")
	if !price.IsZero() {
		t.Error("LastPrice for unknown token should be zero")
	}
}

func TestSDKGetMidPrice(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(50)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, asks, 1)

	midPrice := sdk.GetMidPrice("token-1")
	expected := decimal.NewFromFloat(0.55)
	if !midPrice.Equal(expected) {
		t.Errorf("MidPrice = %s, want %s", midPrice, expected)
	}
}

func TestSDKGetSpread(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, asks, 1)

	spread := sdk.GetSpread("token-1")
	expected := decimal.NewFromFloat(0.05)
	if !spread.Equal(expected) {
		t.Errorf("Spread = %s, want %s", spread, expected)
	}
}

func TestSDKGetSpreadBps(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(50)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, asks, 1)

	spreadBps := sdk.GetSpreadBps("token-1")
	if spreadBps.IsZero() {
		t.Error("SpreadBps should not be zero")
	}
}

func TestSDKScanBidsAbove(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, nil, 1)

	result := sdk.ScanBidsAbove("token-1", decimal.NewFromFloat(0.50))
	if len(result) != 1 {
		t.Errorf("ScanBidsAbove result count = %d, want 1", len(result))
	}
}

func TestSDKScanAsksBelow(t *testing.T) {
	sdk := NewSDK(nil)

	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}
	sdk.ApplySnapshot(123, "token-1", nil, asks, 1)

	result := sdk.ScanAsksBelow("token-1", decimal.NewFromFloat(0.55))
	if len(result) != 1 {
		t.Errorf("ScanAsksBelow result count = %d, want 1", len(result))
	}
}

func TestSDKGetAllBidsAsks(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, asks, 1)

	allBids := sdk.GetAllBids("token-1")
	allAsks := sdk.GetAllAsks("token-1")

	if len(allBids) != 1 {
		t.Errorf("GetAllBids count = %d, want 1", len(allBids))
	}
	if len(allAsks) != 1 {
		t.Errorf("GetAllAsks count = %d, want 1", len(allAsks))
	}
}

func TestSDKGetOrderBook(t *testing.T) {
	sdk := NewSDK(nil)

	// Non-existent
	ob := sdk.GetOrderBook("non-existent")
	if ob != nil {
		t.Error("GetOrderBook for non-existent should return nil")
	}

	// After applying snapshot
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, nil, 1)

	ob = sdk.GetOrderBook("token-1")
	if ob == nil {
		t.Error("GetOrderBook should not be nil after snapshot")
	}
}

func TestSDKGetMarketSummary(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, asks, 1)

	summary := sdk.GetMarketSummary("token-1")
	if summary.TokenID != "token-1" {
		t.Errorf("TokenID = %s, want token-1", summary.TokenID)
	}
	if !summary.BestBid.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("BestBid = %s, want 0.50", summary.BestBid)
	}
}

func TestSDKGetAllMarketSummaries(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, nil, 1)
	sdk.ApplySnapshot(456, "token-2", bids, nil, 1)

	summaries := sdk.GetAllMarketSummaries()
	if len(summaries) != 2 {
		t.Errorf("Summaries count = %d, want 2", len(summaries))
	}
}

func TestSDKGetConnectionState(t *testing.T) {
	sdk := NewSDK(nil)

	state := sdk.GetConnectionState()
	if state != StateDisconnected {
		t.Errorf("Initial state = %v, want StateDisconnected", state)
	}
}

func TestSDKIsConnected(t *testing.T) {
	sdk := NewSDK(nil)

	if sdk.IsConnected() {
		t.Error("SDK should not be connected initially")
	}
}

func TestSDKGetSubscriptions(t *testing.T) {
	sdk := NewSDK(nil)

	subs := sdk.GetSubscriptions()
	if len(subs) != 0 {
		t.Errorf("Initial subscriptions count = %d, want 0", len(subs))
	}
}

func TestSDKGetTokenIDs(t *testing.T) {
	sdk := NewSDK(nil)

	ids := sdk.GetTokenIDs()
	if len(ids) != 0 {
		t.Errorf("Initial token IDs count = %d, want 0", len(ids))
	}

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, nil, 1)

	ids = sdk.GetTokenIDs()
	if len(ids) != 1 {
		t.Errorf("Token IDs count = %d, want 1", len(ids))
	}
}

func TestSDKClearOrderBook(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	sdk.ApplySnapshot(123, "token-1", bids, nil, 1)

	sdk.ClearOrderBook("token-1")

	ob := sdk.GetOrderBook("token-1")
	if ob != nil && !ob.IsEmpty() {
		t.Error("Orderbook should be empty after clear")
	}
}

func TestSDKHealthCheck(t *testing.T) {
	sdk := NewSDK(nil)

	if sdk.HealthCheck() {
		t.Error("HealthCheck should return false when not started")
	}
}

func TestSDKConfig(t *testing.T) {
	config := &Config{
		APIKey:     "test-key",
		WSEndpoint: "wss://test.com",
	}
	sdk := NewSDK(config)

	if sdk.Config().APIKey != "test-key" {
		t.Errorf("Config().APIKey = %s, want test-key", sdk.Config().APIKey)
	}
}

func TestSDKGetWSClient(t *testing.T) {
	sdk := NewSDK(nil)

	wsClient := sdk.GetWSClient()
	if wsClient == nil {
		t.Error("GetWSClient() should not return nil")
	}
}

func TestSDKGetManager(t *testing.T) {
	sdk := NewSDK(nil)

	manager := sdk.GetManager()
	if manager == nil {
		t.Error("GetManager() should not return nil")
	}
}

func TestSDKString(t *testing.T) {
	sdk := NewSDK(nil)

	str := sdk.String()
	if str == "" {
		t.Error("String() should not return empty")
	}
}

func TestSDKApplySnapshot(t *testing.T) {
	sdk := NewSDK(nil)

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}

	sdk.ApplySnapshot(123, "token-1", bids, asks, 5)

	ob := sdk.GetOrderBook("token-1")
	if ob == nil {
		t.Fatal("Orderbook should exist after snapshot")
	}
	if ob.BidCount() != 2 {
		t.Errorf("BidCount = %d, want 2", ob.BidCount())
	}
	if ob.AskCount() != 1 {
		t.Errorf("AskCount = %d, want 1", ob.AskCount())
	}
	// Note: sequence is always 0 now since Opinion API doesn't have sequence field
	if ob.Sequence() != 0 {
		t.Errorf("Sequence = %d, want 0", ob.Sequence())
	}
}

func TestSDKSubscribeOrderUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	err := sdk.SubscribeOrderUpdates()
	if err != nil {
		t.Errorf("SubscribeOrderUpdates() error: %v", err)
	}
}

func TestSDKSubscribeTradeRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	err := sdk.SubscribeTradeRecords()
	if err != nil {
		t.Errorf("SubscribeTradeRecords() error: %v", err)
	}
}

func TestSDKGetHealthStatus(t *testing.T) {
	sdk := NewSDK(nil)

	// Before start, health status should show not connected
	status := sdk.GetHealthStatus()
	if status.Connected {
		t.Error("Expected not connected before start")
	}
	if status.OrderBookCount != 0 {
		t.Errorf("Expected 0 orderbooks, got %d", status.OrderBookCount)
	}
}

func TestSDKHealthStatusWithOrderBooks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	// Apply a snapshot to create an orderbook
	bids := []PriceLevel{{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(100)}}
	asks := []PriceLevel{{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(100)}}
	sdk.ApplySnapshot(123, "test-token", bids, asks, time.Now().UnixMilli())

	status := sdk.GetHealthStatus()
	if !status.Connected {
		t.Error("Expected connected after start")
	}
	if status.OrderBookCount != 1 {
		t.Errorf("Expected 1 orderbook, got %d", status.OrderBookCount)
	}
	if status.HealthyCount != 1 {
		t.Errorf("Expected 1 healthy orderbook, got %d", status.HealthyCount)
	}
	if len(status.OrderBooks) != 1 {
		t.Errorf("Expected 1 orderbook in list, got %d", len(status.OrderBooks))
	}
	if status.OrderBooks[0].TokenID != "test-token" {
		t.Errorf("Expected tokenID 'test-token', got '%s'", status.OrderBooks[0].TokenID)
	}
	if !status.OrderBooks[0].Initialized {
		t.Error("Expected orderbook to be initialized")
	}
}

func TestSDKValidateOrderBook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	sdk := NewSDK(config)

	ctx := context.Background()
	sdk.Start(ctx)
	defer sdk.Stop()

	time.Sleep(100 * time.Millisecond)

	// Test without snapshot fetcher - should error
	result := sdk.ValidateOrderBook(ctx, 123, "test-token")
	if result.Error == nil {
		t.Error("Expected error when no snapshot fetcher set")
	}

	// Set up a mock snapshot fetcher
	fetcher := &mockSnapshotFetcher{
		bids: []PriceLevel{{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(100)}},
		asks: []PriceLevel{{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(100)}},
	}
	sdk.SetSnapshotFetcher(fetcher)

	// Now validation should work and initialize the orderbook
	result = sdk.ValidateOrderBook(ctx, 123, "test-token")
	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
	if !result.Corrected {
		t.Error("Expected orderbook to be corrected (initialized)")
	}

	// Verify orderbook was initialized
	if !sdk.IsOrderBookInitialized("test-token") {
		t.Error("Expected orderbook to be initialized")
	}

	// Validate again - should be valid now
	result = sdk.ValidateOrderBook(ctx, 123, "test-token")
	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
	if !result.Valid {
		t.Error("Expected orderbook to be valid")
	}
}

func TestSDKCountDifferences(t *testing.T) {
	sdk := NewSDK(nil)

	// Test identical data
	local := []OrderSummary{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(200)},
	}
	api := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(200)},
	}

	diff := sdk.countDifferences(local, api)
	if diff != 0 {
		t.Errorf("Expected 0 differences, got %d", diff)
	}

	// Test different size
	api[0].Size = decimal.NewFromInt(150)
	diff = sdk.countDifferences(local, api)
	if diff != 1 {
		t.Errorf("Expected 1 difference, got %d", diff)
	}

	// Test missing level in local
	api = append(api, PriceLevel{Price: decimal.NewFromFloat(0.45), Size: decimal.NewFromInt(50)})
	diff = sdk.countDifferences(local, api)
	if diff != 2 {
		t.Errorf("Expected 2 differences, got %d", diff)
	}
}

// mockSnapshotFetcher for testing
type mockSnapshotFetcher struct {
	bids []PriceLevel
	asks []PriceLevel
}

func (m *mockSnapshotFetcher) GetOrderbookSnapshot(ctx context.Context, tokenID string) ([]PriceLevel, []PriceLevel, error) {
	return m.bids, m.asks, nil
}
