package orderbook

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// SnapshotFetchFunc 是一个获取快照的函数类型
// 可以直接传入 markets.Client.GetOrderbook 方法
type SnapshotFetchFunc func(ctx context.Context, tokenID string) (bids, asks []PriceLevel, err error)

// FuncSnapshotFetcher 将函数包装为 SnapshotFetcher
type FuncSnapshotFetcher struct {
	fetchFunc SnapshotFetchFunc
}

// NewFuncSnapshotFetcher 从函数创建 SnapshotFetcher
func NewFuncSnapshotFetcher(fn SnapshotFetchFunc) *FuncSnapshotFetcher {
	return &FuncSnapshotFetcher{fetchFunc: fn}
}

// GetOrderbookSnapshot 实现 SnapshotFetcher 接口
func (f *FuncSnapshotFetcher) GetOrderbookSnapshot(ctx context.Context, tokenID string) (bids, asks []PriceLevel, err error) {
	return f.fetchFunc(ctx, tokenID)
}

// OrderbookAPIData 订单簿数据（从 REST API 返回）
// 用于适配 markets.Client.GetOrderbook 的返回值
type OrderbookAPIData struct {
	TokenID string
	Bids    []*OrderbookAPILevel
	Asks    []*OrderbookAPILevel
}

// OrderbookAPILevel 订单簿层级（从 REST API 返回，使用 decimal.Decimal）
type OrderbookAPILevel struct {
	Price decimal.Decimal
	Size  decimal.Decimal
}

// CreateSnapshotFetcherFromMarketsClient 从 markets.Client 类型的方法创建 SnapshotFetcher
// 使用示例:
//
//	marketsClient := markets.NewClient(config)
//	fetcher := orderbook.CreateSnapshotFetcherFromMarketsClient(func(ctx context.Context, tokenID string) (*markets.Orderbook, error) {
//	    return marketsClient.GetOrderbook(ctx, tokenID)
//	})
//	orderbookSDK.SetSnapshotFetcher(fetcher)
func CreateSnapshotFetcherFromMarketsClient(
	getOrderbook func(ctx context.Context, tokenID string) (*OrderbookAPIData, error),
) SnapshotFetcher {
	return NewFuncSnapshotFetcher(func(ctx context.Context, tokenID string) (bids, asks []PriceLevel, err error) {
		ob, err := getOrderbook(ctx, tokenID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get orderbook: %w", err)
		}

		if ob == nil {
			return nil, nil, fmt.Errorf("orderbook is nil")
		}

		// 转换 bids
		bids = make([]PriceLevel, 0, len(ob.Bids))
		for _, bid := range ob.Bids {
			if bid != nil {
				bids = append(bids, PriceLevel{Price: bid.Price, Size: bid.Size})
			}
		}

		// 转换 asks
		asks = make([]PriceLevel, 0, len(ob.Asks))
		for _, ask := range ob.Asks {
			if ask != nil {
				asks = append(asks, PriceLevel{Price: ask.Price, Size: ask.Size})
			}
		}

		return bids, asks, nil
	})
}
