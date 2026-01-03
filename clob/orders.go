package clob

import (
	"context"
	"fmt"
)

// PlaceOrder 下单
func (c *Client) PlaceOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	if c.orderSigner == nil {
		return nil, fmt.Errorf("private key not set")
	}

	// 创建签名订单
	signedOrder, err := c.orderSigner.CreateSignedOrder(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed order: %w", err)
	}

	// 构建请求体
	body := map[string]interface{}{
		"marketId": req.MarketID,
		"order":    signedOrder,
	}

	var resp CreateOrderResponse
	err = c.httpClient.Post(ctx, "/order", body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result, nil
}

// PlaceOrdersBatch 批量下单
func (c *Client) PlaceOrdersBatch(ctx context.Context, reqs []*CreateOrderRequest) ([]*Order, error) {
	if c.orderSigner == nil {
		return nil, fmt.Errorf("private key not set")
	}

	signedOrders := make([]*SignedOrder, len(reqs))
	for i, req := range reqs {
		signedOrder, err := c.orderSigner.CreateSignedOrder(req)
		if err != nil {
			return nil, fmt.Errorf("failed to create signed order %d: %w", i, err)
		}
		signedOrders[i] = signedOrder
	}

	body := map[string]interface{}{
		"orders": signedOrders,
	}

	var resp struct {
		APIResponse
		Result []*Order `json:"result"`
	}
	err := c.httpClient.Post(ctx, "/orders/batch", body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to place orders: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result, nil
}

// GetOrder 获取订单详情
func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	var resp OrderDetailResponse
	err := c.httpClient.Get(ctx, "/order/"+orderID, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result, nil
}

// GetOrders 获取订单列表
func (c *Client) GetOrders(ctx context.Context, params *OrderListParams) ([]*Order, error) {
	if params == nil {
		params = &OrderListParams{
			Page:  1,
			Limit: 20,
		}
	}

	var resp OrderListResponse
	err := c.httpClient.Get(ctx, "/orders", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result.Orders, nil
}

// CancelOrder 取消订单
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}

	body := &CancelOrderRequest{
		OrderID: orderID,
	}

	var resp APIResponse
	err := c.httpClient.Post(ctx, "/order/cancel", body, &resp)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	return nil
}

// CancelOrders 批量取消订单
func (c *Client) CancelOrders(ctx context.Context, orderIDs []string) error {
	if len(orderIDs) == 0 {
		return fmt.Errorf("order IDs are required")
	}

	body := &CancelOrdersRequest{
		OrderIDs: orderIDs,
	}

	var resp APIResponse
	err := c.httpClient.Post(ctx, "/orders/cancel", body, &resp)
	if err != nil {
		return fmt.Errorf("failed to cancel orders: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	return nil
}

// CancelAllOrders 取消全部订单
func (c *Client) CancelAllOrders(ctx context.Context, req *CancelAllOrdersRequest) error {
	var resp APIResponse
	err := c.httpClient.Post(ctx, "/orders/cancel-all", req, &resp)
	if err != nil {
		return fmt.Errorf("failed to cancel all orders: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	return nil
}
