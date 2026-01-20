package orderbook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestNewWSClient(t *testing.T) {
	client := NewWSClient(nil)
	if client == nil {
		t.Fatal("NewWSClient returned nil")
	}
	if client.State() != StateDisconnected {
		t.Errorf("Initial state = %v, want StateDisconnected", client.State())
	}
}

func TestNewWSClientWithConfig(t *testing.T) {
	config := &Config{
		APIKey:            "test-key",
		WSEndpoint:        "wss://test.example.com",
		HeartbeatInterval: 60,
		BufferSize:        200,
	}
	client := NewWSClient(config)

	if client.config.APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", client.config.APIKey)
	}
	if client.config.WSEndpoint != "wss://test.example.com" {
		t.Errorf("WSEndpoint = %s, want wss://test.example.com", client.config.WSEndpoint)
	}
}

func TestWSClientConnectDisconnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Keep connection alive
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
		HeartbeatInterval: 60,
		BufferSize:        10,
	}
	client := NewWSClient(config)

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	if client.State() != StateConnected {
		t.Errorf("State after connect = %v, want StateConnected", client.State())
	}

	// Disconnect
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Disconnect() error: %v", err)
	}

	if client.State() != StateDisconnected {
		t.Errorf("State after disconnect = %v, want StateDisconnected", client.State())
	}
}

func TestWSClientConnectAlreadyConnected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 60}
	client := NewWSClient(config)

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	// Connect again should be a no-op
	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Second Connect() should not error: %v", err)
	}
}

func TestWSClientStateChange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 60, BufferSize: 10}
	client := NewWSClient(config)

	// Set up state change callback - buffer multiple states
	stateChanged := make(chan ConnectionState, 10)
	client.SetStateChangeHandler(func(state ConnectionState) {
		stateChanged <- state
	})

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	// Wait for StateConnected (may receive StateConnecting first)
	timeout := time.After(2 * time.Second)
	for {
		select {
		case state := <-stateChanged:
			if state == StateConnected {
				return // Success
			}
			// Continue waiting if we got StateConnecting
		case <-timeout:
			t.Error("Timeout waiting for connected state")
			return
		}
	}
}

func TestWSClientSubscribeWithoutConnection(t *testing.T) {
	client := NewWSClient(nil)

	err := client.Subscribe(ChannelDepthDiff, 123)
	if err != ErrNotConnected {
		t.Errorf("Subscribe without connection error = %v, want ErrNotConnected", err)
	}
}

func TestWSClientSubscribe(t *testing.T) {
	receivedMessages := make(chan []byte, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
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
	client := NewWSClient(config)

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	err := client.Subscribe(ChannelDepthDiff, 123)
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	// Verify message sent
	select {
	case msg := <-receivedMessages:
		var subMsg SubscribeMessage
		if err := json.Unmarshal(msg, &subMsg); err != nil {
			t.Fatalf("Failed to unmarshal subscribe message: %v", err)
		}
		if subMsg.Action != ActionSubscribe {
			t.Errorf("Action = %s, want %s", subMsg.Action, ActionSubscribe)
		}
		if subMsg.Channel != ChannelDepthDiff {
			t.Errorf("Channel = %s, want %s", subMsg.Channel, ChannelDepthDiff)
		}
		if subMsg.MarketID != 123 {
			t.Errorf("MarketID = %d, want 123", subMsg.MarketID)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for subscribe message")
	}

	// Check subscription is tracked
	subs := client.GetSubscriptions()
	if len(subs) != 1 {
		t.Errorf("Subscriptions count = %d, want 1", len(subs))
	}
}

func TestWSClientUnsubscribe(t *testing.T) {
	receivedMessages := make(chan []byte, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
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
	client := NewWSClient(config)

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// Subscribe first
	client.Subscribe(ChannelDepthDiff, 123)
	<-receivedMessages // Consume subscribe message

	// Unsubscribe
	err := client.Unsubscribe(ChannelDepthDiff, 123)
	if err != nil {
		t.Fatalf("Unsubscribe() error: %v", err)
	}

	// Verify unsubscribe message
	select {
	case msg := <-receivedMessages:
		var subMsg SubscribeMessage
		json.Unmarshal(msg, &subMsg)
		if subMsg.Action != ActionUnsubscribe {
			t.Errorf("Action = %s, want %s", subMsg.Action, ActionUnsubscribe)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for unsubscribe message")
	}

	// Check subscription is removed
	subs := client.GetSubscriptions()
	if len(subs) != 0 {
		t.Errorf("Subscriptions count after unsubscribe = %d, want 0", len(subs))
	}
}

func TestWSClientSubscribeHelpers(t *testing.T) {
	receivedCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			receivedCount++
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300, BufferSize: 10}
	client := NewWSClient(config)

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// Test helper methods
	if err := client.SubscribeDepth(123); err != nil {
		t.Errorf("SubscribeDepth() error: %v", err)
	}
	if err := client.SubscribePrice(123); err != nil {
		t.Errorf("SubscribePrice() error: %v", err)
	}
	if err := client.SubscribeTrade(123); err != nil {
		t.Errorf("SubscribeTrade() error: %v", err)
	}
	if err := client.SubscribeOrderUpdates(); err != nil {
		t.Errorf("SubscribeOrderUpdates() error: %v", err)
	}
	if err := client.SubscribeTradeRecords(); err != nil {
		t.Errorf("SubscribeTradeRecords() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if len(client.GetSubscriptions()) != 5 {
		t.Errorf("Subscriptions count = %d, want 5", len(client.GetSubscriptions()))
	}
}

func TestWSClientHandleDepthMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send depth message
		msg := `{"channel":"market.depth.diff","marketId":123,"tokenId":"token-1","bids":[{"price":"0.5","size":"100"}],"asks":[],"sequence":1}`
		conn.WriteMessage(websocket.TextMessage, []byte(msg))

		// Keep connection alive
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
	client := NewWSClient(config)

	// Set up message handler
	receivedMsg := make(chan []byte, 1)
	client.SetMessageHandler(func(data []byte) {
		select {
		case receivedMsg <- data:
		default:
		}
	})

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	// Wait for depth message
	select {
	case data := <-receivedMsg:
		var msg struct {
			Channel  string `json:"channel"`
			MarketID int    `json:"marketId"`
			TokenID  string `json:"tokenId"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
		if msg.Channel != ChannelDepthDiff {
			t.Errorf("Channel = %s, want %s", msg.Channel, ChannelDepthDiff)
		}
		if msg.MarketID != 123 {
			t.Errorf("MarketID = %d, want 123", msg.MarketID)
		}
		if msg.TokenID != "token-1" {
			t.Errorf("TokenID = %s, want token-1", msg.TokenID)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for depth message")
	}
}

func TestWSClientHandlePriceMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := `{"channel":"market.last.price","marketId":123,"tokenId":"token-1","price":"0.55","timestamp":1704067200000}`
		conn.WriteMessage(websocket.TextMessage, []byte(msg))

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
	client := NewWSClient(config)

	// Set up message handler
	receivedMsg := make(chan []byte, 1)
	client.SetMessageHandler(func(data []byte) {
		select {
		case receivedMsg <- data:
		default:
		}
	})

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	select {
	case data := <-receivedMsg:
		var msg struct {
			Channel string `json:"channel"`
		}
		json.Unmarshal(data, &msg)
		if msg.Channel != ChannelLastPrice {
			t.Errorf("Channel = %s, want %s", msg.Channel, ChannelLastPrice)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for price message")
	}
}

func TestWSClientHandleTradeMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := `{"channel":"market.last.trade","marketId":123,"tokenId":"token-1","tradeId":"trade-1","price":"0.55","size":"100","side":"BUY","timestamp":1704067200000}`
		conn.WriteMessage(websocket.TextMessage, []byte(msg))

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
	client := NewWSClient(config)

	// Set up message handler
	receivedMsg := make(chan []byte, 1)
	client.SetMessageHandler(func(data []byte) {
		select {
		case receivedMsg <- data:
		default:
		}
	})

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	select {
	case data := <-receivedMsg:
		var msg struct {
			Channel string `json:"channel"`
		}
		json.Unmarshal(data, &msg)
		if msg.Channel != ChannelLastTrade {
			t.Errorf("Channel = %s, want %s", msg.Channel, ChannelLastTrade)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for trade message")
	}
}

func TestWSClientHandleSubscribedEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Wait a bit for the client to be ready
		time.Sleep(50 * time.Millisecond)

		msg := `{"event":"subscribed","channel":"market.depth.diff","marketId":123}`
		conn.WriteMessage(websocket.TextMessage, []byte(msg))

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
	client := NewWSClient(config)

	// Set up message handler
	receivedMsg := make(chan []byte, 1)
	client.SetMessageHandler(func(data []byte) {
		select {
		case receivedMsg <- data:
		default:
		}
	})

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	select {
	case data := <-receivedMsg:
		var msg struct {
			Event string `json:"event"`
		}
		json.Unmarshal(data, &msg)
		if msg.Event != "subscribed" {
			t.Errorf("Event = %s, want subscribed", msg.Event)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for subscribed message")
	}
}

func TestWSClientClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	config := &Config{WSEndpoint: wsURL, HeartbeatInterval: 300}
	client := NewWSClient(config)

	ctx := context.Background()
	client.Connect(ctx)

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	if client.State() != StateDisconnected {
		t.Errorf("State after Close = %v, want StateDisconnected", client.State())
	}
}

func TestWSClientLastError(t *testing.T) {
	client := NewWSClient(nil)

	if client.LastError() != nil {
		t.Error("LastError should be nil initially")
	}
}

func TestWSClientAPIKeyInURL(t *testing.T) {
	var receivedAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.URL.Query().Get("apikey")
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
		WSEndpoint: wsURL,
		APIKey:     "my-api-key",
		HeartbeatInterval: 300,
	}
	client := NewWSClient(config)

	ctx := context.Background()
	client.Connect(ctx)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	if receivedAPIKey != "my-api-key" {
		t.Errorf("Received API key = %s, want my-api-key", receivedAPIKey)
	}
}
