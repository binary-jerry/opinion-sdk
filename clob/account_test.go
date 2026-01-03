package clob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

func TestGetBalances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/balances" {
			t.Errorf("Expected /balances, got %s", r.URL.Path)
		}

		resp := BalanceListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: []*Balance{
				{TokenID: "usdt", Symbol: "USDT", Available: decimal.NewFromInt(1000)},
				{TokenID: "token-1", Symbol: "YES", Available: decimal.NewFromInt(100)},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	balances, err := client.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}

	if len(balances) != 2 {
		t.Errorf("Balances count = %d, want 2", len(balances))
	}
	if balances[0].Symbol != "USDT" {
		t.Errorf("First balance symbol = %s, want USDT", balances[0].Symbol)
	}
}

func TestGetBalancesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BalanceListResponse{
			APIResponse: APIResponse{Code: 1, Message: "error"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.GetBalances(context.Background())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetPositions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			t.Errorf("Expected /positions, got %s", r.URL.Path)
		}

		resp := PositionListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: []*Position{
				{
					MarketID: "market-1",
					TokenID:  "12345",
					Outcome:  "Yes",
					Size:     decimal.NewFromInt(100),
					PnL:      decimal.NewFromInt(10),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	positions, err := client.GetPositions(context.Background())
	if err != nil {
		t.Fatalf("GetPositions() error: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("Positions count = %d, want 1", len(positions))
	}
	if positions[0].MarketID != "market-1" {
		t.Errorf("Position marketID = %s, want market-1", positions[0].MarketID)
	}
}

func TestEnableTrading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/enable-trading" {
			t.Errorf("Expected /enable-trading, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var body EnableTradingRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.QuoteTokenID != "usdt" {
			t.Errorf("Expected quoteTokenId=usdt, got %s", body.QuoteTokenID)
		}

		resp := TxResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.TxHash = "0xabc123"
		resp.Result.Success = true

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	txResp, err := client.EnableTrading(context.Background(), "usdt")
	if err != nil {
		t.Fatalf("EnableTrading() error: %v", err)
	}

	if txResp.Result.TxHash != "0xabc123" {
		t.Errorf("TxHash = %s, want 0xabc123", txResp.Result.TxHash)
	}
}

func TestSplit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/split" {
			t.Errorf("Expected /split, got %s", r.URL.Path)
		}

		resp := TxResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.TxHash = "0xdef456"
		resp.Result.Success = true

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	txResp, err := client.Split(context.Background(), &SplitRequest{
		MarketID: "market-1",
		Amount:   decimal.NewFromInt(100),
	})
	if err != nil {
		t.Fatalf("Split() error: %v", err)
	}

	if txResp.Result.TxHash != "0xdef456" {
		t.Errorf("TxHash = %s, want 0xdef456", txResp.Result.TxHash)
	}
}

func TestSplitAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TxResponse{
			APIResponse: APIResponse{Code: 1, Message: "insufficient balance"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.Split(context.Background(), &SplitRequest{
		MarketID: "market-1",
		Amount:   decimal.NewFromInt(100),
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMerge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/merge" {
			t.Errorf("Expected /merge, got %s", r.URL.Path)
		}

		resp := TxResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.TxHash = "0xghi789"
		resp.Result.Success = true

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	txResp, err := client.Merge(context.Background(), &MergeRequest{
		MarketID: "market-1",
		Amount:   decimal.NewFromInt(50),
	})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	if !txResp.Result.Success {
		t.Error("Expected success=true")
	}
}

func TestRedeem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/redeem" {
			t.Errorf("Expected /redeem, got %s", r.URL.Path)
		}

		var body RedeemRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.MarketID != "market-1" {
			t.Errorf("Expected marketId=market-1, got %s", body.MarketID)
		}

		resp := TxResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.TxHash = "0xjkl012"
		resp.Result.Success = true

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	txResp, err := client.Redeem(context.Background(), &RedeemRequest{
		MarketID: "market-1",
	})
	if err != nil {
		t.Fatalf("Redeem() error: %v", err)
	}

	if txResp.Result.TxHash != "0xjkl012" {
		t.Errorf("TxHash = %s, want 0xjkl012", txResp.Result.TxHash)
	}
}
