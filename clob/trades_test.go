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

// ========================================
// GetMyTrades 单元测试
// 对应 Python SDK 的 get_my_trades()
// ========================================

func TestGetMyTrades(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证端点路径（对应官方 SDK /openapi/trade）
		if r.URL.Path != "/openapi/trade" {
			t.Errorf("Expected /openapi/trade, got %s", r.URL.Path)
		}

		// 验证请求方法
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// 验证 API Key 请求头
		if r.Header.Get("apikey") == "" {
			t.Error("Expected apikey header to be set")
		}

		// 模拟真实 API 响应格式（2026-01-22 验证）
		resp := MyTradesResponse{
			Errno:  0,
			Errmsg: "success",
		}
		resp.Result.List = []*MyTrade{
			{
				OrderNo:         "order-1",
				TradeNo:         "trade-1",
				MarketID:        123,
				MarketTitle:     "Test Market",
				RootMarketID:    100,
				RootMarketTitle: "Root Market",
				Side:            "Buy", // 字符串！真实 API 返回字符串
				Outcome:         "YES",
				OutcomeSide:     1,
				OutcomeSideEnum: "Yes",
				Price:           "0.55",
				Shares:          "100",
				Amount:          "55",
				Status:          2,
				StatusEnum:      "Finished",
				CreatedAt:       1704067200,
			},
			{
				OrderNo:         "order-2",
				TradeNo:         "trade-2",
				MarketID:        123,
				MarketTitle:     "Test Market",
				RootMarketID:    100,
				RootMarketTitle: "Root Market",
				Side:            "Sell", // 字符串！
				Outcome:         "NO",
				OutcomeSide:     2,
				OutcomeSideEnum: "No",
				Price:           "0.60",
				Shares:          "50",
				Amount:          "30",
				Status:          2,
				StatusEnum:      "Finished",
				CreatedAt:       1704153600,
			},
		}
		resp.Result.Total = 2

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{
		Host:    server.URL,
		APIKey:  "test-api-key",
		ChainID: 56,
	})

	trades, total, err := client.GetMyTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMyTrades() error: %v", err)
	}

	if total != 2 {
		t.Errorf("Total = %d, want 2", total)
	}
	if len(trades) != 2 {
		t.Errorf("Trades count = %d, want 2", len(trades))
	}
	if trades[0].TradeNo != "trade-1" {
		t.Errorf("First trade TradeNo = %s, want trade-1", trades[0].TradeNo)
	}
	if trades[0].Side != "Buy" {
		t.Errorf("First trade side = %s, want Buy", trades[0].Side)
	}
	if trades[0].MarketTitle != "Test Market" {
		t.Errorf("First trade MarketTitle = %s, want 'Test Market'", trades[0].MarketTitle)
	}
}

func TestGetMyTrades_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证查询参数
		if r.URL.Query().Get("market_id") != "123" {
			t.Errorf("Expected market_id=123, got %s", r.URL.Query().Get("market_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("chain_id") != "56" {
			t.Errorf("Expected chain_id=56, got %s", r.URL.Query().Get("chain_id"))
		}

		resp := MyTradesResponse{Errno: 0, Errmsg: "success"}
		resp.Result.List = []*MyTrade{}
		resp.Result.Total = 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{
		Host:    server.URL,
		APIKey:  "test-api-key",
		ChainID: 56,
	})

	_, _, err := client.GetMyTrades(context.Background(), &MyTradesParams{
		MarketID: 123,
		Page:     2,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("GetMyTrades() error: %v", err)
	}
}

func TestGetMyTrades_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MyTradesResponse{
			Errno:  1001,
			Errmsg: "API key invalid",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL, APIKey: "invalid-key"})

	_, _, err := client.GetMyTrades(context.Background(), nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !contains(err.Error(), "1001") {
		t.Errorf("Error should contain errno 1001, got: %v", err)
	}
}

func TestGetMyTrades_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MyTradesResponse{Errno: 0, Errmsg: "success"}
		resp.Result.List = []*MyTrade{}
		resp.Result.Total = 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL, APIKey: "test-key"})

	trades, total, err := client.GetMyTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMyTrades() error: %v", err)
	}

	if total != 0 {
		t.Errorf("Total = %d, want 0", total)
	}
	if len(trades) != 0 {
		t.Errorf("Trades count = %d, want 0", len(trades))
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
