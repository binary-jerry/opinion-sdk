package orderbook

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// ConnectionState represents the WebSocket connection state
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

// EventType represents different types of orderbook events
type EventType int

const (
	EventConnected EventType = iota
	EventDisconnected
	EventReconnecting
	EventReconnected // New: fired after successful reconnection
	EventSubscribed
	EventUnsubscribed
	EventDepthUpdate
	EventPriceUpdate
	EventTradeUpdate
	EventOrderUpdate
	EventTradeRecord
	EventError
)

func (e EventType) String() string {
	switch e {
	case EventConnected:
		return "connected"
	case EventDisconnected:
		return "disconnected"
	case EventReconnecting:
		return "reconnecting"
	case EventReconnected:
		return "reconnected"
	case EventSubscribed:
		return "subscribed"
	case EventUnsubscribed:
		return "unsubscribed"
	case EventDepthUpdate:
		return "depth_update"
	case EventPriceUpdate:
		return "price_update"
	case EventTradeUpdate:
		return "trade_update"
	case EventOrderUpdate:
		return "order_update"
	case EventTradeRecord:
		return "trade_record"
	case EventError:
		return "error"
	default:
		return "unknown"
	}
}

// Channel names for Opinion WebSocket
const (
	ChannelDepthDiff  = "market.depth.diff"
	ChannelLastPrice  = "market.last.price"
	ChannelLastTrade  = "market.last.trade"
	ChannelOrderUpdate = "trade.order.update"
	ChannelTradeRecord = "trade.record.new"
)

// Action types for WebSocket messages
const (
	ActionSubscribe   = "SUBSCRIBE"
	ActionUnsubscribe = "UNSUBSCRIBE"
	ActionHeartbeat   = "HEARTBEAT"
)

// SubscribeMessage is the message to subscribe to a channel
type SubscribeMessage struct {
	Action   string `json:"action"`
	Channel  string `json:"channel"`
	MarketID int64  `json:"marketId,omitempty"`
}

// HeartbeatMessage is the heartbeat message
type HeartbeatMessage struct {
	Action string `json:"action"`
}

// ServerMessage is the base structure for server messages
type ServerMessage struct {
	Channel  string `json:"channel,omitempty"`
	MarketID int64  `json:"marketId,omitempty"`
	Event    string `json:"event,omitempty"`
	Message  string `json:"message,omitempty"`
}

// PriceLevel represents a single price level in the orderbook
type PriceLevel struct {
	Price  decimal.Decimal `json:"price"`
	Size   decimal.Decimal `json:"size"`
}

// DepthMessage represents a depth update message
type DepthMessage struct {
	Channel  string       `json:"channel"`
	MarketID int64        `json:"marketId"`
	TokenID  string       `json:"tokenId"`
	Bids     []PriceLevel `json:"bids"`
	Asks     []PriceLevel `json:"asks"`
	Sequence int64        `json:"sequence,omitempty"`
}

// DepthDiffMessage represents incremental depth updates
type DepthDiffMessage struct {
	Channel  string           `json:"channel"`
	MarketID int64            `json:"marketId"`
	TokenID  string           `json:"tokenId"`
	Bids     []PriceLevelDiff `json:"bids"`
	Asks     []PriceLevelDiff `json:"asks"`
	Sequence int64            `json:"sequence"`
}

// PriceLevelDiff represents a price level change (size=0 means remove)
type PriceLevelDiff struct {
	Price decimal.Decimal `json:"price"`
	Size  decimal.Decimal `json:"size"`
}

// SingleDepthDiffMessage represents a single price level update (actual WebSocket format)
// This is the real format from Opinion WebSocket, not bids/asks arrays
type SingleDepthDiffMessage struct {
	MsgType     string `json:"msgType"`     // "market.depth.diff"
	MarketID    int64  `json:"marketId"`
	TokenID     string `json:"tokenId"`
	OutcomeSide int    `json:"outcomeSide"` // 1=YES, 0=NO
	Side        string `json:"side"`        // "bids" or "asks"
	Price       string `json:"price"`
	Size        string `json:"size"`
}

// OutcomeSide constants
const (
	OutcomeSideNO  = 0
	OutcomeSideYES = 1
)

// MarketTokenPair represents the YES/NO token pair for a market
type MarketTokenPair struct {
	MarketID   int64
	YesTokenID string
	NoTokenID  string
}

// LastPriceMessage represents a price update message
type LastPriceMessage struct {
	Channel   string          `json:"channel"`
	MarketID  int64           `json:"marketId"`
	TokenID   string          `json:"tokenId"`
	Price     decimal.Decimal `json:"price"`
	Timestamp int64           `json:"timestamp"`
}

// LastTradeMessage represents a trade update message
type LastTradeMessage struct {
	Channel   string          `json:"channel"`
	MarketID  int64           `json:"marketId"`
	TokenID   string          `json:"tokenId"`
	TradeID   string          `json:"tradeId"`
	Price     decimal.Decimal `json:"price"`
	Size      decimal.Decimal `json:"size"`
	Side      string          `json:"side"`
	Timestamp int64           `json:"timestamp"`
}

// OrderUpdateMessage represents an order update message (user channel)
type OrderUpdateMessage struct {
	Channel     string          `json:"channel"`
	OrderID     string          `json:"orderId"`
	MarketID    int64           `json:"marketId"`
	TokenID     string          `json:"tokenId"`
	Status      string          `json:"status"`
	Side        string          `json:"side"`
	Price       decimal.Decimal `json:"price"`
	Size        decimal.Decimal `json:"size"`
	FilledSize  decimal.Decimal `json:"filledSize"`
	Timestamp   int64           `json:"timestamp"`
}

// TradeRecordMessage represents a new trade record (user channel)
type TradeRecordMessage struct {
	Channel   string          `json:"channel"`
	TradeID   string          `json:"tradeId"`
	OrderID   string          `json:"orderId"`
	MarketID  int64           `json:"marketId"`
	TokenID   string          `json:"tokenId"`
	Side      string          `json:"side"`
	Price     decimal.Decimal `json:"price"`
	Size      decimal.Decimal `json:"size"`
	Fee       decimal.Decimal `json:"fee"`
	Timestamp int64           `json:"timestamp"`
}

// Event represents an orderbook event
type Event struct {
	Type      EventType
	MarketID  int64
	TokenID   string
	Data      interface{}
	Error     error
	Timestamp int64
}

// OrderSummary represents a single order in the orderbook
type OrderSummary struct {
	Price decimal.Decimal
	Size  decimal.Decimal
}

// BBO represents the best bid and offer
type BBO struct {
	BestBid *OrderSummary
	BestAsk *OrderSummary
}

// Depth represents orderbook depth at multiple levels
type Depth struct {
	Bids []OrderSummary
	Asks []OrderSummary
}

// Config contains configuration for the orderbook SDK
type Config struct {
	// APIKey for authentication
	APIKey string

	// WebSocket endpoint (default: wss://ws.opinion.trade)
	WSEndpoint string

	// HeartbeatInterval in seconds (default: 30)
	HeartbeatInterval int

	// ReconnectDelay in milliseconds (default: 1000)
	ReconnectDelay int

	// MaxReconnectAttempts (default: 10, 0 for unlimited)
	MaxReconnectAttempts int

	// BufferSize for event channel (default: 100)
	BufferSize int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		WSEndpoint:           "wss://ws.opinion.trade",
		HeartbeatInterval:    30,
		ReconnectDelay:       1000,
		MaxReconnectAttempts: 10,
		BufferSize:           100,
	}
}

// Subscription represents an active subscription
type Subscription struct {
	Channel  string
	MarketID int64
}

// SubscriptionKey generates a unique key for a subscription
func (s *Subscription) Key() string {
	if s.MarketID > 0 {
		return fmt.Sprintf("%s:%d", s.Channel, s.MarketID)
	}
	return s.Channel
}

// OrderBookHealth represents the health status of an orderbook
type OrderBookHealth struct {
	TokenID       string
	Initialized   bool
	LastUpdate    int64 // Unix milliseconds
	BidCount      int
	AskCount      int
	IsStale       bool  // True if no update for too long
	StaleSeconds  int64 // Seconds since last update
}

// HealthStatus represents the overall health of the SDK
type HealthStatus struct {
	Connected        bool
	OrderBookCount   int
	HealthyCount     int
	UninitializedCount int
	StaleCount       int
	OrderBooks       []OrderBookHealth
}

// ValidationResult represents the result of orderbook validation
type ValidationResult struct {
	TokenID       string
	Valid         bool
	BidDifference int // Number of bid levels that differ
	AskDifference int // Number of ask levels that differ
	Corrected     bool
	Error         error
}
