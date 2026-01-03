package clob

import "github.com/shopspring/decimal"

// OrderSide 订单方向
type OrderSide int

const (
	OrderSideBuy  OrderSide = 0
	OrderSideSell OrderSide = 1
)

// String 返回订单方向字符串
func (s OrderSide) String() string {
	switch s {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// OrderType 订单类型
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusExpired   OrderStatus = "expired"
)

// APIResponse API 响应基础结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// Order 订单
type Order struct {
	ID              string          `json:"id"`
	MarketID        string          `json:"marketId"`
	TokenID         string          `json:"tokenId"`
	Side            string          `json:"side"`
	Type            OrderType       `json:"type"`
	Price           decimal.Decimal `json:"price"`
	Size            decimal.Decimal `json:"size"`
	FilledSize      decimal.Decimal `json:"filledSize"`
	RemainingSize   decimal.Decimal `json:"remainingSize"`
	Status          OrderStatus     `json:"status"`
	CreatedAt       string          `json:"createdAt"`
	UpdatedAt       string          `json:"updatedAt"`
	ExpiresAt       string          `json:"expiresAt,omitempty"`
	Maker           string          `json:"maker"`
	Signature       string          `json:"signature"`
}

// SignedOrder 已签名订单
type SignedOrder struct {
	Salt          string `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenId       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          string `json:"side"`
	SignatureType string `json:"signatureType"`
	Signature     string `json:"signature"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	MarketID               string          `json:"marketId"`
	TokenID                string          `json:"tokenId"`
	Side                   OrderSide       `json:"side"`
	Type                   OrderType       `json:"type"`
	Price                  decimal.Decimal `json:"price"`
	MakerAmountInQuoteToken decimal.Decimal `json:"makerAmountInQuoteToken,omitempty"` // USDT 金额
	MakerAmountInBaseToken  decimal.Decimal `json:"makerAmountInBaseToken,omitempty"`  // 代币数量
	ExpiresAt              int64           `json:"expiresAt,omitempty"`
	Nonce                  string          `json:"nonce,omitempty"`
	FeeRateBps             int             `json:"feeRateBps,omitempty"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	APIResponse
	Result *Order `json:"result"`
}

// OrderListParams 订单列表查询参数
type OrderListParams struct {
	MarketID string      `url:"marketId,omitempty"`
	TokenID  string      `url:"tokenId,omitempty"`
	Status   OrderStatus `url:"status,omitempty"`
	Side     string      `url:"side,omitempty"`
	Page     int         `url:"page,omitempty"`
	Limit    int         `url:"limit,omitempty"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	APIResponse
	Result struct {
		Orders []*Order `json:"orders"`
		Total  int      `json:"total"`
	} `json:"result"`
}

// OrderDetailResponse 订单详情响应
type OrderDetailResponse struct {
	APIResponse
	Result *Order `json:"result"`
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	OrderID string `json:"orderId"`
}

// CancelOrdersRequest 批量取消订单请求
type CancelOrdersRequest struct {
	OrderIDs []string `json:"orderIds"`
}

// CancelAllOrdersRequest 取消全部订单请求
type CancelAllOrdersRequest struct {
	MarketID string `json:"marketId,omitempty"`
	Side     string `json:"side,omitempty"`
}

// Trade 交易记录
type Trade struct {
	ID        string          `json:"id"`
	OrderID   string          `json:"orderId"`
	MarketID  string          `json:"marketId"`
	TokenID   string          `json:"tokenId"`
	Side      string          `json:"side"`
	Price     decimal.Decimal `json:"price"`
	Size      decimal.Decimal `json:"size"`
	Fee       decimal.Decimal `json:"fee"`
	Timestamp string          `json:"timestamp"`
	TxHash    string          `json:"txHash,omitempty"`
}

// TradeListParams 交易历史查询参数
type TradeListParams struct {
	MarketID string `url:"marketId,omitempty"`
	TokenID  string `url:"tokenId,omitempty"`
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
}

// TradeListResponse 交易历史响应
type TradeListResponse struct {
	APIResponse
	Result struct {
		Trades []*Trade `json:"trades"`
		Total  int      `json:"total"`
	} `json:"result"`
}

// Balance 余额
type Balance struct {
	TokenID   string          `json:"tokenId"`
	Symbol    string          `json:"symbol"`
	Available decimal.Decimal `json:"available"`
	Frozen    decimal.Decimal `json:"frozen"`
	Total     decimal.Decimal `json:"total"`
}

// BalanceListResponse 余额列表响应
type BalanceListResponse struct {
	APIResponse
	Result []*Balance `json:"result"`
}

// Position 持仓
type Position struct {
	MarketID     string          `json:"marketId"`
	TokenID      string          `json:"tokenId"`
	Outcome      string          `json:"outcome"`
	Size         decimal.Decimal `json:"size"`
	AvgPrice     decimal.Decimal `json:"avgPrice"`
	CurrentPrice decimal.Decimal `json:"currentPrice"`
	PnL          decimal.Decimal `json:"pnl"`
	PnLPercent   decimal.Decimal `json:"pnlPercent"`
}

// PositionListResponse 持仓列表响应
type PositionListResponse struct {
	APIResponse
	Result []*Position `json:"result"`
}

// SplitRequest 拆分请求
type SplitRequest struct {
	MarketID string          `json:"marketId"`
	Amount   decimal.Decimal `json:"amount"` // USDT 金额
}

// MergeRequest 合并请求
type MergeRequest struct {
	MarketID string          `json:"marketId"`
	Amount   decimal.Decimal `json:"amount"` // 代币数量
}

// RedeemRequest 赎回请求
type RedeemRequest struct {
	MarketID string `json:"marketId"`
}

// TxResponse 交易响应
type TxResponse struct {
	APIResponse
	Result struct {
		TxHash  string `json:"txHash"`
		Success bool   `json:"success"`
	} `json:"result"`
}

// EnableTradingRequest 启用交易请求
type EnableTradingRequest struct {
	QuoteTokenID string `json:"quoteTokenId"`
}
