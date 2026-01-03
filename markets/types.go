package markets

import "github.com/shopspring/decimal"

// MarketStatus 市场状态
type MarketStatus string

const (
	MarketStatusActive   MarketStatus = "active"
	MarketStatusResolved MarketStatus = "resolved"
	MarketStatusPaused   MarketStatus = "paused"
)

// MarketType 市场类型
type MarketType string

const (
	MarketTypeBinary      MarketType = "binary"      // 二元市场
	MarketTypeCategorical MarketType = "categorical" // 多选市场
)

// APIResponse API 响应基础结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// Market 市场数据
type Market struct {
	ID              string         `json:"id"`
	Question        string         `json:"question"`
	Description     string         `json:"description"`
	Type            MarketType     `json:"type"`
	Status          MarketStatus   `json:"status"`
	Resolution      string         `json:"resolution,omitempty"`
	CreatedAt       string         `json:"createdAt"`
	EndAt           string         `json:"endAt"`
	ResolvedAt      string         `json:"resolvedAt,omitempty"`
	QuoteTokenID    string         `json:"quoteTokenId"`
	QuoteToken      *QuoteToken    `json:"quoteToken,omitempty"`
	Volume          decimal.Decimal `json:"volume"`
	Liquidity       decimal.Decimal `json:"liquidity"`
	Tokens          []*Token       `json:"tokens,omitempty"`
	ChildMarkets    []*ChildMarket `json:"childMarkets,omitempty"`
	Category        string         `json:"category,omitempty"`
	Tags            []string       `json:"tags,omitempty"`
	ImageURL        string         `json:"imageUrl,omitempty"`
}

// ChildMarket 子市场 (用于多选市场)
type ChildMarket struct {
	ID       string  `json:"id"`
	Question string  `json:"question"`
	Tokens   []*Token `json:"tokens,omitempty"`
}

// Token 代币数据
type Token struct {
	ID        string          `json:"id"`
	TokenID   string          `json:"tokenId"`
	Outcome   string          `json:"outcome"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
	MarketID  string          `json:"marketId"`
}

// QuoteToken 报价代币 (USDT)
type QuoteToken struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

// Orderbook 订单簿
type Orderbook struct {
	TokenID string           `json:"tokenId"`
	Bids    []*OrderbookLevel `json:"bids"`
	Asks    []*OrderbookLevel `json:"asks"`
}

// OrderbookLevel 订单簿层级
type OrderbookLevel struct {
	Price  decimal.Decimal `json:"price"`
	Size   decimal.Decimal `json:"size"`
}

// LatestPrice 最新价格
type LatestPrice struct {
	TokenID   string          `json:"tokenId"`
	Price     decimal.Decimal `json:"price"`
	Timestamp int64           `json:"timestamp"`
}

// PriceHistory 价格历史
type PriceHistory struct {
	TokenID string        `json:"tokenId"`
	Points  []*PricePoint `json:"points"`
}

// PricePoint 价格点
type PricePoint struct {
	Timestamp int64           `json:"timestamp"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    decimal.Decimal `json:"volume"`
}

// FeeRates 费率
type FeeRates struct {
	MakerFeeRate decimal.Decimal `json:"makerFeeRate"`
	TakerFeeRate decimal.Decimal `json:"takerFeeRate"`
}

// MarketListParams 市场列表查询参数
type MarketListParams struct {
	Page     int          `url:"page,omitempty"`
	Limit    int          `url:"limit,omitempty"`
	Status   MarketStatus `url:"status,omitempty"`
	Type     MarketType   `url:"type,omitempty"`
	Category string       `url:"category,omitempty"`
}

// MarketListResponse 市场列表响应
type MarketListResponse struct {
	APIResponse
	Result struct {
		Markets []*Market `json:"markets"`
		Total   int       `json:"total"`
		Page    int       `json:"page"`
		Limit   int       `json:"limit"`
	} `json:"result"`
}

// MarketDetailResponse 市场详情响应
type MarketDetailResponse struct {
	APIResponse
	Result *Market `json:"result"`
}

// OrderbookResponse 订单簿响应
type OrderbookResponse struct {
	APIResponse
	Result *Orderbook `json:"result"`
}

// LatestPriceResponse 最新价格响应
type LatestPriceResponse struct {
	APIResponse
	Result *LatestPrice `json:"result"`
}

// PriceHistoryParams 价格历史查询参数
type PriceHistoryParams struct {
	TokenID  string `url:"tokenId"`
	Interval string `url:"interval,omitempty"` // 1m, 1h, 1d, 1w, max
	StartAt  int64  `url:"startAt,omitempty"`
	EndAt    int64  `url:"endAt,omitempty"`
}

// PriceHistoryResponse 价格历史响应
type PriceHistoryResponse struct {
	APIResponse
	Result *PriceHistory `json:"result"`
}

// QuoteTokenListResponse 报价代币列表响应
type QuoteTokenListResponse struct {
	APIResponse
	Result []*QuoteToken `json:"result"`
}

// FeeRatesResponse 费率响应
type FeeRatesResponse struct {
	APIResponse
	Result *FeeRates `json:"result"`
}
