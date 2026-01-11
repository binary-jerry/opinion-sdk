package orderbook

import (
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

// OrderBook maintains the order book state for a single token
type OrderBook struct {
	mu          sync.RWMutex
	tokenID     string
	marketID    int64
	bids        map[string]*OrderSummary // price string -> order
	asks        map[string]*OrderSummary // price string -> order
	sequence    int64
	initialized bool  // 是否已初始化（收到过快照）
	timestamp   int64 // 上次更新时间戳（毫秒）
}

// NewOrderBook creates a new order book for a token
func NewOrderBook(marketID int64, tokenID string) *OrderBook {
	return &OrderBook{
		tokenID:  tokenID,
		marketID: marketID,
		bids:     make(map[string]*OrderSummary),
		asks:     make(map[string]*OrderSummary),
	}
}

// TokenID returns the token ID
func (ob *OrderBook) TokenID() string {
	return ob.tokenID
}

// MarketID returns the market ID
func (ob *OrderBook) MarketID() int64 {
	return ob.marketID
}

// Sequence returns the current sequence number
func (ob *OrderBook) Sequence() int64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.sequence
}

// IsInitialized returns whether the orderbook has received a snapshot
func (ob *OrderBook) IsInitialized() bool {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.initialized
}

// Timestamp returns the last update timestamp
func (ob *OrderBook) Timestamp() int64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.timestamp
}

// ApplySnapshot applies a full orderbook snapshot
func (ob *OrderBook) ApplySnapshot(bids, asks []PriceLevel, sequence int64, timestamp int64) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Clear existing data
	ob.bids = make(map[string]*OrderSummary)
	ob.asks = make(map[string]*OrderSummary)

	// Apply bids
	for _, bid := range bids {
		if bid.Size.IsPositive() {
			ob.bids[bid.Price.String()] = &OrderSummary{
				Price: bid.Price,
				Size:  bid.Size,
			}
		}
	}

	// Apply asks
	for _, ask := range asks {
		if ask.Size.IsPositive() {
			ob.asks[ask.Price.String()] = &OrderSummary{
				Price: ask.Price,
				Size:  ask.Size,
			}
		}
	}

	ob.sequence = sequence
	ob.timestamp = timestamp
	ob.initialized = true
}

// ApplyDiff applies incremental updates to the orderbook
// Returns true if applied, false if skipped (not initialized or old sequence)
func (ob *OrderBook) ApplyDiff(bids, asks []PriceLevelDiff, sequence int64, timestamp int64) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Skip if not initialized (no snapshot received yet)
	if !ob.initialized {
		return false
	}

	// Only apply if sequence is newer (skip check if sequence is 0, meaning no sequence support)
	if sequence > 0 && sequence <= ob.sequence {
		return false
	}

	// Apply bid updates
	for _, bid := range bids {
		key := bid.Price.String()
		if bid.Size.IsZero() {
			delete(ob.bids, key)
		} else {
			ob.bids[key] = &OrderSummary{
				Price: bid.Price,
				Size:  bid.Size,
			}
		}
	}

	// Apply ask updates
	for _, ask := range asks {
		key := ask.Price.String()
		if ask.Size.IsZero() {
			delete(ob.asks, key)
		} else {
			ob.asks[key] = &OrderSummary{
				Price: ask.Price,
				Size:  ask.Size,
			}
		}
	}

	ob.sequence = sequence
	ob.timestamp = timestamp
	return true
}

// GetBBO returns the best bid and offer
func (ob *OrderBook) GetBBO() *BBO {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bbo := &BBO{}

	// Find best bid (highest price)
	var bestBid *OrderSummary
	for _, bid := range ob.bids {
		if bestBid == nil || bid.Price.GreaterThan(bestBid.Price) {
			bestBid = &OrderSummary{Price: bid.Price, Size: bid.Size}
		}
	}
	bbo.BestBid = bestBid

	// Find best ask (lowest price)
	var bestAsk *OrderSummary
	for _, ask := range ob.asks {
		if bestAsk == nil || ask.Price.LessThan(bestAsk.Price) {
			bestAsk = &OrderSummary{Price: ask.Price, Size: ask.Size}
		}
	}
	bbo.BestAsk = bestAsk

	return bbo
}

// GetDepth returns orderbook depth up to the specified number of levels
func (ob *OrderBook) GetDepth(levels int) *Depth {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	depth := &Depth{
		Bids: make([]OrderSummary, 0, levels),
		Asks: make([]OrderSummary, 0, levels),
	}

	// Collect and sort bids (descending by price)
	bidList := make([]OrderSummary, 0, len(ob.bids))
	for _, bid := range ob.bids {
		bidList = append(bidList, *bid)
	}
	sort.Slice(bidList, func(i, j int) bool {
		return bidList[i].Price.GreaterThan(bidList[j].Price)
	})
	if len(bidList) > levels {
		bidList = bidList[:levels]
	}
	depth.Bids = bidList

	// Collect and sort asks (ascending by price)
	askList := make([]OrderSummary, 0, len(ob.asks))
	for _, ask := range ob.asks {
		askList = append(askList, *ask)
	}
	sort.Slice(askList, func(i, j int) bool {
		return askList[i].Price.LessThan(askList[j].Price)
	})
	if len(askList) > levels {
		askList = askList[:levels]
	}
	depth.Asks = askList

	return depth
}

// GetAllBids returns all bids sorted by price (descending)
func (ob *OrderBook) GetAllBids() []OrderSummary {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bids := make([]OrderSummary, 0, len(ob.bids))
	for _, bid := range ob.bids {
		bids = append(bids, *bid)
	}
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Price.GreaterThan(bids[j].Price)
	})
	return bids
}

// GetAllAsks returns all asks sorted by price (ascending)
func (ob *OrderBook) GetAllAsks() []OrderSummary {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	asks := make([]OrderSummary, 0, len(ob.asks))
	for _, ask := range ob.asks {
		asks = append(asks, *ask)
	}
	sort.Slice(asks, func(i, j int) bool {
		return asks[i].Price.LessThan(asks[j].Price)
	})
	return asks
}

// ScanBidsAbove returns all bids at or above the given price
func (ob *OrderBook) ScanBidsAbove(price decimal.Decimal) []OrderSummary {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	result := make([]OrderSummary, 0)
	for _, bid := range ob.bids {
		if bid.Price.GreaterThanOrEqual(price) {
			result = append(result, *bid)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Price.GreaterThan(result[j].Price)
	})
	return result
}

// ScanAsksBelow returns all asks at or below the given price
func (ob *OrderBook) ScanAsksBelow(price decimal.Decimal) []OrderSummary {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	result := make([]OrderSummary, 0)
	for _, ask := range ob.asks {
		if ask.Price.LessThanOrEqual(price) {
			result = append(result, *ask)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Price.LessThan(result[j].Price)
	})
	return result
}

// GetMidPrice returns the mid price between best bid and best ask
func (ob *OrderBook) GetMidPrice() decimal.Decimal {
	bbo := ob.GetBBO()
	if bbo.BestBid == nil || bbo.BestAsk == nil {
		return decimal.Zero
	}
	return bbo.BestBid.Price.Add(bbo.BestAsk.Price).Div(decimal.NewFromInt(2))
}

// GetSpread returns the spread between best bid and best ask
func (ob *OrderBook) GetSpread() decimal.Decimal {
	bbo := ob.GetBBO()
	if bbo.BestBid == nil || bbo.BestAsk == nil {
		return decimal.Zero
	}
	return bbo.BestAsk.Price.Sub(bbo.BestBid.Price)
}

// GetSpreadBps returns the spread in basis points
func (ob *OrderBook) GetSpreadBps() decimal.Decimal {
	midPrice := ob.GetMidPrice()
	if midPrice.IsZero() {
		return decimal.Zero
	}
	spread := ob.GetSpread()
	return spread.Div(midPrice).Mul(decimal.NewFromInt(10000))
}

// IsEmpty returns true if the orderbook has no orders
func (ob *OrderBook) IsEmpty() bool {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return len(ob.bids) == 0 && len(ob.asks) == 0
}

// BidCount returns the number of bid levels
func (ob *OrderBook) BidCount() int {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return len(ob.bids)
}

// AskCount returns the number of ask levels
func (ob *OrderBook) AskCount() int {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return len(ob.asks)
}

// TotalBidSize returns the total size of all bids
func (ob *OrderBook) TotalBidSize() decimal.Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	total := decimal.Zero
	for _, bid := range ob.bids {
		total = total.Add(bid.Size)
	}
	return total
}

// TotalAskSize returns the total size of all asks
func (ob *OrderBook) TotalAskSize() decimal.Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	total := decimal.Zero
	for _, ask := range ob.asks {
		total = total.Add(ask.Size)
	}
	return total
}

// Clear clears all orders from the orderbook and resets initialized state
func (ob *OrderBook) Clear() {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.bids = make(map[string]*OrderSummary)
	ob.asks = make(map[string]*OrderSummary)
	ob.sequence = 0
	ob.timestamp = 0
	ob.initialized = false
}

// Clone creates a deep copy of the orderbook
func (ob *OrderBook) Clone() *OrderBook {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	clone := NewOrderBook(ob.marketID, ob.tokenID)
	clone.sequence = ob.sequence
	clone.timestamp = ob.timestamp
	clone.initialized = ob.initialized

	for k, v := range ob.bids {
		clone.bids[k] = &OrderSummary{Price: v.Price, Size: v.Size}
	}
	for k, v := range ob.asks {
		clone.asks[k] = &OrderSummary{Price: v.Price, Size: v.Size}
	}

	return clone
}
