package clob

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// PlaceOrder 下单
// 返回 PlaceOrderData，包含 OrderID、Filled（成交数量）、Status 等信息
func (c *Client) PlaceOrder(ctx context.Context, req *CreateOrderRequest) (*PlaceOrderData, error) {
	if c.orderSigner == nil {
		return nil, fmt.Errorf("private key not set")
	}

	// 创建签名订单
	signedOrder, err := c.orderSigner.CreateSignedOrder(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed order: %w", err)
	}

	// 构建请求体 - 使用 V2AddOrderReq 扁平结构（与 Python SDK 一致）
	marketIDInt, _ := strconv.ParseInt(req.MarketID, 10, 64)

	// side: 0=BUY, 1=SELL（数字字符串）
	side := "0"
	if req.Side == OrderSideSell {
		side = "1"
	}

	// tradingMethod: 1=市价, 2=限价
	tradingMethod := 2
	if req.Type == OrderTypeMarket {
		tradingMethod = 1
	}

	body := map[string]interface{}{
		"topicId":         marketIDInt,
		"salt":            strconv.FormatInt(signedOrder.Salt, 10), // 字符串
		"maker":           signedOrder.Maker,
		"signer":          signedOrder.Signer,
		"taker":           signedOrder.Taker,
		"tokenId":         signedOrder.TokenId,
		"makerAmount":     signedOrder.MakerAmount,
		"takerAmount":     signedOrder.TakerAmount,
		"expiration":      signedOrder.Expiration,
		"nonce":           signedOrder.Nonce,
		"feeRateBps":      signedOrder.FeeRateBps,
		"side":            side,                                     // 数字字符串
		"signatureType":   strconv.Itoa(signedOrder.SignatureType), // 字符串
		"signature":       signedOrder.Signature,
		"sign":            signedOrder.Signature, // 同 signature
		"contractAddress": "",
		"currencyAddress": req.CurrencyAddress, // 报价代币地址 (从市场信息获取)
		"price":           req.Price.String(),                      // 价格字符串
		"tradingMethod":   tradingMethod,                           // 整数
		"timestamp":       int(c.timestamp()),                      // 当前时间戳
		"safeRate":        "0",
		"orderExpTime":    "0",
	}

	var resp PlaceOrderResponse
	err = c.httpClient.Post(ctx, "/openapi/order", body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("API error (errno=%d): %s", resp.Errno, resp.Errmsg)
	}

	if resp.Result == nil || resp.Result.OrderData == nil {
		return nil, fmt.Errorf("invalid response: missing order data")
	}

	return resp.Result.OrderData, nil
}

// timestamp 获取当前时间戳（秒）
func (c *Client) timestamp() int64 {
	return time.Now().Unix()
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

// GetOrder 获取订单详情（旧方法，保留兼容）
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

// GetOrders 获取订单列表（旧方法，保留兼容）
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

// GetMyOrders 获取用户的订单列表
// 对应 Python SDK 的 get_my_orders()
// 需要 API Key 认证
func (c *Client) GetMyOrders(ctx context.Context, params *MyOrdersParams) ([]*MyOrder, int, error) {
	if params == nil {
		params = &MyOrdersParams{
			Page:  1,
			Limit: 10,
		}
	}

	// 设置 chain_id
	if params.ChainID == "" {
		params.ChainID = strconv.Itoa(c.chainID)
	}

	var resp MyOrdersResponse
	err := c.httpClient.Get(ctx, "/openapi/order", params, &resp)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get my orders: %w", err)
	}

	if resp.Errno != 0 {
		return nil, 0, fmt.Errorf("API error (errno=%d): %s", resp.Errno, resp.Errmsg)
	}

	return resp.Result.List, resp.Result.Total, nil
}

// GetOrderByID 根据订单 ID 获取订单详情（包含关联交易记录）
// 对应 Python SDK 的 get_order_by_id()
// 需要 API Key 认证
func (c *Client) GetOrderByID(ctx context.Context, orderID string) (*MyOrderDetail, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	var resp MyOrderDetailResponse
	err := c.httpClient.Get(ctx, "/openapi/order/"+orderID, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("API error (errno=%d): %s", resp.Errno, resp.Errmsg)
	}

	return resp.Result.OrderData, nil
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
	err := c.httpClient.Post(ctx, "/openapi/order/cancel", body, &resp)
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

// PreSignedOrder 预签名订单（包含签名后的订单和提交请求体）
type PreSignedOrder struct {
	SignedOrder *SignedOrder // 已签名的订单
	RequestBody map[string]interface{} // 提交请求体
	Request     *CreateOrderRequest // 原始请求（用于参考）
}

// CreatePreSignedOrder 创建预签名订单（不提交）
// 返回预签名订单，可以在之后快速提交
func (c *Client) CreatePreSignedOrder(req *CreateOrderRequest) (*PreSignedOrder, error) {
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

	return &PreSignedOrder{
		SignedOrder: signedOrder,
		RequestBody: body,
		Request:     req,
	}, nil
}

// SubmitPreSignedOrder 提交预签名订单
// 使用之前创建的预签名订单快速提交，节省签名时间
func (c *Client) SubmitPreSignedOrder(ctx context.Context, preSignedOrder *PreSignedOrder) (*PlaceOrderData, error) {
	if preSignedOrder == nil || preSignedOrder.RequestBody == nil {
		return nil, fmt.Errorf("invalid pre-signed order")
	}

	var resp PlaceOrderResponse
	err := c.httpClient.Post(ctx, "/openapi/order", preSignedOrder.RequestBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to submit pre-signed order: %w", err)
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("API error (errno=%d): %s", resp.Errno, resp.Errmsg)
	}

	if resp.Result == nil || resp.Result.OrderData == nil {
		return nil, fmt.Errorf("invalid response: missing order data")
	}

	return resp.Result.OrderData, nil
}

// CreatePreSignedOrders 批量创建预签名订单（不提交）
func (c *Client) CreatePreSignedOrders(reqs []*CreateOrderRequest) ([]*PreSignedOrder, error) {
	if c.orderSigner == nil {
		return nil, fmt.Errorf("private key not set")
	}

	if len(reqs) == 0 {
		return nil, nil
	}

	preSignedOrders := make([]*PreSignedOrder, 0, len(reqs))
	for _, req := range reqs {
		preSignedOrder, err := c.CreatePreSignedOrder(req)
		if err != nil {
			return nil, fmt.Errorf("failed to create pre-signed order: %w", err)
		}
		preSignedOrders = append(preSignedOrders, preSignedOrder)
	}

	return preSignedOrders, nil
}

// SubmitPreSignedOrders 批量提交预签名订单
func (c *Client) SubmitPreSignedOrders(ctx context.Context, preSignedOrders []*PreSignedOrder) ([]*Order, error) {
	if len(preSignedOrders) == 0 {
		return nil, nil
	}

	// 提取已签名订单
	signedOrders := make([]*SignedOrder, 0, len(preSignedOrders))
	for _, preSignedOrder := range preSignedOrders {
		if preSignedOrder == nil || preSignedOrder.SignedOrder == nil {
			return nil, fmt.Errorf("invalid pre-signed order in batch")
		}
		signedOrders = append(signedOrders, preSignedOrder.SignedOrder)
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
		return nil, fmt.Errorf("failed to submit pre-signed orders: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result, nil
}
