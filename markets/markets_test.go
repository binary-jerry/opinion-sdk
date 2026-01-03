package markets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

func TestGetMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/market" {
			t.Errorf("Expected /market, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", r.URL.Query().Get("page"))
		}

		resp := MarketListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.Markets = []*Market{
			{ID: "1", Question: "Test market 1"},
			{ID: "2", Question: "Test market 2"},
		}
		resp.Result.Total = 2
		resp.Result.Page = 1
		resp.Result.Limit = 20

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	result, err := client.GetMarkets(context.Background(), &MarketListParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("GetMarkets() error: %v", err)
	}

	if len(result.Result.Markets) != 2 {
		t.Errorf("Markets count = %d, want 2", len(result.Result.Markets))
	}
	if result.Result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Result.Total)
	}
}

func TestGetMarketsDefaultParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MarketListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.GetMarkets(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMarkets(nil) error: %v", err)
	}
}

func TestGetMarketsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MarketListResponse{
			APIResponse: APIResponse{Code: 1, Message: "error"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.GetMarkets(context.Background(), nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/market/test-id" {
			t.Errorf("Expected /market/test-id, got %s", r.URL.Path)
		}

		resp := MarketDetailResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &Market{
				ID:       "test-id",
				Question: "Test market",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	market, err := client.GetMarket(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("GetMarket() error: %v", err)
	}

	if market.ID != "test-id" {
		t.Errorf("Market ID = %s, want test-id", market.ID)
	}
}

func TestGetMarketEmptyID(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetMarket(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty market ID")
	}
}

func TestGetCategoricalMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/market/categorical/test-id" {
			t.Errorf("Expected /market/categorical/test-id, got %s", r.URL.Path)
		}

		resp := MarketDetailResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &Market{
				ID:   "test-id",
				Type: MarketTypeCategorical,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	market, err := client.GetCategoricalMarket(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("GetCategoricalMarket() error: %v", err)
	}

	if market.Type != MarketTypeCategorical {
		t.Errorf("Market Type = %s, want categorical", market.Type)
	}
}

func TestGetCategoricalMarketEmptyID(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetCategoricalMarket(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty market ID")
	}
}

func TestGetOrderbook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token/orderbook" {
			t.Errorf("Expected /token/orderbook, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("tokenId") != "12345" {
			t.Errorf("Expected tokenId=12345, got %s", r.URL.Query().Get("tokenId"))
		}

		resp := OrderbookResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &Orderbook{
				TokenID: "12345",
				Bids: []*OrderbookLevel{
					{Price: decimal.NewFromFloat(0.50), Size: decimal.NewFromInt(100)},
				},
				Asks: []*OrderbookLevel{
					{Price: decimal.NewFromFloat(0.51), Size: decimal.NewFromInt(50)},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	orderbook, err := client.GetOrderbook(context.Background(), "12345")
	if err != nil {
		t.Fatalf("GetOrderbook() error: %v", err)
	}

	if orderbook.TokenID != "12345" {
		t.Errorf("TokenID = %s, want 12345", orderbook.TokenID)
	}
	if len(orderbook.Bids) != 1 {
		t.Errorf("Bids count = %d, want 1", len(orderbook.Bids))
	}
}

func TestGetOrderbookEmptyTokenID(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetOrderbook(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty token ID")
	}
}

func TestGetLatestPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token/latest-price" {
			t.Errorf("Expected /token/latest-price, got %s", r.URL.Path)
		}

		resp := LatestPriceResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &LatestPrice{
				TokenID:   "12345",
				Price:     decimal.NewFromFloat(0.55),
				Timestamp: 1234567890,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	price, err := client.GetLatestPrice(context.Background(), "12345")
	if err != nil {
		t.Fatalf("GetLatestPrice() error: %v", err)
	}

	if price.TokenID != "12345" {
		t.Errorf("TokenID = %s, want 12345", price.TokenID)
	}
	if !price.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("Price = %s, want 0.55", price.Price)
	}
}

func TestGetLatestPriceEmptyTokenID(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetLatestPrice(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty token ID")
	}
}

func TestGetPriceHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token/price-history" {
			t.Errorf("Expected /token/price-history, got %s", r.URL.Path)
		}

		resp := PriceHistoryResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &PriceHistory{
				TokenID: "12345",
				Points: []*PricePoint{
					{
						Timestamp: 1234567890,
						Open:      decimal.NewFromFloat(0.50),
						Close:     decimal.NewFromFloat(0.55),
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	history, err := client.GetPriceHistory(context.Background(), &PriceHistoryParams{
		TokenID:  "12345",
		Interval: "1h",
	})
	if err != nil {
		t.Fatalf("GetPriceHistory() error: %v", err)
	}

	if history.TokenID != "12345" {
		t.Errorf("TokenID = %s, want 12345", history.TokenID)
	}
	if len(history.Points) != 1 {
		t.Errorf("Points count = %d, want 1", len(history.Points))
	}
}

func TestGetPriceHistoryNilParams(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetPriceHistory(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil params")
	}
}

func TestGetPriceHistoryEmptyTokenID(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetPriceHistory(context.Background(), &PriceHistoryParams{})
	if err == nil {
		t.Error("Expected error for empty token ID")
	}
}

func TestGetQuoteTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/quoteToken" {
			t.Errorf("Expected /quoteToken, got %s", r.URL.Path)
		}

		resp := QuoteTokenListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: []*QuoteToken{
				{ID: "usdt", Symbol: "USDT", Decimals: 18},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	tokens, err := client.GetQuoteTokens(context.Background())
	if err != nil {
		t.Fatalf("GetQuoteTokens() error: %v", err)
	}

	if len(tokens) != 1 {
		t.Errorf("Tokens count = %d, want 1", len(tokens))
	}
	if tokens[0].Symbol != "USDT" {
		t.Errorf("Token symbol = %s, want USDT", tokens[0].Symbol)
	}
}

func TestGetFeeRates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fee-rates" {
			t.Errorf("Expected /fee-rates, got %s", r.URL.Path)
		}

		resp := FeeRatesResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &FeeRates{
				MakerFeeRate: decimal.Zero,
				TakerFeeRate: decimal.NewFromFloat(0.01),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{Host: server.URL})

	fees, err := client.GetFeeRates(context.Background())
	if err != nil {
		t.Fatalf("GetFeeRates() error: %v", err)
	}

	if !fees.MakerFeeRate.IsZero() {
		t.Errorf("MakerFeeRate = %s, want 0", fees.MakerFeeRate)
	}
	if !fees.TakerFeeRate.Equal(decimal.NewFromFloat(0.01)) {
		t.Errorf("TakerFeeRate = %s, want 0.01", fees.TakerFeeRate)
	}
}
