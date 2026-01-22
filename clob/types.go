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
	Salt          int64  `json:"salt"`          // 整数类型，与 Python SDK 一致
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
	SignatureType int    `json:"signatureType"` // 数字类型
	Signature     string `json:"signature"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	MarketID                string          `json:"marketId"`
	TokenID                 string          `json:"tokenId"`
	Side                    OrderSide       `json:"side"`
	Type                    OrderType       `json:"type"`
	Price                   decimal.Decimal `json:"price"`
	MakerAmountInQuoteToken decimal.Decimal `json:"makerAmountInQuoteToken,omitempty"` // USDT 金额
	MakerAmountInBaseToken  decimal.Decimal `json:"makerAmountInBaseToken,omitempty"`  // 代币数量
	ExpiresAt               int64           `json:"expiresAt,omitempty"`
	Nonce                   string          `json:"nonce,omitempty"`
	FeeRateBps              int             `json:"feeRateBps,omitempty"`
	CurrencyAddress         string          `json:"currencyAddress,omitempty"` // 报价代币地址 (从市场信息获取)
}

// PlaceOrderResponse 下单响应 (匹配官方 API: OpenapiOrderPost200Response)
type PlaceOrderResponse struct {
	Errno  int             `json:"errno"`  // 0=success
	Errmsg string          `json:"errmsg"` // 错误消息
	Result *PlaceOrderResult `json:"result"`
}

// PlaceOrderResult 下单结果 (匹配官方 API: V2AddOrderResp)
type PlaceOrderResult struct {
	OrderData *PlaceOrderData `json:"orderData"`
}

// PlaceOrderData 下单数据 (匹配官方 API: V2OrderData)
type PlaceOrderData struct {
	OrderID         string `json:"orderId"`
	Amount          string `json:"amount"`          // BUY=USDC金额, SELL=份额
	Filled          string `json:"filled"`          // 已成交数量
	Price           string `json:"price"`
	TotalPrice      string `json:"totalPrice"`
	Status          int    `json:"status"`          // 1=pending, 2=finished, 3=canceled, 4=expired, 5=failed
	Side            int    `json:"side"`            // 1=buy, 2=sell
	OutcomeSide     int    `json:"outcomeSide"`     // 1=yes, 2=no
	Outcome         string `json:"outcome"`
	TradingMethod   int    `json:"tradingMethod"`   // 1=market, 2=limit
	TopicID         int    `json:"topicId"`
	TopicTitle      string `json:"topicTitle"`
	CurrencyAddress string `json:"currencyAddress"`
	CreatedAt       int64  `json:"createdAt"`
	Expiration      int64  `json:"expiration"`
	TransNo         string `json:"transNo"`
}

// CreateOrderResponse 创建订单响应 (旧结构，保留兼容)
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

// ===============================
// 用户交易历史和订单相关类型
// 对应 Python SDK 的 get_my_trades / get_my_orders / get_order_by_id
// ===============================

// MyTradesParams 用户交易历史查询参数
type MyTradesParams struct {
	MarketID int    `url:"market_id,omitempty"`
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	ChainID  string `url:"chain_id,omitempty"`
}

// MyTrade 用户交易记录（Opinion API 格式）
// 字段名称和类型与真实 API 响应一致（2026-01-22 验证）
type MyTrade struct {
	// 交易标识
	OrderNo string `json:"orderNo,omitempty"` // 订单号
	TradeNo string `json:"tradeNo,omitempty"` // 交易号
	TxHash  string `json:"txHash,omitempty"`  // 链上交易哈希

	// 市场信息
	MarketID        int    `json:"marketId,omitempty"`        // 市场 ID
	MarketTitle     string `json:"marketTitle,omitempty"`     // 市场标题
	RootMarketID    int    `json:"rootMarketId,omitempty"`    // 根市场 ID
	RootMarketTitle string `json:"rootMarketTitle,omitempty"` // 根市场标题

	// 交易方向
	Side           string `json:"side,omitempty"`           // "Buy" 或 "Sell" (字符串!)
	Outcome        string `json:"outcome,omitempty"`        // "YES" 或 "NO"
	OutcomeSide    int    `json:"outcomeSide,omitempty"`    // 1=yes, 2=no
	OutcomeSideEnum string `json:"outcomeSideEnum,omitempty"` // "Yes" 或 "No"

	// 价格和数量
	Price              string `json:"price,omitempty"`              // 成交价格
	Shares             string `json:"shares,omitempty"`             // 成交份额
	Amount             string `json:"amount,omitempty"`             // 成交金额 (USDT)
	Fee                int64  `json:"fee,omitempty"`                // 手续费 (原始值，可能为负数)
	FeeFormatted       string `json:"feeFormatted,omitempty"`       // 格式化手续费
	Profit             string `json:"profit,omitempty"`             // 利润
	QuoteToken         string `json:"quoteToken,omitempty"`         // 报价代币地址
	QuoteTokenUsdPrice string `json:"quoteTokenUsdPrice,omitempty"` // 报价代币美元价格
	UsdAmount          string `json:"usdAmount,omitempty"`          // 美元金额

	// 状态
	Status     int    `json:"status,omitempty"`     // 2=finished
	StatusEnum string `json:"statusEnum,omitempty"` // "Finished"

	// 其他
	ChainID   string `json:"chainId,omitempty"`   // 链 ID (字符串)
	CreatedAt int64  `json:"createdAt,omitempty"` // 创建时间戳
}

// MyTradesResponse 用户交易历史响应（列表格式）
type MyTradesResponse struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	Result struct {
		List  []*MyTrade `json:"list"`
		Total int        `json:"total"`
	} `json:"result"`
}

// MyOrdersParams 用户订单列表查询参数
type MyOrdersParams struct {
	MarketID int    `url:"market_id,omitempty"`
	Status   string `url:"status,omitempty"`   // 1=pending, 2=finished, 3=canceled, 4=expired, 5=failed
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	ChainID  string `url:"chain_id,omitempty"`
}

// MyOrder 用户订单（Opinion API 格式）
// 字段名称和类型与真实 API 响应一致（2026-01-22 验证）
type MyOrder struct {
	// 订单标识
	OrderID string `json:"orderId,omitempty"` // 订单 ID
	TransNo string `json:"transNo,omitempty"` // 交易号

	// 状态
	Status     int    `json:"status,omitempty"`     // 1=pending, 2=finished, 3=canceled, 4=expired, 5=failed
	StatusEnum string `json:"statusEnum,omitempty"` // "Finished", "Pending" 等

	// 市场信息
	MarketID        int    `json:"marketId,omitempty"`        // 市场 ID
	MarketTitle     string `json:"marketTitle,omitempty"`     // 市场标题
	RootMarketID    int    `json:"rootMarketId,omitempty"`    // 根市场 ID
	RootMarketTitle string `json:"rootMarketTitle,omitempty"` // 根市场标题

	// 交易方向 (注意: 订单 API 返回数字，交易 API 返回字符串)
	Side              int    `json:"side,omitempty"`              // 1=buy, 2=sell (数字)
	SideEnum          string `json:"sideEnum,omitempty"`          // "Buy" 或 "Sell"
	TradingMethod     int    `json:"tradingMethod,omitempty"`     // 1=market, 2=limit
	TradingMethodEnum string `json:"tradingMethodEnum,omitempty"` // "Market" 或 "Limit"
	Outcome           string `json:"outcome,omitempty"`           // "YES" 或 "NO"
	OutcomeSide       int    `json:"outcomeSide,omitempty"`       // 1=yes, 2=no
	OutcomeSideEnum   string `json:"outcomeSideEnum,omitempty"`   // "Yes" 或 "No"

	// 价格和数量
	Price        string `json:"price,omitempty"`        // 订单价格
	OrderShares  string `json:"orderShares,omitempty"`  // 订单份额
	OrderAmount  string `json:"orderAmount,omitempty"`  // 订单金额 (USDT)
	FilledShares string `json:"filledShares,omitempty"` // 已成交份额
	FilledAmount string `json:"filledAmount,omitempty"` // 已成交金额
	Profit       string `json:"profit,omitempty"`       // 利润
	QuoteToken   string `json:"quoteToken,omitempty"`   // 报价代币地址

	// 时间
	CreatedAt int64 `json:"createdAt,omitempty"` // 创建时间戳
	ExpiresAt int64 `json:"expiresAt,omitempty"` // 过期时间戳

	// 关联交易
	Trades []interface{} `json:"trades,omitempty"` // 关联的交易列表
}

// MyOrdersResponse 用户订单列表响应
type MyOrdersResponse struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	Result struct {
		List  []*MyOrder `json:"list"`
		Total int        `json:"total"`
	} `json:"result"`
}

// MyOrderDetailResponse 用户订单详情响应
// 注意：API 返回的 result 中有 orderData 包装层
type MyOrderDetailResponse struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	Result struct {
		OrderData *MyOrderDetail `json:"orderData"`
	} `json:"result"`
}

// MyOrderDetail 订单详情（包含关联交易记录）
type MyOrderDetail struct {
	// 订单标识
	OrderID string `json:"orderId,omitempty"` // 订单 ID
	TransNo string `json:"transNo,omitempty"` // 交易号

	// 状态
	Status     int    `json:"status,omitempty"`     // 1=pending, 2=finished, 3=canceled, 4=expired, 5=failed
	StatusEnum string `json:"statusEnum,omitempty"` // "Finished", "Pending" 等

	// 市场信息
	MarketID        int    `json:"marketId,omitempty"`        // 市场 ID
	MarketTitle     string `json:"marketTitle,omitempty"`     // 市场标题
	RootMarketID    int    `json:"rootMarketId,omitempty"`    // 根市场 ID
	RootMarketTitle string `json:"rootMarketTitle,omitempty"` // 根市场标题

	// 交易方向
	Side              int    `json:"side,omitempty"`              // 1=buy, 2=sell (数字)
	SideEnum          string `json:"sideEnum,omitempty"`          // "Buy" 或 "Sell"
	TradingMethod     int    `json:"tradingMethod,omitempty"`     // 1=market, 2=limit
	TradingMethodEnum string `json:"tradingMethodEnum,omitempty"` // "Market" 或 "Limit"
	Outcome           string `json:"outcome,omitempty"`           // "YES" 或 "NO"
	OutcomeSide       int    `json:"outcomeSide,omitempty"`       // 1=yes, 2=no
	OutcomeSideEnum   string `json:"outcomeSideEnum,omitempty"`   // "Yes" 或 "No"

	// 价格和数量
	Price        string `json:"price,omitempty"`        // 订单价格
	OrderShares  string `json:"orderShares,omitempty"`  // 订单份额
	OrderAmount  string `json:"orderAmount,omitempty"`  // 订单金额 (USDT)
	FilledShares string `json:"filledShares,omitempty"` // 已成交份额
	FilledAmount string `json:"filledAmount,omitempty"` // 已成交金额
	Profit       string `json:"profit,omitempty"`       // 利润
	QuoteToken   string `json:"quoteToken,omitempty"`   // 报价代币地址

	// 时间
	CreatedAt int64 `json:"createdAt,omitempty"` // 创建时间戳
	ExpiresAt int64 `json:"expiresAt,omitempty"` // 过期时间戳

	// 关联交易记录
	Trades []*MyTrade `json:"trades,omitempty"`
}
