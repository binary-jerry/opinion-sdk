package clob

import (
	"context"
	"fmt"
)

// GetBalances 获取余额列表
func (c *Client) GetBalances(ctx context.Context) ([]*Balance, error) {
	var resp BalanceListResponse
	err := c.httpClient.Get(ctx, "/balances", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result, nil
}

// GetPositions 获取持仓列表
func (c *Client) GetPositions(ctx context.Context) ([]*Position, error) {
	var resp PositionListResponse
	err := c.httpClient.Get(ctx, "/positions", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return resp.Result, nil
}

// EnableTrading 启用交易（授权代币）
func (c *Client) EnableTrading(ctx context.Context, quoteTokenID string) (*TxResponse, error) {
	body := &EnableTradingRequest{
		QuoteTokenID: quoteTokenID,
	}

	var resp TxResponse
	err := c.httpClient.Post(ctx, "/enable-trading", body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to enable trading: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return &resp, nil
}

// Split 拆分代币（USDT -> YES + NO）
func (c *Client) Split(ctx context.Context, req *SplitRequest) (*TxResponse, error) {
	var resp TxResponse
	err := c.httpClient.Post(ctx, "/split", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to split: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return &resp, nil
}

// Merge 合并代币（YES + NO -> USDT）
func (c *Client) Merge(ctx context.Context, req *MergeRequest) (*TxResponse, error) {
	var resp TxResponse
	err := c.httpClient.Post(ctx, "/merge", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to merge: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return &resp, nil
}

// Redeem 赎回（从已结算市场获取收益）
func (c *Client) Redeem(ctx context.Context, req *RedeemRequest) (*TxResponse, error) {
	var resp TxResponse
	err := c.httpClient.Post(ctx, "/redeem", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to redeem: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	return &resp, nil
}
