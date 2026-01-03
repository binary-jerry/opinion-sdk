package clob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

func TestGetTrades(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Errorf("Expected /trades, got %s", r.URL.Path)
		}

		resp := TradeListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.Trades = []*Trade{
			{
				ID:        "trade-1",
				OrderID:   "order-1",
				MarketID:  "market-1",
				TokenID:   "12345",
				Side:      "BUY",
				Price:     decimal.NewFromFloat(0.55),
				Size:      decimal.NewFromInt(100),
				Fee:       decimal.NewFromFloat(0.5),
				Timestamp: "2024-01-01T00:00:00Z",
			},
			{
				ID:        "trade-2",
				OrderID:   "order-2",
				MarketID:  "market-1",
				TokenID:   "12345",
				Side:      "SELL",
				Price:     decimal.NewFromFloat(0.60),
				Size:      decimal.NewFromInt(50),
				Fee:       decimal.NewFromFloat(0.3),
				Timestamp: "2024-01-02T00:00:00Z",
			},
		}
		resp.Result.Total = 2

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	trades, err := client.GetTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}

	if len(trades) != 2 {
		t.Errorf("Trades count = %d, want 2", len(trades))
	}
	if trades[0].ID != "trade-1" {
		t.Errorf("First trade ID = %s, want trade-1", trades[0].ID)
	}
	if trades[0].Side != "BUY" {
		t.Errorf("First trade side = %s, want BUY", trades[0].Side)
	}
}

func TestGetTradesWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("marketId") != "market-1" {
			t.Errorf("Expected marketId=market-1, got %s", r.URL.Query().Get("marketId"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		resp := TradeListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.Trades = []*Trade{}
		resp.Result.Total = 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.GetTrades(context.Background(), &TradeListParams{
		MarketID: "market-1",
		Page:     2,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}
}

func TestGetTradesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TradeListResponse{
			APIResponse: APIResponse{Code: 1, Message: "error"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.GetTrades(context.Background(), nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetTradesEmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TradeListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.Trades = []*Trade{}
		resp.Result.Total = 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	trades, err := client.GetTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}

	if len(trades) != 0 {
		t.Errorf("Trades count = %d, want 0", len(trades))
	}
}
