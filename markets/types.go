package markets

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// MarketStatus 市场状态
type MarketStatus string

const (
	MarketStatusActive   MarketStatus = "activated" // 活跃
	MarketStatusResolved MarketStatus = "resolved"  // 已结算
	MarketStatusCreated  MarketStatus = "created"   // 已创建
)

// MarketType 市场类型
type MarketType int

const (
	MarketTypeBinary      MarketType = 0 // 二元市场
	MarketTypeCategorical MarketType = 1 // 多选市场
	MarketTypeAll         MarketType = 2 // ALL
)

// APIResponse API 响应基础结构
type APIResponse struct {
	ErrNo  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
}

// IsSuccess 检查 API 是否成功
func (r *APIResponse) IsSuccess() bool {
	return r.ErrNo == 0
}

// Market 市场数据 (匹配 Opinion OpenAPI 响应)
type Market struct {
	MarketID      int64           `json:"marketId"`
	MarketTitle   string          `json:"marketTitle"`
	StatusEnum    string          `json:"statusEnum"`
	MarketType    MarketType      `json:"marketType"`
	ChildMarkets  []*ChildMarket  `json:"childMarkets,omitempty"`
	YesLabel      string          `json:"yesLabel"`
	NoLabel       string          `json:"noLabel"`
	Rules         string          `json:"rules"`
	YesTokenID    string          `json:"yesTokenId"`
	NoTokenID     string          `json:"noTokenId"`
	ConditionID   string          `json:"conditionId"`
	ResultTokenID string          `json:"resultTokenId"`
	Volume        decimal.Decimal `json:"volume"`
	Volume24h     decimal.Decimal `json:"volume24h"`
	Volume7d      decimal.Decimal `json:"volume7d"`
	QuoteToken    string          `json:"quoteToken"`
	ChainID       string          `json:"chainId"`
	QuestionID    string          `json:"questionId"`
	CreatedAt     int64           `json:"createdAt"`
	CutoffAt      int64           `json:"cutoffAt"`
	ResolvedAt    int64           `json:"resolvedAt"`
}

// GetID 返回市场 ID 字符串
func (m *Market) GetID() string {
	return fmt.Sprintf("%d", m.MarketID)
}

// ChildMarket 子市场 (用于多选市场)
type ChildMarket struct {
	MarketID      int64           `json:"marketId"`
	MarketTitle   string          `json:"marketTitle"`
	Status        int             `json:"status"`
	StatusEnum    string          `json:"statusEnum"`
	YesLabel      string          `json:"yesLabel"`
	NoLabel       string          `json:"noLabel"`
	Rules         string          `json:"rules"`
	YesTokenID    string          `json:"yesTokenId"`
	NoTokenID     string          `json:"noTokenId"`
	ConditionID   string          `json:"conditionId"`
	ResultTokenID string          `json:"resultTokenId"`
	Volume        decimal.Decimal `json:"volume"`
	QuoteToken    string          `json:"quoteToken"`
	ChainID       string          `json:"chainId"`
	QuestionID    string          `json:"questionId"`
	CreatedAt     int64           `json:"createdAt"`
	CutoffAt      int64           `json:"cutoffAt"`
	ResolvedAt    int64           `json:"resolvedAt"`
}

// GetID 返回子市场 ID 字符串
func (c *ChildMarket) GetID() string {
	return fmt.Sprintf("%d", c.MarketID)
}

// IsResolved 检查子市场是否已结算
func (c *ChildMarket) IsResolved() bool {
	return c.Status == 4
}

// Token 代币数据
type Token struct {
	ID       string          `json:"id"`
	TokenID  string          `json:"tokenId"`
	Outcome  string          `json:"outcome"`
	Price    decimal.Decimal `json:"price"`
	Volume   decimal.Decimal `json:"volume"`
	MarketID string          `json:"marketId"`
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
	TokenID string            `json:"tokenId"`
	Bids    []*OrderbookLevel `json:"bids"`
	Asks    []*OrderbookLevel `json:"asks"`
}

// OrderbookLevel 订单簿层级
type OrderbookLevel struct {
	Price decimal.Decimal `json:"price"`
	Size  decimal.Decimal `json:"size"`
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
	Page       int          `url:"page,omitempty"`
	PageSize   int          `url:"limit,omitempty"`
	Status     MarketStatus `url:"status,omitempty"`
	MarketType MarketType   `url:"marketType,omitempty"`
	ChainId    int          `url:"chainId,omitempty"`
	SortBy     int          `url:"sortBy,omitempty"`
}

// MarketListResponse 市场列表响应
type MarketListResponse struct {
	APIResponse
	Result struct {
		Total int       `json:"total"`
		List  []*Market `json:"list"`
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
