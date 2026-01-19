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
		if r.URL.Path != "/order" {
			t.Errorf("Expected /order, got %s", r.URL.Path)
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
		if r.URL.Path != "/order/cancel" {
			t.Errorf("Expected /order/cancel, got %s", r.URL.Path)
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
