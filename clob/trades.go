package clob

import (
	"context"
	"fmt"
	"strconv"
)

// GetTrades 获取交易历史（公开市场数据）
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

// GetMyTrades 获取用户的交易历史
// 对应 Python SDK 的 get_my_trades()
// 需要 API Key 认证
func (c *Client) GetMyTrades(ctx context.Context, params *MyTradesParams) ([]*MyTrade, int, error) {
	if params == nil {
		params = &MyTradesParams{
			Page:  1,
			Limit: 10,
		}
	}

	// 设置 chain_id
	if params.ChainID == "" {
		params.ChainID = strconv.Itoa(c.chainID)
	}

	var resp MyTradesResponse
	err := c.httpClient.Get(ctx, "/openapi/trade", params, &resp)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get my trades: %w", err)
	}

	if resp.Errno != 0 {
		return nil, 0, fmt.Errorf("API error (errno=%d): %s", resp.Errno, resp.Errmsg)
	}

	return resp.Result.List, resp.Result.Total, nil
}
