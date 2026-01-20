package orderbook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrNotConnected = errors.New("websocket not connected")
	ErrClosed       = errors.New("client closed")
)

// WSClient manages the WebSocket connection to Opinion
// Uses callback pattern (like Polymarket SDK) instead of channels
type WSClient struct {
	config *Config
	conn   *websocket.Conn
	state  ConnectionState

	mu            sync.RWMutex
	subscriptions map[string]*Subscription
	writeChan     chan []byte
	closeChan     chan struct{}
	closeOnce     sync.Once

	// Callback handlers (like Polymarket SDK)
	onMessage     func(data []byte)
	onStateChange func(state ConnectionState)

	// goroutine lifecycle control
	loopCtx    context.Context
	loopCancel context.CancelFunc
	loopWg     sync.WaitGroup

	// Reconnect control
	reconnectAttempts int32
	reconnecting      int32 // atomic flag to prevent multiple reconnects
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
		writeChan:     make(chan []byte, 100),
		closeChan:     make(chan struct{}),
	}
}

// SetMessageHandler sets the message callback (like Polymarket SDK)
func (c *WSClient) SetMessageHandler(handler func(data []byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = handler
}

// SetStateChangeHandler sets the state change callback (like Polymarket SDK)
func (c *WSClient) SetStateChangeHandler(handler func(state ConnectionState)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onStateChange = handler
}

// Connect establishes the WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.state == StateConnected || c.state == StateConnecting {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	// Stop old goroutines first
	c.stopLoops()

	c.setState(StateConnecting)

	if err := c.connect(ctx); err != nil {
		c.setState(StateDisconnected)
		return err
	}

	c.setState(StateConnected)
	atomic.StoreInt32(&c.reconnectAttempts, 0)
	atomic.StoreInt32(&c.reconnecting, 0)

	// Create new loop context
	c.mu.Lock()
	c.loopCtx, c.loopCancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	// Start read/write/heartbeat loops
	c.loopWg.Add(3)
	go c.readLoop()
	go c.writeLoop()
	go c.heartbeatLoop()

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

	c.closeOnce.Do(func() {
		close(c.closeChan)
	})

	c.stopLoopsLocked()
	c.closeConnectionLocked()
	c.state = StateDisconnected

	return nil
}

// Close closes the client permanently
func (c *WSClient) Close() error {
	return c.Disconnect()
}

// State returns the current connection state
func (c *WSClient) State() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *WSClient) setState(state ConnectionState) {
	c.mu.Lock()
	oldState := c.state
	c.state = state
	handler := c.onStateChange
	c.mu.Unlock()

	// Call state change handler outside lock
	if oldState != state && handler != nil {
		handler(state)
	}
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
	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	select {
	case c.writeChan <- data:
		return nil
	case <-c.closeChan:
		return ErrClosed
	case <-loopCtx.Done():
		return ErrClosed
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}

// stopLoops stops all loop goroutines
func (c *WSClient) stopLoops() {
	c.mu.Lock()
	c.stopLoopsLocked()
	c.mu.Unlock()

	// Wait for all goroutines to exit
	c.loopWg.Wait()
}

func (c *WSClient) stopLoopsLocked() {
	if c.loopCancel != nil {
		c.loopCancel()
	}
}

func (c *WSClient) readLoop() {
	defer c.loopWg.Done()
	defer c.triggerReconnect()

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	for {
		select {
		case <-c.closeChan:
			return
		case <-loopCtx.Done():
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

		// Call message handler directly (like Polymarket SDK)
		c.mu.RLock()
		handler := c.onMessage
		c.mu.RUnlock()

		if handler != nil {
			handler(message)
		}
	}
}

func (c *WSClient) writeLoop() {
	defer c.loopWg.Done()

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	for {
		select {
		case <-c.closeChan:
			return
		case <-loopCtx.Done():
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
	defer c.loopWg.Done()

	interval := time.Duration(c.config.HeartbeatInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	for {
		select {
		case <-c.closeChan:
			return
		case <-loopCtx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			state := c.state
			conn := c.conn
			c.mu.RUnlock()

			if state != StateConnected || conn == nil {
				continue
			}

			msg := HeartbeatMessage{Action: ActionHeartbeat}
			data, _ := json.Marshal(msg)
			c.write(data)
		}
	}
}

// triggerReconnect triggers reconnection (ensures only triggered once)
func (c *WSClient) triggerReconnect() {
	// Check if already closed
	select {
	case <-c.closeChan:
		return
	default:
	}

	// Use CAS to ensure only one goroutine triggers reconnect
	if !atomic.CompareAndSwapInt32(&c.reconnecting, 0, 1) {
		return
	}

	// Cancel loopCtx to notify all loop goroutines to exit
	c.mu.Lock()
	if c.loopCancel != nil {
		c.loopCancel()
	}
	c.mu.Unlock()

	c.closeConnection()
	c.setState(StateReconnecting)

	// Start reconnect in new goroutine
	go c.reconnect()
}

// closeConnection closes current connection (doesn't trigger reconnect)
func (c *WSClient) closeConnection() {
	c.mu.Lock()
	c.closeConnectionLocked()
	c.mu.Unlock()
}

func (c *WSClient) closeConnectionLocked() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *WSClient) reconnect() {
	// Wait for old goroutines to exit
	c.loopWg.Wait()

	// Drain writeChan
	c.drainWriteChan()

	maxAttempts := c.config.MaxReconnectAttempts
	delay := time.Duration(c.config.ReconnectDelay) * time.Millisecond

	for {
		select {
		case <-c.closeChan:
			return
		default:
		}

		attempts := atomic.AddInt32(&c.reconnectAttempts, 1)

		if maxAttempts > 0 && int(attempts) > maxAttempts {
			c.setState(StateDisconnected)
			atomic.StoreInt32(&c.reconnecting, 0)
			return
		}

		select {
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
		atomic.StoreInt32(&c.reconnectAttempts, 0)
		atomic.StoreInt32(&c.reconnecting, 0)

		// Create new loop context
		c.mu.Lock()
		c.loopCtx, c.loopCancel = context.WithCancel(context.Background())
		c.mu.Unlock()

		// Restart loops
		c.loopWg.Add(3)
		go c.readLoop()
		go c.writeLoop()
		go c.heartbeatLoop()

		// Resubscribe to all channels
		c.resubscribe()

		return
	}
}

func (c *WSClient) drainWriteChan() {
	for {
		select {
		case <-c.writeChan:
		default:
			return
		}
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
