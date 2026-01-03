package main

import (
	"context"
	"fmt"
	"log"

	opinion "github.com/binary-jerry/opinion-sdk"
	"github.com/binary-jerry/opinion-sdk/markets"
)

func main() {
	// 创建公开 SDK（无需私钥）
	config := opinion.DefaultConfig()
	config.APIKey = "your-api-key" // 替换为你的 API Key

	sdk := opinion.NewPublicSDK(config)

	ctx := context.Background()

	// 获取市场列表
	fmt.Println("=== 获取市场列表 ===")
	marketList, err := sdk.Markets.GetMarkets(ctx, &markets.MarketListParams{
		Page:   1,
		Limit:  10,
		Status: markets.MarketStatusActive,
	})
	if err != nil {
		log.Printf("获取市场列表失败: %v", err)
	} else {
		fmt.Printf("共找到 %d 个市场\n", marketList.Result.Total)
		for _, m := range marketList.Result.Markets {
			fmt.Printf("- [%s] %s\n", m.ID, m.Question)
		}
	}

	// 获取报价代币
	fmt.Println("\n=== 获取报价代币 ===")
	quoteTokens, err := sdk.Markets.GetQuoteTokens(ctx)
	if err != nil {
		log.Printf("获取报价代币失败: %v", err)
	} else {
		for _, qt := range quoteTokens {
			fmt.Printf("- %s (%s): %s\n", qt.Symbol, qt.Name, qt.Address)
		}
	}

	// 如果有市场，获取第一个市场的订单簿
	if marketList != nil && len(marketList.Result.Markets) > 0 {
		market := marketList.Result.Markets[0]
		if len(market.Tokens) > 0 {
			tokenID := market.Tokens[0].TokenID

			fmt.Println("\n=== 获取订单簿 ===")
			orderbook, err := sdk.Markets.GetOrderbook(ctx, tokenID)
			if err != nil {
				log.Printf("获取订单簿失败: %v", err)
			} else {
				fmt.Printf("Token ID: %s\n", orderbook.TokenID)
				fmt.Printf("Bids: %d 层\n", len(orderbook.Bids))
				fmt.Printf("Asks: %d 层\n", len(orderbook.Asks))

				if len(orderbook.Bids) > 0 {
					fmt.Printf("最高买价: %s @ %s\n", orderbook.Bids[0].Price, orderbook.Bids[0].Size)
				}
				if len(orderbook.Asks) > 0 {
					fmt.Printf("最低卖价: %s @ %s\n", orderbook.Asks[0].Price, orderbook.Asks[0].Size)
				}
			}

			// 获取最新价格
			fmt.Println("\n=== 获取最新价格 ===")
			price, err := sdk.Markets.GetLatestPrice(ctx, tokenID)
			if err != nil {
				log.Printf("获取最新价格失败: %v", err)
			} else {
				fmt.Printf("Token ID: %s, 价格: %s\n", price.TokenID, price.Price)
			}
		}
	}

	fmt.Println("\n=== 完成 ===")
}
