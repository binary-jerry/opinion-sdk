package orderbook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrNotConnected = errors.New("websocket not connected")
	ErrClosed       = errors.New("client closed")
)

// WSClient manages the WebSocket connection to Opinion
type WSClient struct {
	config *Config
	conn   *websocket.Conn
	state  ConnectionState

	mu            sync.RWMutex
	subscriptions map[string]*Subscription
	eventChan     chan *Event
	writeChan     chan []byte
	doneChan      chan struct{}
	closeChan     chan struct{}

	reconnectAttempts int
	lastError         error
}

// NewWSClient creates a new WebSocket client
func NewWSClient(config *Config) *WSClient {
	if config == nil {
		config = DefaultConfig()
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 100
	}

	return &WSClient{
		config:        config,
		state:         StateDisconnected,
		subscriptions: make(map[string]*Subscription),
		eventChan:     make(chan *Event, config.BufferSize),
		writeChan:     make(chan []byte, 100),
		doneChan:      make(chan struct{}),
		closeChan:     make(chan struct{}),
	}
}

// Connect establishes the WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.state == StateConnected || c.state == StateConnecting {
		c.mu.Unlock()
		return nil
	}
	c.state = StateConnecting
	c.mu.Unlock()

	if err := c.connect(ctx); err != nil {
		c.setState(StateDisconnected)
		return err
	}

	c.setState(StateConnected)
	c.reconnectAttempts = 0

	// Start read/write/heartbeat loops
	go c.readLoop()
	go c.writeLoop()
	go c.heartbeatLoop()

	// Send connected event
	c.sendEvent(&Event{
		Type:      EventConnected,
		Timestamp: time.Now().UnixMilli(),
	})

	return nil
}

func (c *WSClient) connect(ctx context.Context) error {
	// Build WebSocket URL with API key
	wsURL, err := url.Parse(c.config.WSEndpoint)
	if err != nil {
		return fmt.Errorf("invalid websocket endpoint: %w", err)
	}

	q := wsURL.Query()
	if c.config.APIKey != "" {
		q.Set("apikey", c.config.APIKey)
	}
	wsURL.RawQuery = q.Encode()

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	return nil
}

// Disconnect closes the WebSocket connection
func (c *WSClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateDisconnected {
		return nil
	}

	close(c.closeChan)

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.state = StateDisconnected
	return nil
}

// Close closes the client permanently
func (c *WSClient) Close() error {
	c.Disconnect()
	close(c.doneChan)
	return nil
}

// State returns the current connection state
func (c *WSClient) State() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *WSClient) setState(state ConnectionState) {
	c.mu.Lock()
	c.state = state
	c.mu.Unlock()
}

// Events returns the event channel
func (c *WSClient) Events() <-chan *Event {
	return c.eventChan
}

// Subscribe subscribes to a channel
func (c *WSClient) Subscribe(channel string, marketID int64) error {
	c.mu.RLock()
	if c.state != StateConnected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	msg := SubscribeMessage{
		Action:   ActionSubscribe,
		Channel:  channel,
		MarketID: marketID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Track subscription
	sub := &Subscription{Channel: channel, MarketID: marketID}
	c.mu.Lock()
	c.subscriptions[sub.Key()] = sub
	c.mu.Unlock()

	return c.write(data)
}

// Unsubscribe unsubscribes from a channel
func (c *WSClient) Unsubscribe(channel string, marketID int64) error {
	c.mu.RLock()
	if c.state != StateConnected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	msg := SubscribeMessage{
		Action:   ActionUnsubscribe,
		Channel:  channel,
		MarketID: marketID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Remove subscription tracking
	sub := &Subscription{Channel: channel, MarketID: marketID}
	c.mu.Lock()
	delete(c.subscriptions, sub.Key())
	c.mu.Unlock()

	return c.write(data)
}

// SubscribeDepth subscribes to orderbook depth updates
func (c *WSClient) SubscribeDepth(marketID int64) error {
	return c.Subscribe(ChannelDepthDiff, marketID)
}

// SubscribePrice subscribes to price updates
func (c *WSClient) SubscribePrice(marketID int64) error {
	return c.Subscribe(ChannelLastPrice, marketID)
}

// SubscribeTrade subscribes to trade updates
func (c *WSClient) SubscribeTrade(marketID int64) error {
	return c.Subscribe(ChannelLastTrade, marketID)
}

// SubscribeOrderUpdates subscribes to order updates (user channel)
func (c *WSClient) SubscribeOrderUpdates() error {
	return c.Subscribe(ChannelOrderUpdate, 0)
}

// SubscribeTradeRecords subscribes to trade records (user channel)
func (c *WSClient) SubscribeTradeRecords() error {
	return c.Subscribe(ChannelTradeRecord, 0)
}

func (c *WSClient) write(data []byte) error {
	select {
	case c.writeChan <- data:
		return nil
	case <-c.closeChan:
		return ErrClosed
	case <-c.doneChan:
		return ErrClosed
	}
}

func (c *WSClient) sendEvent(event *Event) {
	select {
	case c.eventChan <- event:
	default:
		// Drop event if channel is full
	}
}

func (c *WSClient) readLoop() {
	defer func() {
		c.handleDisconnect()
	}()

	for {
		select {
		case <-c.closeChan:
			return
		case <-c.doneChan:
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			c.lastError = err
			return
		}

		c.handleMessage(message)
	}
}

func (c *WSClient) writeLoop() {
	for {
		select {
		case <-c.closeChan:
			return
		case <-c.doneChan:
			return
		case data := <-c.writeChan:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.lastError = err
			}
		}
	}
}

func (c *WSClient) heartbeatLoop() {
	interval := time.Duration(c.config.HeartbeatInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closeChan:
			return
		case <-c.doneChan:
			return
		case <-ticker.C:
			c.mu.RLock()
			state := c.state
			c.mu.RUnlock()

			if state != StateConnected {
				continue
			}

			msg := HeartbeatMessage{Action: ActionHeartbeat}
			data, _ := json.Marshal(msg)
			c.write(data)
		}
	}
}

func (c *WSClient) handleMessage(data []byte) {
	// Try to parse as different message types
	var base ServerMessage
	if err := json.Unmarshal(data, &base); err != nil {
		c.sendEvent(&Event{
			Type:      EventError,
			Error:     fmt.Errorf("failed to parse message: %w", err),
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// Check for event notifications first (subscribed/unsubscribed)
	if base.Event == "subscribed" {
		c.sendEvent(&Event{
			Type:      EventSubscribed,
			MarketID:  base.MarketID,
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}
	if base.Event == "unsubscribed" {
		c.sendEvent(&Event{
			Type:      EventUnsubscribed,
			MarketID:  base.MarketID,
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	switch base.Channel {
	case ChannelDepthDiff:
		var msg DepthDiffMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			c.sendEvent(&Event{
				Type:      EventDepthUpdate,
				MarketID:  msg.MarketID,
				TokenID:   msg.TokenID,
				Data:      &msg,
				Timestamp: time.Now().UnixMilli(),
			})
		}

	case ChannelLastPrice:
		var msg LastPriceMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			c.sendEvent(&Event{
				Type:      EventPriceUpdate,
				MarketID:  msg.MarketID,
				TokenID:   msg.TokenID,
				Data:      &msg,
				Timestamp: time.Now().UnixMilli(),
			})
		}

	case ChannelLastTrade:
		var msg LastTradeMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			c.sendEvent(&Event{
				Type:      EventTradeUpdate,
				MarketID:  msg.MarketID,
				TokenID:   msg.TokenID,
				Data:      &msg,
				Timestamp: time.Now().UnixMilli(),
			})
		}

	case ChannelOrderUpdate:
		var msg OrderUpdateMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			c.sendEvent(&Event{
				Type:      EventOrderUpdate,
				MarketID:  msg.MarketID,
				TokenID:   msg.TokenID,
				Data:      &msg,
				Timestamp: time.Now().UnixMilli(),
			})
		}

	case ChannelTradeRecord:
		var msg TradeRecordMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			c.sendEvent(&Event{
				Type:      EventTradeRecord,
				MarketID:  msg.MarketID,
				TokenID:   msg.TokenID,
				Data:      &msg,
				Timestamp: time.Now().UnixMilli(),
			})
		}

	default:
		// Unknown message type, ignore
	}
}

func (c *WSClient) handleDisconnect() {
	c.sendEvent(&Event{
		Type:      EventDisconnected,
		Error:     c.lastError,
		Timestamp: time.Now().UnixMilli(),
	})

	// Check if we should reconnect
	select {
	case <-c.doneChan:
		return
	case <-c.closeChan:
		return
	default:
	}

	c.reconnect()
}

func (c *WSClient) reconnect() {
	c.setState(StateReconnecting)

	c.sendEvent(&Event{
		Type:      EventReconnecting,
		Timestamp: time.Now().UnixMilli(),
	})

	maxAttempts := c.config.MaxReconnectAttempts
	delay := time.Duration(c.config.ReconnectDelay) * time.Millisecond

	for {
		c.reconnectAttempts++

		if maxAttempts > 0 && c.reconnectAttempts > maxAttempts {
			c.setState(StateDisconnected)
			c.sendEvent(&Event{
				Type:      EventError,
				Error:     errors.New("max reconnect attempts exceeded"),
				Timestamp: time.Now().UnixMilli(),
			})
			return
		}

		select {
		case <-c.doneChan:
			return
		case <-c.closeChan:
			return
		case <-time.After(delay):
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := c.connect(ctx)
		cancel()

		if err != nil {
			c.lastError = err
			delay = delay * 2 // Exponential backoff
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			continue
		}

		// Reconnected successfully
		c.setState(StateConnected)
		c.reconnectAttempts = 0

		// Restart loops
		go c.readLoop()
		go c.writeLoop()
		go c.heartbeatLoop()

		c.sendEvent(&Event{
			Type:      EventConnected,
			Timestamp: time.Now().UnixMilli(),
		})

		// Resubscribe to all channels
		c.resubscribe()
		return
	}
}

func (c *WSClient) resubscribe() {
	c.mu.RLock()
	subs := make([]*Subscription, 0, len(c.subscriptions))
	for _, sub := range c.subscriptions {
		subs = append(subs, sub)
	}
	c.mu.RUnlock()

	for _, sub := range subs {
		c.Subscribe(sub.Channel, sub.MarketID)
	}
}

// GetSubscriptions returns a list of active subscriptions
func (c *WSClient) GetSubscriptions() []*Subscription {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subs := make([]*Subscription, 0, len(c.subscriptions))
	for _, sub := range c.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

// LastError returns the last error encountered
func (c *WSClient) LastError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastError
}
