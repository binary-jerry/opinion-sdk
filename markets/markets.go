package markets

import (
	"context"
	"fmt"
)

// GetMarkets 获取市场列表
func (c *Client) GetMarkets(ctx context.Context, params *MarketListParams) (*MarketListResponse, error) {
	if params == nil {
		params = &MarketListParams{
			Page:     1,
			PageSize: 20,
			Status:   MarketStatusActive,
		}
	}

	var resp MarketListResponse
	err := c.httpClient.Get(ctx, "/market", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return &resp, nil
}

// GetMarket 获取单个市场详情
func (c *Client) GetMarket(ctx context.Context, marketID string) (*Market, error) {
	if marketID == "" {
		return nil, fmt.Errorf("market ID is required")
	}

	var resp MarketDetailResponse
	err := c.httpClient.Get(ctx, "/market/"+marketID, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get market: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}

// GetCategoricalMarket 获取多选市场详情
func (c *Client) GetCategoricalMarket(ctx context.Context, marketID string) (*Market, error) {
	if marketID == "" {
		return nil, fmt.Errorf("market ID is required")
	}

	var resp MarketDetailResponse
	err := c.httpClient.Get(ctx, "/market/categorical/"+marketID, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get categorical market: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}

// GetOrderbook 获取订单簿
func (c *Client) GetOrderbook(ctx context.Context, tokenID string) (*Orderbook, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}

	params := struct {
		TokenID string `url:"tokenId"`
	}{
		TokenID: tokenID,
	}

	var resp OrderbookResponse
	err := c.httpClient.Get(ctx, "/token/orderbook", &params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get orderbook: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}

// GetLatestPrice 获取最新价格
func (c *Client) GetLatestPrice(ctx context.Context, tokenID string) (*LatestPrice, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}

	params := struct {
		TokenID string `url:"tokenId"`
	}{
		TokenID: tokenID,
	}

	var resp LatestPriceResponse
	err := c.httpClient.Get(ctx, "/token/latest-price", &params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest price: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}

// GetPriceHistory 获取价格历史
func (c *Client) GetPriceHistory(ctx context.Context, params *PriceHistoryParams) (*PriceHistory, error) {
	if params == nil || params.TokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}

	var resp PriceHistoryResponse
	err := c.httpClient.Get(ctx, "/token/price-history", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}

// GetQuoteTokens 获取报价代币列表
func (c *Client) GetQuoteTokens(ctx context.Context) ([]*QuoteToken, error) {
	var resp QuoteTokenListResponse
	err := c.httpClient.Get(ctx, "/quoteToken", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote tokens: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}

// GetFeeRates 获取费率
func (c *Client) GetFeeRates(ctx context.Context) (*FeeRates, error) {
	var resp FeeRatesResponse
	err := c.httpClient.Get(ctx, "/fee-rates", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get fee rates: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API error: [%d] %s", resp.ErrNo, resp.ErrMsg)
	}

	return resp.Result, nil
}
