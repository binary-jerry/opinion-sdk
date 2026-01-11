package orderbook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func nowMs() int64 {
	return time.Now().UnixMilli()
}

func TestNewOrderBook(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	if ob.MarketID() != 123 {
		t.Errorf("MarketID() = %d, want 123", ob.MarketID())
	}
	if ob.TokenID() != "token-1" {
		t.Errorf("TokenID() = %s, want token-1", ob.TokenID())
	}
	if ob.Sequence() != 0 {
		t.Errorf("Sequence() = %d, want 0", ob.Sequence())
	}
	if !ob.IsEmpty() {
		t.Error("New orderbook should be empty")
	}
}

func TestOrderBookApplySnapshot(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}

	ob.ApplySnapshot(bids, asks, 1, nowMs())

	if ob.Sequence() != 1 {
		t.Errorf("Sequence() = %d, want 1", ob.Sequence())
	}
	if ob.BidCount() != 2 {
		t.Errorf("BidCount() = %d, want 2", ob.BidCount())
	}
	if ob.AskCount() != 2 {
		t.Errorf("AskCount() = %d, want 2", ob.AskCount())
	}
	if ob.IsEmpty() {
		t.Error("Orderbook should not be empty after snapshot")
	}
}

func TestOrderBookApplySnapshotFiltersZeroSize(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.Zero}, // Should be ignored
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.Zero}, // Should be ignored
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}

	ob.ApplySnapshot(bids, asks, 1, nowMs())

	if ob.BidCount() != 1 {
		t.Errorf("BidCount() = %d, want 1 (zero size filtered)", ob.BidCount())
	}
	if ob.AskCount() != 1 {
		t.Errorf("AskCount() = %d, want 1 (zero size filtered)", ob.AskCount())
	}
}

func TestOrderBookApplyDiff(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	// Initial snapshot
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	ob.ApplySnapshot(bids, asks, 1, nowMs())

	// Apply diff - add new level, update existing
	bidDiffs := []PriceLevelDiff{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(150)}, // Update
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)}, // Add
	}
	askDiffs := []PriceLevelDiff{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.Zero}, // Remove
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)}, // Add
	}
	ob.ApplyDiff(bidDiffs, askDiffs, 2, nowMs())

	if ob.Sequence() != 2 {
		t.Errorf("Sequence() = %d, want 2", ob.Sequence())
	}
	if ob.BidCount() != 2 {
		t.Errorf("BidCount() = %d, want 2", ob.BidCount())
	}
	if ob.AskCount() != 1 {
		t.Errorf("AskCount() = %d, want 1 (one removed)", ob.AskCount())
	}
}

func TestOrderBookApplyDiffIgnoresOldSequence(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	ob.ApplySnapshot(bids, nil, 5, nowMs())

	// Try to apply diff with older sequence
	bidDiffs := []PriceLevelDiff{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(999)},
	}
	ob.ApplyDiff(bidDiffs, nil, 3, nowMs()) // Old sequence, should be ignored

	// Check size wasn't updated
	allBids := ob.GetAllBids()
	if len(allBids) != 1 {
		t.Fatal("Expected 1 bid")
	}
	if !allBids[0].Size.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Size = %s, want 100 (old diff should be ignored)", allBids[0].Size)
	}
}

func TestOrderBookGetBBO(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	// Empty orderbook
	bbo := ob.GetBBO()
	if bbo.BestBid != nil {
		t.Error("BestBid should be nil for empty orderbook")
	}
	if bbo.BestAsk != nil {
		t.Error("BestAsk should be nil for empty orderbook")
	}

	// Add data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}
	ob.ApplySnapshot(bids, asks, 1, nowMs())

	bbo = ob.GetBBO()
	if bbo.BestBid == nil {
		t.Fatal("BestBid should not be nil")
	}
	if bbo.BestAsk == nil {
		t.Fatal("BestAsk should not be nil")
	}
	if !bbo.BestBid.Price.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("BestBid.Price = %s, want 0.50", bbo.BestBid.Price)
	}
	if !bbo.BestAsk.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("BestAsk.Price = %s, want 0.55", bbo.BestAsk.Price)
	}
}

func TestOrderBookGetDepth(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
		{Price: decimal.NewFromFloat(0.48), Size: decimal.NewFromInt(300)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
		{Price: decimal.NewFromFloat(0.57), Size: decimal.NewFromInt(100)},
	}
	ob.ApplySnapshot(bids, asks, 1, nowMs())

	// Get top 2 levels
	depth := ob.GetDepth(2)
	if len(depth.Bids) != 2 {
		t.Errorf("Bids count = %d, want 2", len(depth.Bids))
	}
	if len(depth.Asks) != 2 {
		t.Errorf("Asks count = %d, want 2", len(depth.Asks))
	}

	// Verify sorting - bids descending
	if !depth.Bids[0].Price.Equal(decimal.NewFromFloat(0.50)) {
		t.Errorf("First bid price = %s, want 0.50", depth.Bids[0].Price)
	}
	if !depth.Bids[1].Price.Equal(decimal.NewFromFloat(0.49)) {
		t.Errorf("Second bid price = %s, want 0.49", depth.Bids[1].Price)
	}

	// Verify sorting - asks ascending
	if !depth.Asks[0].Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("First ask price = %s, want 0.55", depth.Asks[0].Price)
	}
	if !depth.Asks[1].Price.Equal(decimal.NewFromFloat(0.56)) {
		t.Errorf("Second ask price = %s, want 0.56", depth.Asks[1].Price)
	}
}

func TestOrderBookGetAllBids(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.48), Size: decimal.NewFromInt(300)},
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	ob.ApplySnapshot(bids, nil, 1, nowMs())

	allBids := ob.GetAllBids()
	if len(allBids) != 3 {
		t.Fatalf("Bids count = %d, want 3", len(allBids))
	}

	// Verify descending order
	if !allBids[0].Price.GreaterThan(allBids[1].Price) {
		t.Error("Bids should be sorted descending")
	}
	if !allBids[1].Price.GreaterThan(allBids[2].Price) {
		t.Error("Bids should be sorted descending")
	}
}

func TestOrderBookGetAllAsks(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.57), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}
	ob.ApplySnapshot(nil, asks, 1, nowMs())

	allAsks := ob.GetAllAsks()
	if len(allAsks) != 3 {
		t.Fatalf("Asks count = %d, want 3", len(allAsks))
	}

	// Verify ascending order
	if !allAsks[0].Price.LessThan(allAsks[1].Price) {
		t.Error("Asks should be sorted ascending")
	}
	if !allAsks[1].Price.LessThan(allAsks[2].Price) {
		t.Error("Asks should be sorted ascending")
	}
}

func TestOrderBookScanBidsAbove(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
		{Price: decimal.NewFromFloat(0.48), Size: decimal.NewFromInt(300)},
	}
	ob.ApplySnapshot(bids, nil, 1, nowMs())

	// Scan bids at or above 0.49
	result := ob.ScanBidsAbove(decimal.NewFromFloat(0.49))
	if len(result) != 2 {
		t.Errorf("Result count = %d, want 2", len(result))
	}
}

func TestOrderBookScanAsksBelow(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
		{Price: decimal.NewFromFloat(0.57), Size: decimal.NewFromInt(100)},
	}
	ob.ApplySnapshot(nil, asks, 1, nowMs())

	// Scan asks at or below 0.56
	result := ob.ScanAsksBelow(decimal.NewFromFloat(0.56))
	if len(result) != 2 {
		t.Errorf("Result count = %d, want 2", len(result))
	}
}

func TestOrderBookGetMidPrice(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	// Empty orderbook
	midPrice := ob.GetMidPrice()
	if !midPrice.IsZero() {
		t.Errorf("MidPrice for empty orderbook = %s, want 0", midPrice)
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(50)},
	}
	ob.ApplySnapshot(bids, asks, 1, nowMs())

	midPrice = ob.GetMidPrice()
	expected := decimal.NewFromFloat(0.55) // (0.50 + 0.60) / 2
	if !midPrice.Equal(expected) {
		t.Errorf("MidPrice = %s, want %s", midPrice, expected)
	}
}

func TestOrderBookGetSpread(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	// Empty orderbook
	spread := ob.GetSpread()
	if !spread.IsZero() {
		t.Errorf("Spread for empty orderbook = %s, want 0", spread)
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	ob.ApplySnapshot(bids, asks, 1, nowMs())

	spread = ob.GetSpread()
	expected := decimal.NewFromFloat(0.05) // 0.55 - 0.50
	if !spread.Equal(expected) {
		t.Errorf("Spread = %s, want %s", spread, expected)
	}
}

func TestOrderBookGetSpreadBps(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	// Empty orderbook
	spreadBps := ob.GetSpreadBps()
	if !spreadBps.IsZero() {
		t.Errorf("SpreadBps for empty orderbook = %s, want 0", spreadBps)
	}

	// With data
	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.60), Size: decimal.NewFromInt(50)},
	}
	ob.ApplySnapshot(bids, asks, 1, nowMs())

	spreadBps = ob.GetSpreadBps()
	// Spread = 0.10, MidPrice = 0.55, SpreadBps = 0.10 / 0.55 * 10000 ≈ 1818
	if spreadBps.LessThan(decimal.NewFromInt(1800)) || spreadBps.GreaterThan(decimal.NewFromInt(1900)) {
		t.Errorf("SpreadBps = %s, expected ~1818", spreadBps)
	}
}

func TestOrderBookTotalBidSize(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
		{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
	}
	ob.ApplySnapshot(bids, nil, 1, nowMs())

	total := ob.TotalBidSize()
	expected := decimal.NewFromInt(300)
	if !total.Equal(expected) {
		t.Errorf("TotalBidSize() = %s, want %s", total, expected)
	}
}

func TestOrderBookTotalAskSize(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
		{Price: decimal.NewFromFloat(0.56), Size: decimal.NewFromInt(75)},
	}
	ob.ApplySnapshot(nil, asks, 1, nowMs())

	total := ob.TotalAskSize()
	expected := decimal.NewFromInt(125)
	if !total.Equal(expected) {
		t.Errorf("TotalAskSize() = %s, want %s", total, expected)
	}
}

func TestOrderBookClear(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	ob.ApplySnapshot(bids, asks, 5, nowMs())

	ob.Clear()

	if !ob.IsEmpty() {
		t.Error("Orderbook should be empty after Clear()")
	}
	if ob.Sequence() != 0 {
		t.Errorf("Sequence() = %d, want 0 after Clear()", ob.Sequence())
	}
}

func TestOrderBookClone(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	bids := []PriceLevel{
		{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
	}
	asks := []PriceLevel{
		{Price: decimal.NewFromFloat(0.55), Size: decimal.NewFromInt(50)},
	}
	ob.ApplySnapshot(bids, asks, 5, nowMs())

	clone := ob.Clone()

	if clone.MarketID() != ob.MarketID() {
		t.Errorf("Clone MarketID = %d, want %d", clone.MarketID(), ob.MarketID())
	}
	if clone.TokenID() != ob.TokenID() {
		t.Errorf("Clone TokenID = %s, want %s", clone.TokenID(), ob.TokenID())
	}
	if clone.Sequence() != ob.Sequence() {
		t.Errorf("Clone Sequence = %d, want %d", clone.Sequence(), ob.Sequence())
	}
	if clone.BidCount() != ob.BidCount() {
		t.Errorf("Clone BidCount = %d, want %d", clone.BidCount(), ob.BidCount())
	}
	if clone.AskCount() != ob.AskCount() {
		t.Errorf("Clone AskCount = %d, want %d", clone.AskCount(), ob.AskCount())
	}

	// Verify independence
	ob.Clear()
	if clone.IsEmpty() {
		t.Error("Clone should be independent from original")
	}
}

func TestOrderBookConcurrentAccess(t *testing.T) {
	ob := NewOrderBook(123, "token-1")

	// Run concurrent reads and writes
	done := make(chan bool)

	// Writer
	go func() {
		for i := 0; i < 100; i++ {
			bids := []PriceLevel{
				{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(int64(i))},
			}
			ob.ApplySnapshot(bids, nil, int64(i), nowMs())
		}
		done <- true
	}()

	// Reader
	go func() {
		for i := 0; i < 100; i++ {
			ob.GetBBO()
			ob.GetDepth(5)
			ob.GetAllBids()
		}
		done <- true
	}()

	<-done
	<-done
}
