package clob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

func TestPlaceOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/openapi/order" {
			t.Errorf("Expected /openapi/order, got %s", r.URL.Path)
		}

		// 使用实际的 API 响应结构
		resp := PlaceOrderResponse{
			Errno:  0,
			Errmsg: "success",
			Result: &PlaceOrderResult{
				OrderData: &PlaceOrderData{
					OrderID: "order-123",
					Status:  1, // pending
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &ClientConfig{
		Host:            server.URL,
		PrivateKey:      testPrivateKeyClient,
		ChainID:         56,
		ExchangeAddress: "0x0000000000000000000000000000000000000001",
	}
	client, _ := NewClient(config)

	order, err := client.PlaceOrder(context.Background(), &CreateOrderRequest{
		MarketID:                "market-1",
		TokenID:                 "12345",
		Side:                    OrderSideBuy,
		Type:                    OrderTypeLimit,
		Price:                   decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
	})
	if err != nil {
		t.Fatalf("PlaceOrder() error: %v", err)
	}

	if order.OrderID != "order-123" {
		t.Errorf("Order ID = %s, want order-123", order.OrderID)
	}
}

func TestPlaceOrderWithoutSigner(t *testing.T) {
	client, _ := NewClient(nil)

	_, err := client.PlaceOrder(context.Background(), &CreateOrderRequest{
		TokenID: "12345",
		Side:    OrderSideBuy,
		Price:   decimal.NewFromFloat(0.55),
	})
	if err == nil {
		t.Error("PlaceOrder() should fail without signer")
	}
}

func TestPlaceOrdersBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/batch" {
			t.Errorf("Expected /orders/batch, got %s", r.URL.Path)
		}

		resp := struct {
			APIResponse
			Result []*Order `json:"result"`
		}{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: []*Order{
				{ID: "order-1"},
				{ID: "order-2"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &ClientConfig{
		Host:            server.URL,
		PrivateKey:      testPrivateKeyClient,
		ChainID:         56,
		ExchangeAddress: "0x0000000000000000000000000000000000000001",
	}
	client, _ := NewClient(config)

	orders, err := client.PlaceOrdersBatch(context.Background(), []*CreateOrderRequest{
		{TokenID: "12345", Side: OrderSideBuy, Price: decimal.NewFromFloat(0.55), MakerAmountInQuoteToken: decimal.NewFromInt(100)},
		{TokenID: "12345", Side: OrderSideSell, Price: decimal.NewFromFloat(0.60), MakerAmountInBaseToken: decimal.NewFromInt(50)},
	})
	if err != nil {
		t.Fatalf("PlaceOrdersBatch() error: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Orders count = %d, want 2", len(orders))
	}
}

func TestPlaceOrdersBatchWithoutSigner(t *testing.T) {
	client, _ := NewClient(nil)

	_, err := client.PlaceOrdersBatch(context.Background(), []*CreateOrderRequest{})
	if err == nil {
		t.Error("PlaceOrdersBatch() should fail without signer")
	}
}

func TestGetOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/order-123" {
			t.Errorf("Expected /order/order-123, got %s", r.URL.Path)
		}

		resp := OrderDetailResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
			Result: &Order{
				ID:     "order-123",
				Status: OrderStatusPending,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	order, err := client.GetOrder(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("GetOrder() error: %v", err)
	}

	if order.ID != "order-123" {
		t.Errorf("Order ID = %s, want order-123", order.ID)
	}
}

func TestGetOrderEmptyID(t *testing.T) {
	client, _ := NewClient(nil)

	_, err := client.GetOrder(context.Background(), "")
	if err == nil {
		t.Error("GetOrder() should fail with empty ID")
	}
}

func TestGetOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Errorf("Expected /orders, got %s", r.URL.Path)
		}

		resp := OrderListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.Orders = []*Order{
			{ID: "order-1"},
			{ID: "order-2"},
		}
		resp.Result.Total = 2

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	orders, err := client.GetOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetOrders() error: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Orders count = %d, want 2", len(orders))
	}
}

func TestGetOrdersWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "pending" {
			t.Errorf("Expected status=pending, got %s", r.URL.Query().Get("status"))
		}

		resp := OrderListResponse{
			APIResponse: APIResponse{Code: 0, Message: "success"},
		}
		resp.Result.Orders = []*Order{}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	_, err := client.GetOrders(context.Background(), &OrderListParams{
		Status: OrderStatusPending,
	})
	if err != nil {
		t.Fatalf("GetOrders() error: %v", err)
	}
}

func TestCancelOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/openapi/order/cancel" {
			t.Errorf("Expected /openapi/order/cancel, got %s", r.URL.Path)
		}

		var body CancelOrderRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.OrderID != "order-123" {
			t.Errorf("Expected orderId=order-123, got %s", body.OrderID)
		}

		resp := APIResponse{Code: 0, Message: "success"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	err := client.CancelOrder(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("CancelOrder() error: %v", err)
	}
}

func TestCancelOrderEmptyID(t *testing.T) {
	client, _ := NewClient(nil)

	err := client.CancelOrder(context.Background(), "")
	if err == nil {
		t.Error("CancelOrder() should fail with empty ID")
	}
}

func TestCancelOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/cancel" {
			t.Errorf("Expected /orders/cancel, got %s", r.URL.Path)
		}

		var body CancelOrdersRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.OrderIDs) != 2 {
			t.Errorf("Expected 2 orderIds, got %d", len(body.OrderIDs))
		}

		resp := APIResponse{Code: 0, Message: "success"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	err := client.CancelOrders(context.Background(), []string{"order-1", "order-2"})
	if err != nil {
		t.Fatalf("CancelOrders() error: %v", err)
	}
}

func TestCancelOrdersEmptyIDs(t *testing.T) {
	client, _ := NewClient(nil)

	err := client.CancelOrders(context.Background(), []string{})
	if err == nil {
		t.Error("CancelOrders() should fail with empty IDs")
	}
}

func TestCancelAllOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/cancel-all" {
			t.Errorf("Expected /orders/cancel-all, got %s", r.URL.Path)
		}

		resp := APIResponse{Code: 0, Message: "success"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL})

	err := client.CancelAllOrders(context.Background(), &CancelAllOrdersRequest{
		MarketID: "market-1",
	})
	if err != nil {
		t.Fatalf("CancelAllOrders() error: %v", err)
	}
}

// ========================================
// GetMyOrders 单元测试
// 对应 Python SDK 的 get_my_orders()
// ========================================

func TestGetMyOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证端点路径（对应官方 SDK /openapi/order）
		if r.URL.Path != "/openapi/order" {
			t.Errorf("Expected /openapi/order, got %s", r.URL.Path)
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
		resp := MyOrdersResponse{
			Errno:  0,
			Errmsg: "success",
		}
		resp.Result.List = []*MyOrder{
			{
				OrderID:           "order-1",
				TransNo:           "order-1",
				MarketID:          123,
				MarketTitle:       "Test Market",
				RootMarketID:      100,
				RootMarketTitle:   "Root Market",
				Side:              1, // BUY (订单 API 返回数字)
				SideEnum:          "Buy",
				OutcomeSide:       1,
				OutcomeSideEnum:   "Yes",
				TradingMethod:     2,
				TradingMethodEnum: "Limit",
				Price:             "0.55",
				OrderShares:       "100",
				OrderAmount:       "55",
				FilledShares:      "50",
				FilledAmount:      "27.5",
				Status:            1, // pending
				StatusEnum:        "Pending",
				CreatedAt:         1704067200,
			},
			{
				OrderID:           "order-2",
				TransNo:           "order-2",
				MarketID:          123,
				MarketTitle:       "Test Market",
				RootMarketID:      100,
				RootMarketTitle:   "Root Market",
				Side:              2, // SELL
				SideEnum:          "Sell",
				OutcomeSide:       2,
				OutcomeSideEnum:   "No",
				TradingMethod:     2,
				TradingMethodEnum: "Limit",
				Price:             "0.60",
				OrderShares:       "50",
				OrderAmount:       "30",
				FilledShares:      "50",
				FilledAmount:      "30",
				Status:            2, // finished
				StatusEnum:        "Finished",
				CreatedAt:         1704153600,
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

	orders, total, err := client.GetMyOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMyOrders() error: %v", err)
	}

	if total != 2 {
		t.Errorf("Total = %d, want 2", total)
	}
	if len(orders) != 2 {
		t.Errorf("Orders count = %d, want 2", len(orders))
	}
	if orders[0].OrderID != "order-1" {
		t.Errorf("First order ID = %s, want order-1", orders[0].OrderID)
	}
	if orders[0].Side != 1 {
		t.Errorf("First order side = %d, want 1 (BUY)", orders[0].Side)
	}
	if orders[0].Status != 1 {
		t.Errorf("First order status = %d, want 1 (pending)", orders[0].Status)
	}
}

func TestGetMyOrders_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证查询参数
		if r.URL.Query().Get("market_id") != "123" {
			t.Errorf("Expected market_id=123, got %s", r.URL.Query().Get("market_id"))
		}
		if r.URL.Query().Get("status") != "1" {
			t.Errorf("Expected status=1, got %s", r.URL.Query().Get("status"))
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

		resp := MyOrdersResponse{Errno: 0, Errmsg: "success"}
		resp.Result.List = []*MyOrder{}
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

	_, _, err := client.GetMyOrders(context.Background(), &MyOrdersParams{
		MarketID: 123,
		Status:   "1",
		Page:     2,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("GetMyOrders() error: %v", err)
	}
}

func TestGetMyOrders_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MyOrdersResponse{
			Errno:  1001,
			Errmsg: "API key invalid",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL, APIKey: "invalid-key"})

	_, _, err := client.GetMyOrders(context.Background(), nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetMyOrders_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MyOrdersResponse{Errno: 0, Errmsg: "success"}
		resp.Result.List = []*MyOrder{}
		resp.Result.Total = 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL, APIKey: "test-key"})

	orders, total, err := client.GetMyOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMyOrders() error: %v", err)
	}

	if total != 0 {
		t.Errorf("Total = %d, want 0", total)
	}
	if len(orders) != 0 {
		t.Errorf("Orders count = %d, want 0", len(orders))
	}
}

// ========================================
// GetOrderByID 单元测试
// 对应 Python SDK 的 get_order_by_id()
// ========================================

func TestGetOrderByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证端点路径（对应官方 SDK /openapi/order/{orderId}）
		if r.URL.Path != "/openapi/order/order-123" {
			t.Errorf("Expected /openapi/order/order-123, got %s", r.URL.Path)
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
		// 注意：result 中有 orderData 包装层
		resp := map[string]interface{}{
			"errno":  0,
			"errmsg": "success",
			"result": map[string]interface{}{
				"orderData": map[string]interface{}{
					"orderId":           "order-123",
					"transNo":           "order-123",
					"marketId":          123,
					"marketTitle":       "Test Market",
					"rootMarketId":      100,
					"rootMarketTitle":   "Root Market",
					"side":              1, // BUY (订单 API 返回数字)
					"sideEnum":          "Buy",
					"outcomeSide":       1,
					"outcomeSideEnum":   "Yes",
					"tradingMethod":     2,
					"tradingMethodEnum": "Limit",
					"price":             "0.55",
					"orderShares":       "100",
					"orderAmount":       "55",
					"filledShares":      "100",
					"filledAmount":      "55",
					"status":            2, // finished
					"statusEnum":        "Finished",
					"createdAt":         1704067200,
					"trades": []map[string]interface{}{
						{
							"orderNo":   "order-123",
							"tradeNo":   "trade-1",
							"side":      "Buy",
							"price":     "0.55",
							"shares":    "50",
							"amount":    "27.5",
							"status":    2,
							"statusEnum": "Finished",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{
		Host:   server.URL,
		APIKey: "test-api-key",
	})

	order, err := client.GetOrderByID(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("GetOrderByID() error: %v", err)
	}

	if order.OrderID != "order-123" {
		t.Errorf("Order ID = %s, want order-123", order.OrderID)
	}
	if order.Status != 2 {
		t.Errorf("Order status = %d, want 2 (finished)", order.Status)
	}
	if order.MarketTitle != "Test Market" {
		t.Errorf("Order MarketTitle = %s, want 'Test Market'", order.MarketTitle)
	}
	if len(order.Trades) != 1 {
		t.Errorf("Trades count = %d, want 1", len(order.Trades))
	}
}

func TestGetOrderByID_EmptyID(t *testing.T) {
	client, _ := NewClient(&ClientConfig{APIKey: "test-key"})

	_, err := client.GetOrderByID(context.Background(), "")
	if err == nil {
		t.Error("GetOrderByID() should fail with empty ID")
	}
}

func TestGetOrderByID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MyOrderDetailResponse{
			Errno:  1002,
			Errmsg: "order not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL, APIKey: "test-key"})

	_, err := client.GetOrderByID(context.Background(), "non-existent-order")
	if err == nil {
		t.Error("Expected error for non-existent order, got nil")
	}
}

func TestGetOrderByID_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MyOrderDetailResponse{
			Errno:  1001,
			Errmsg: "API key invalid",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(&ClientConfig{Host: server.URL, APIKey: "invalid-key"})

	_, err := client.GetOrderByID(context.Background(), "order-123")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
