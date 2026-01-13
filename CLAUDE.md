# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Opinion SDK 是一个 Go 语言编写的 Opinion 预测市场 SDK，提供市场数据查询、订单簿管理、交易操作、EIP-712 签名和 WebSocket 实时订阅。

## Build and Run Commands

```bash
# 构建
go build ./...

# 运行所有测试
go test ./...

# 运行单个模块测试
go test ./orderbook/... -v

# 运行单个测试函数
go test ./orderbook/... -v -run TestOrderBook_ApplyDiff

# 运行测试（带覆盖率）
go test ./... -cover

# 格式化和检查
go fmt ./...
go vet ./...

# 运行示例
go run examples/markets/main.go      # 需要 API Key
go run examples/trading/main.go      # 需要私钥
```

**Go 版本要求**: 1.21+

## Architecture

### Core Modules

| 模块 | 说明 |
|------|------|
| `sdk.go` / `config.go` | 统一 SDK 入口和配置 |
| `common/` | 公共模块（HTTP、错误、工具） |
| `markets/` | Markets API（市场/订单簿/价格查询） |
| `auth/` | EIP-712 签名器 |
| `clob/` | 交易模块（订单、账户、签名） |
| `orderbook/` | WebSocket 订单簿模块 |

### SDK Initialization

```go
// 公开 SDK（无需私钥，仅查询）
sdk := opinion.NewPublicSDK(nil)

// 完整 SDK（需要私钥，支持交易）
config := opinion.DefaultConfig()
config.APIKey = "your-api-key"
sdk, err := opinion.NewSDK(config, privateKey)
```

### Orderbook 模块架构

订单簿模块由三层组成：

1. **WSClient** (`ws_client.go`) - WebSocket 连接管理、自动重连、心跳
2. **Manager** (`manager.go`) - 多订单簿管理、增量更新处理、镜像同步
3. **SDK** (`sdk.go`) - 高层 API、快照获取、周期性校验

**镜像同步机制**：二元市场中 YES 和 NO 订单簿互为镜像。当收到一个 token 的更新时，自动计算镜像数据（`mirrorPrice = 1 - price`，`mirrorSide` 反转）更新另一个 token。

```go
// 推荐的订阅方式：订阅整个市场并自动获取快照
sdk.SetSnapshotFetcher(snapshotFetcher)
sdk.Start(ctx)
sdk.SubscribeMarketWithSnapshot(ctx, marketID, yesTokenID, noTokenID)

// 获取订单簿数据（始终从内存获取，不要调用 REST API）
if sdk.IsOrderBookInitialized(tokenID) {
    depth := sdk.GetDepth(tokenID, 50)
    bidsAbove := sdk.ScanBidsAbove(tokenID, minPrice)
    asksBelow := sdk.ScanAsksBelow(tokenID, maxPrice)
}
```

### 数据一致性保障

- **initialized 状态**: 订单簿需要收到快照后才可用
- **pendingDiffs 缓冲**: 快照到达前的增量更新会被缓冲并在快照后重放
- **重连处理**: 断开时清空所有订单簿和 pendingDiffs，重连后自动刷新快照

## Network Configuration

| 配置 | 值 |
|------|------|
| Chain ID | 56 (BNB Chain) |
| API Host | https://openapi.opinion.trade/openapi |
| WebSocket | wss://ws.opinion.trade |
| Rate Limit | 15 requests/second |

### WebSocket Channels

| Channel | 说明 |
|---------|------|
| market.depth.diff | 订单簿增量更新 |
| market.last.price | 最新价格更新 |
| market.last.trade | 最新交易更新 |
| trade.order.update | 用户订单状态更新 |
| trade.record.new | 用户交易记录 |

## Key Constraints

1. **价格范围**: 必须在 0.01 - 0.99 之间，最多 4 位小数
2. **API Key**: 所有请求头中需要 `apikey: your-api-key`
3. **金额精度**: USDT 使用 6 位小数 (`Decimal6 = 1000000`)
4. **EIP-712**: 订单签名使用 EIP-712 类型数据签名
5. **金额类型**: `MakerAmountInQuoteToken`（USDT）或 `MakerAmountInBaseToken`（代币）

## Dependencies

- `github.com/ethereum/go-ethereum`: EIP-712 签名
- `github.com/shopspring/decimal`: 精确十进制运算
- `github.com/gorilla/websocket`: WebSocket 连接

## Smart Contract Operations

- **Split**: 将 USDT 拆分为 YES + NO 代币
- **Merge**: 将 YES + NO 代币合并为 USDT
- **Redeem**: 从已结算市场赎回收益
