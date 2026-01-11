package orderbook

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestConnectionStateString(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateDisconnected, "disconnected"},
		{StateConnecting, "connecting"},
		{StateConnected, "connected"},
		{StateReconnecting, "reconnecting"},
		{ConnectionState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("ConnectionState(%d).String() = %s, want %s", tt.state, got, tt.expected)
		}
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		event    EventType
		expected string
	}{
		{EventConnected, "connected"},
		{EventDisconnected, "disconnected"},
		{EventReconnecting, "reconnecting"},
		{EventSubscribed, "subscribed"},
		{EventUnsubscribed, "unsubscribed"},
		{EventDepthUpdate, "depth_update"},
		{EventPriceUpdate, "price_update"},
		{EventTradeUpdate, "trade_update"},
		{EventOrderUpdate, "order_update"},
		{EventTradeRecord, "trade_record"},
		{EventError, "error"},
		{EventType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.event.String(); got != tt.expected {
			t.Errorf("EventType(%d).String() = %s, want %s", tt.event, got, tt.expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.WSEndpoint != "wss://ws.opinion.trade" {
		t.Errorf("WSEndpoint = %s, want wss://ws.opinion.trade", config.WSEndpoint)
	}
	if config.HeartbeatInterval != 30 {
		t.Errorf("HeartbeatInterval = %d, want 30", config.HeartbeatInterval)
	}
	if config.ReconnectDelay != 1000 {
		t.Errorf("ReconnectDelay = %d, want 1000", config.ReconnectDelay)
	}
	if config.MaxReconnectAttempts != 10 {
		t.Errorf("MaxReconnectAttempts = %d, want 10", config.MaxReconnectAttempts)
	}
	if config.BufferSize != 100 {
		t.Errorf("BufferSize = %d, want 100", config.BufferSize)
	}
}

func TestSubscriptionKey(t *testing.T) {
	tests := []struct {
		sub      *Subscription
		expected string
	}{
		{&Subscription{Channel: "market.depth.diff", MarketID: 0}, "market.depth.diff"},
		{&Subscription{Channel: "market.depth.diff", MarketID: 123}, "market.depth.diff:123"},
		{&Subscription{Channel: "market.depth.diff", MarketID: 3892}, "market.depth.diff:3892"},
		{&Subscription{Channel: "market.last.price", MarketID: 99999}, "market.last.price:99999"},
		{&Subscription{Channel: "trade.order.update", MarketID: 0}, "trade.order.update"},
		// 测试大数字确保正确转换
		{&Subscription{Channel: "market.depth.diff", MarketID: 123456789}, "market.depth.diff:123456789"},
	}

	for i, tt := range tests {
		got := tt.sub.Key()
		if got != tt.expected {
			t.Errorf("Test %d: Key() = %q, want %q", i, got, tt.expected)
		}
	}

	// 确保不同的 marketID 产生不同的 key
	sub1 := &Subscription{Channel: "market.depth.diff", MarketID: 3892}
	sub2 := &Subscription{Channel: "market.depth.diff", MarketID: 3893}
	if sub1.Key() == sub2.Key() {
		t.Errorf("Different marketIDs should produce different keys: %s == %s",
			sub1.Key(), sub2.Key())
	}
}

func TestPriceLevel(t *testing.T) {
	level := PriceLevel{
		Price: decimal.NewFromFloat(0.55),
		Size:  decimal.NewFromInt(100),
	}

	if !level.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("Price = %s, want 0.55", level.Price)
	}
	if !level.Size.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Size = %s, want 100", level.Size)
	}
}

func TestOrderSummary(t *testing.T) {
	summary := OrderSummary{
		Price: decimal.NewFromFloat(0.60),
		Size:  decimal.NewFromInt(50),
	}

	if !summary.Price.Equal(decimal.NewFromFloat(0.60)) {
		t.Errorf("Price = %s, want 0.60", summary.Price)
	}
	if !summary.Size.Equal(decimal.NewFromInt(50)) {
		t.Errorf("Size = %s, want 50", summary.Size)
	}
}

func TestBBO(t *testing.T) {
	bbo := &BBO{
		BestBid: &OrderSummary{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		BestAsk: &OrderSummary{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}

	if bbo.BestBid == nil {
		t.Error("BestBid should not be nil")
	}
	if bbo.BestAsk == nil {
		t.Error("BestAsk should not be nil")
	}
	if !bbo.BestBid.Price.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("BestBid.Price = %s, want 0.50", bbo.BestBid.Price)
	}
	if !bbo.BestAsk.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("BestAsk.Price = %s, want 0.55", bbo.BestAsk.Price)
	}
}

func TestDepth(t *testing.T) {
	depth := &Depth{
		Bids: []OrderSummary{
			{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
			{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
		},
		Asks: []OrderSummary{
			{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
			{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
		},
	}

	if len(depth.Bids) != 2 {
		t.Errorf("Bids count = %d, want 2", len(depth.Bids))
	}
	if len(depth.Asks) != 2 {
		t.Errorf("Asks count = %d, want 2", len(depth.Asks))
	}
}

func TestDepthMessage(t *testing.T) {
	msg := DepthMessage{
		Channel:  ChannelDepthDiff,
		MarketID: 123,
		TokenID:  "token-1",
		Bids: []PriceLevel{
			{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		},
		Asks: []PriceLevel{
			{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		},
		Sequence: 1,
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
}

func TestDepthDiffMessage(t *testing.T) {
	msg := DepthDiffMessage{
		Channel:  ChannelDepthDiff,
		MarketID: 123,
		TokenID:  "token-1",
		Bids: []PriceLevelDiff{
			{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		},
		Asks: []PriceLevelDiff{
			{Price: decimal.NewFromFloat(0.55), Size: decimal.Zero},
		},
		Sequence: 2,
	}

	if len(msg.Bids) != 1 {
		t.Errorf("Bids count = %d, want 1", len(msg.Bids))
	}
	if len(msg.Asks) != 1 {
		t.Errorf("Asks count = %d, want 1", len(msg.Asks))
	}
	if !msg.Asks[0].Size.IsZero() {
		t.Error("Ask size should be zero (removal)")
	}
}

func TestLastPriceMessage(t *testing.T) {
	msg := LastPriceMessage{
		Channel:   ChannelLastPrice,
		MarketID:  123,
		TokenID:   "token-1",
		Price:     decimal.NewFromFloat(0.55),
		Timestamp: 1704067200000,
	}

	if msg.Channel != ChannelLastPrice {
		t.Errorf("Channel = %s, want %s", msg.Channel, ChannelLastPrice)
	}
	if !msg.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("Price = %s, want 0.55", msg.Price)
	}
}

func TestLastTradeMessage(t *testing.T) {
	msg := LastTradeMessage{
		Channel:   ChannelLastTrade,
		MarketID:  123,
		TokenID:   "token-1",
		TradeID:   "trade-1",
		Price:     decimal.NewFromFloat(0.55),
		Size:      decimal.NewFromInt(100),
		Side:      "BUY",
		Timestamp: 1704067200000,
	}

	if msg.Channel != ChannelLastTrade {
		t.Errorf("Channel = %s, want %s", msg.Channel, ChannelLastTrade)
	}
	if msg.TradeID != "trade-1" {
		t.Errorf("TradeID = %s, want trade-1", msg.TradeID)
	}
	if msg.Side != "BUY" {
		t.Errorf("Side = %s, want BUY", msg.Side)
	}
}

func TestOrderUpdateMessage(t *testing.T) {
	msg := OrderUpdateMessage{
		Channel:    ChannelOrderUpdate,
		OrderID:    "order-1",
		MarketID:   123,
		TokenID:    "token-1",
		Status:     "FILLED",
		Side:       "BUY",
		Price:      decimal.NewFromFloat(0.55),
		Size:       decimal.NewFromInt(100),
		FilledSize: decimal.NewFromInt(100),
		Timestamp:  1704067200000,
	}

	if msg.OrderID != "order-1" {
		t.Errorf("OrderID = %s, want order-1", msg.OrderID)
	}
	if msg.Status != "FILLED" {
		t.Errorf("Status = %s, want FILLED", msg.Status)
	}
	if !msg.FilledSize.Equal(msg.Size) {
		t.Error("FilledSize should equal Size for filled order")
	}
}

func TestTradeRecordMessage(t *testing.T) {
	msg := TradeRecordMessage{
		Channel:   ChannelTradeRecord,
		TradeID:   "trade-1",
		OrderID:   "order-1",
		MarketID:  123,
		TokenID:   "token-1",
		Side:      "BUY",
		Price:     decimal.NewFromFloat(0.55),
		Size:      decimal.NewFromInt(100),
		Fee:       decimal.NewFromFloat(0.5),
		Timestamp: 1704067200000,
	}

	if msg.TradeID != "trade-1" {
		t.Errorf("TradeID = %s, want trade-1", msg.TradeID)
	}
	if !msg.Fee.Equal(decimal.NewFromFloat(0.5)) {
		t.Errorf("Fee = %s, want 0.5", msg.Fee)
	}
}

func TestEvent(t *testing.T) {
	event := &Event{
		Type:      EventDepthUpdate,
		MarketID:  123,
		TokenID:   "token-1",
		Data:      nil,
		Error:     nil,
		Timestamp: 1704067200000,
	}

	if event.Type != EventDepthUpdate {
		t.Errorf("Type = %v, want %v", event.Type, EventDepthUpdate)
	}
	if event.MarketID != 123 {
		t.Errorf("MarketID = %d, want 123", event.MarketID)
	}
}

func TestChannelConstants(t *testing.T) {
	if ChannelDepthDiff != "market.depth.diff" {
		t.Errorf("ChannelDepthDiff = %s, want market.depth.diff", ChannelDepthDiff)
	}
	if ChannelLastPrice != "market.last.price" {
		t.Errorf("ChannelLastPrice = %s, want market.last.price", ChannelLastPrice)
	}
	if ChannelLastTrade != "market.last.trade" {
		t.Errorf("ChannelLastTrade = %s, want market.last.trade", ChannelLastTrade)
	}
	if ChannelOrderUpdate != "trade.order.update" {
		t.Errorf("ChannelOrderUpdate = %s, want trade.order.update", ChannelOrderUpdate)
	}
	if ChannelTradeRecord != "trade.record.new" {
		t.Errorf("ChannelTradeRecord = %s, want trade.record.new", ChannelTradeRecord)
	}
}

func TestActionConstants(t *testing.T) {
	if ActionSubscribe != "SUBSCRIBE" {
		t.Errorf("ActionSubscribe = %s, want SUBSCRIBE", ActionSubscribe)
	}
	if ActionUnsubscribe != "UNSUBSCRIBE" {
		t.Errorf("ActionUnsubscribe = %s, want UNSUBSCRIBE", ActionUnsubscribe)
	}
	if ActionHeartbeat != "HEARTBEAT" {
		t.Errorf("ActionHeartbeat = %s, want HEARTBEAT", ActionHeartbeat)
	}
}

func TestSubscribeMessage(t *testing.T) {
	msg := SubscribeMessage{
		Action:   ActionSubscribe,
		Channel:  ChannelDepthDiff,
		MarketID: 123,
	}

	if msg.Action != ActionSubscribe {
		t.Errorf("Action = %s, want %s", msg.Action, ActionSubscribe)
	}
	if msg.Channel != ChannelDepthDiff {
		t.Errorf("Channel = %s, want %s", msg.Channel, ChannelDepthDiff)
	}
	if msg.MarketID != 123 {
		t.Errorf("MarketID = %d, want 123", msg.MarketID)
	}
}

func TestHeartbeatMessage(t *testing.T) {
	msg := HeartbeatMessage{
		Action: ActionHeartbeat,
	}

	if msg.Action != ActionHeartbeat {
		t.Errorf("Action = %s, want %s", msg.Action, ActionHeartbeat)
	}
}

func TestServerMessage(t *testing.T) {
	msg := ServerMessage{
		Channel:  ChannelDepthDiff,
		MarketID: 123,
		Event:    "subscribed",
		Message:  "Subscribed to market.depth.diff",
	}

	if msg.Event != "subscribed" {
		t.Errorf("Event = %s, want subscribed", msg.Event)
	}
}

