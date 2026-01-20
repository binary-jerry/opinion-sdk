package orderbook

import (
	"context"

	"github.com/shopspring/decimal"
)

// SnapshotFetchFunc 是一个获取快照的函数类型
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

// APIOrderbook REST API 返回的订单簿结构
// 用于适配 markets.Client.GetOrderbook 的返回值
type APIOrderbook struct {
	Bids []*APIOrderbookLevel
	Asks []*APIOrderbookLevel
}

// APIOrderbookLevel REST API 返回的订单簿层级
type APIOrderbookLevel struct {
	Price decimal.Decimal
	Size  decimal.Decimal
}

// NewSnapshotFetcherFromAPI 从 REST API 创建 SnapshotFetcher
// 使用示例:
//
//	marketsClient := markets.NewClient(config)
//	fetcher := orderbook.NewSnapshotFetcherFromAPI(func(ctx context.Context, tokenID string) (*orderbook.APIOrderbook, error) {
//	    ob, err := marketsClient.GetOrderbook(ctx, tokenID)
//	    if err != nil {
//	        return nil, err
//	    }
//	    return &orderbook.APIOrderbook{
//	        Bids: convertLevels(ob.Bids),
//	        Asks: convertLevels(ob.Asks),
//	    }, nil
//	})
//	sdk.SetSnapshotFetcher(fetcher)
func NewSnapshotFetcherFromAPI(
	getOrderbook func(ctx context.Context, tokenID string) (*APIOrderbook, error),
) SnapshotFetcher {
	return NewFuncSnapshotFetcher(func(ctx context.Context, tokenID string) (bids, asks []PriceLevel, err error) {
		ob, err := getOrderbook(ctx, tokenID)
		if err != nil {
			return nil, nil, err
		}
		if ob == nil {
			return nil, nil, nil
		}

		bids = make([]PriceLevel, 0, len(ob.Bids))
		for _, bid := range ob.Bids {
			if bid != nil {
				bids = append(bids, PriceLevel{Price: bid.Price, Size: bid.Size})
			}
		}

		asks = make([]PriceLevel, 0, len(ob.Asks))
		for _, ask := range ob.Asks {
			if ask != nil {
				asks = append(asks, PriceLevel{Price: ask.Price, Size: ask.Size})
			}
		}

		return bids, asks, nil
	})
}
