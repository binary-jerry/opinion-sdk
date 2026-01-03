package markets

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMarketStatusConstants(t *testing.T) {
	if MarketStatusActive != "active" {
		t.Errorf("MarketStatusActive = %s, want active", MarketStatusActive)
	}
	if MarketStatusResolved != "resolved" {
		t.Errorf("MarketStatusResolved = %s, want resolved", MarketStatusResolved)
	}
	if MarketStatusPaused != "paused" {
		t.Errorf("MarketStatusPaused = %s, want paused", MarketStatusPaused)
	}
}

func TestMarketTypeConstants(t *testing.T) {
	if MarketTypeBinary != "binary" {
		t.Errorf("MarketTypeBinary = %s, want binary", MarketTypeBinary)
	}
	if MarketTypeCategorical != "categorical" {
		t.Errorf("MarketTypeCategorical = %s, want categorical", MarketTypeCategorical)
	}
}

func TestMarketJSON(t *testing.T) {
	market := &Market{
		ID:       "test-id",
		Question: "Will it rain?",
		Type:     MarketTypeBinary,
		Status:   MarketStatusActive,
		Volume:   decimal.NewFromInt(1000),
	}

	data, err := json.Marshal(market)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Market
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != market.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, market.ID)
	}
	if decoded.Question != market.Question {
		t.Errorf("Question = %s, want %s", decoded.Question, market.Question)
	}
}

func TestTokenJSON(t *testing.T) {
	token := &Token{
		ID:       "token-1",
		TokenID:  "12345",
		Outcome:  "Yes",
		Price:    decimal.NewFromFloat(0.65),
		Volume:   decimal.NewFromInt(500),
		MarketID: "market-1",
	}

	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Token
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.TokenID != token.TokenID {
		t.Errorf("TokenID = %s, want %s", decoded.TokenID, token.TokenID)
	}
	if !decoded.Price.Equal(token.Price) {
		t.Errorf("Price = %s, want %s", decoded.Price, token.Price)
	}
}

func TestOrderbookJSON(t *testing.T) {
	orderbook := &Orderbook{
		TokenID: "12345",
		Bids: []*OrderbookLevel{
			{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
			{Price: decimal.NewFromFloat(0.49), Size: decimal.NewFromInt(200)},
		},
		Asks: []*OrderbookLevel{
			{Price: decimal.NewFromFloat(0.51), Size: decimal.NewFromInt(150)},
		},
	}

	data, err := json.Marshal(orderbook)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Orderbook
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.Bids) != 2 {
		t.Errorf("Bids length = %d, want 2", len(decoded.Bids))
	}
	if len(decoded.Asks) != 1 {
		t.Errorf("Asks length = %d, want 1", len(decoded.Asks))
	}
}

func TestLatestPriceJSON(t *testing.T) {
	price := &LatestPrice{
		TokenID:   "12345",
		Price:     decimal.NewFromFloat(0.55),
		Timestamp: 1234567890,
	}

	data, err := json.Marshal(price)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded LatestPrice
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Timestamp != price.Timestamp {
		t.Errorf("Timestamp = %d, want %d", decoded.Timestamp, price.Timestamp)
	}
}

func TestPriceHistoryJSON(t *testing.T) {
	history := &PriceHistory{
		TokenID: "12345",
		Points: []*PricePoint{
			{
				Timestamp: 1234567890,
				Open:      decimal.NewFromFloat(0.50),
				High:      decimal.NewFromFloat(0.55),
				Low:       decimal.NewFromFloat(0.48),
				Close:     decimal.NewFromFloat(0.52),
				Volume:    decimal.NewFromInt(1000),
			},
		},
	}

	data, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded PriceHistory
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.Points) != 1 {
		t.Errorf("Points length = %d, want 1", len(decoded.Points))
	}
}

func TestQuoteTokenJSON(t *testing.T) {
	qt := &QuoteToken{
		ID:       "usdt",
		Symbol:   "USDT",
		Name:     "Tether USD",
		Address:  "0x55d398326f99059fF775485246999027B3197955",
		Decimals: 18,
	}

	data, err := json.Marshal(qt)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded QuoteToken
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Symbol != qt.Symbol {
		t.Errorf("Symbol = %s, want %s", decoded.Symbol, qt.Symbol)
	}
	if decoded.Decimals != qt.Decimals {
		t.Errorf("Decimals = %d, want %d", decoded.Decimals, qt.Decimals)
	}
}

func TestFeeRatesJSON(t *testing.T) {
	fees := &FeeRates{
		MakerFeeRate: decimal.NewFromFloat(0),
		TakerFeeRate: decimal.NewFromFloat(0.01),
	}

	data, err := json.Marshal(fees)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded FeeRates
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !decoded.TakerFeeRate.Equal(fees.TakerFeeRate) {
		t.Errorf("TakerFeeRate = %s, want %s", decoded.TakerFeeRate, fees.TakerFeeRate)
	}
}

func TestMarketListParams(t *testing.T) {
	params := &MarketListParams{
		Page:     1,
		Limit:    20,
		Status:   MarketStatusActive,
		Type:     MarketTypeBinary,
		Category: "politics",
	}

	if params.Page != 1 {
		t.Errorf("Page = %d, want 1", params.Page)
	}
	if params.Limit != 20 {
		t.Errorf("Limit = %d, want 20", params.Limit)
	}
	if params.Status != MarketStatusActive {
		t.Errorf("Status = %s, want active", params.Status)
	}
}

func TestPriceHistoryParams(t *testing.T) {
	params := &PriceHistoryParams{
		TokenID:  "12345",
		Interval: "1h",
		StartAt:  1000000000,
		EndAt:    2000000000,
	}

	if params.TokenID != "12345" {
		t.Errorf("TokenID = %s, want 12345", params.TokenID)
	}
	if params.Interval != "1h" {
		t.Errorf("Interval = %s, want 1h", params.Interval)
	}
}

func TestChildMarketJSON(t *testing.T) {
	cm := &ChildMarket{
		ID:       "child-1",
		Question: "Who will win?",
		Tokens: []*Token{
			{ID: "token-1", Outcome: "Team A"},
			{ID: "token-2", Outcome: "Team B"},
		},
	}

	data, err := json.Marshal(cm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded ChildMarket
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.Tokens) != 2 {
		t.Errorf("Tokens length = %d, want 2", len(decoded.Tokens))
	}
}

func TestAPIResponse(t *testing.T) {
	resp := &APIResponse{
		Code:    0,
		Message: "success",
	}

	if resp.Code != 0 {
		t.Errorf("Code = %d, want 0", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("Message = %s, want success", resp.Message)
	}
}
