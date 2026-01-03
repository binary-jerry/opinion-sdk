package clob

import (
	"context"
	"fmt"
)

// GetTrades 获取交易历史
func (c *Client) GetTrades(ctx context.Context, params *TradeListParams) ([]*Trade, error) {
	if params == nil {
		params = &TradeListParams{
			Page:  1,
			Limit: 20,
		}
	}

	var resp TradeListResponse
	err := c.httpClient.Get(ctx, "/trades", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result.Trades, nil
}
