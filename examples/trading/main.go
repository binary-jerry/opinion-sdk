package main

import (
	"context"
	"fmt"
	"log"

	"github.com/shopspring/decimal"

	opinion "github.com/binary-jerry/opinion-sdk"
	"github.com/binary-jerry/opinion-sdk/clob"
)

func main() {
	// 配置
	config := opinion.DefaultConfig()
	config.APIKey = "your-api-key" // 替换为你的 API Key

	// 私钥（请勿在生产环境硬编码）
	privateKey := "your-private-key" // 替换为你的私钥

	// 创建交易 SDK
	sdk, err := opinion.NewSDK(config, privateKey)
	if err != nil {
		log.Fatalf("创建 SDK 失败: %v", err)
	}

	fmt.Printf("钱包地址: %s\n", sdk.GetAddress())

	ctx := context.Background()

	// 获取余额
	fmt.Println("\n=== 获取余额 ===")
	balances, err := sdk.Trading.GetBalances(ctx)
	if err != nil {
		log.Printf("获取余额失败: %v", err)
	} else {
		for _, b := range balances {
			fmt.Printf("- %s: 可用 %s, 冻结 %s, 总计 %s\n",
				b.Symbol, b.Available, b.Frozen, b.Total)
		}
	}

	// 获取持仓
	fmt.Println("\n=== 获取持仓 ===")
	positions, err := sdk.Trading.GetPositions(ctx)
	if err != nil {
		log.Printf("获取持仓失败: %v", err)
	} else {
		if len(positions) == 0 {
			fmt.Println("暂无持仓")
		}
		for _, p := range positions {
			fmt.Printf("- Market: %s, Token: %s, Size: %s, PnL: %s\n",
				p.MarketID, p.TokenID, p.Size, p.PnL)
		}
	}

	// 获取订单列表
	fmt.Println("\n=== 获取订单列表 ===")
	orders, err := sdk.Trading.GetOrders(ctx, &clob.OrderListParams{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		log.Printf("获取订单列表失败: %v", err)
	} else {
		if len(orders) == 0 {
			fmt.Println("暂无订单")
		}
		for _, o := range orders {
			fmt.Printf("- [%s] %s %s @ %s, 状态: %s\n",
				o.ID, o.Side, o.Size, o.Price, o.Status)
		}
	}

	// 获取交易历史
	fmt.Println("\n=== 获取交易历史 ===")
	trades, err := sdk.Trading.GetTrades(ctx, &clob.TradeListParams{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		log.Printf("获取交易历史失败: %v", err)
	} else {
		if len(trades) == 0 {
			fmt.Println("暂无交易")
		}
		for _, t := range trades {
			fmt.Printf("- [%s] %s %s @ %s, 手续费: %s\n",
				t.ID, t.Side, t.Size, t.Price, t.Fee)
		}
	}

	// 示例：下单（注释掉以避免实际交易）
	/*
	fmt.Println("\n=== 下单示例 ===")
	order, err := sdk.Trading.PlaceOrder(ctx, &clob.CreateOrderRequest{
		MarketID:               "market-id",
		TokenID:                "token-id",
		Side:                   clob.OrderSideBuy,
		Type:                   clob.OrderTypeLimit,
		Price:                  decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(10), // 10 USDT
	})
	if err != nil {
		log.Printf("下单失败: %v", err)
	} else {
		fmt.Printf("订单创建成功: %s\n", order.ID)
	}
	*/

	// 示例：取消订单（注释掉以避免实际操作）
	/*
	fmt.Println("\n=== 取消订单 ===")
	err = sdk.Trading.CancelOrder(ctx, "order-id")
	if err != nil {
		log.Printf("取消订单失败: %v", err)
	} else {
		fmt.Println("订单取消成功")
	}
	*/

	fmt.Println("\n=== 完成 ===")

	// 使用 decimal 避免编译警告
	_ = decimal.Zero
}
