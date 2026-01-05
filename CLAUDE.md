# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Opinion SDK 是一个 Go 语言编写的完整 Opinion 预测市场 SDK，提供：
- 市场数据查询（Markets API）
- 订单簿和价格数据
- 交易操作（CLOB API）
- EIP-712 签名
- WebSocket 实时订单簿订阅

## Build and Run Commands

```bash
# 构建
go build ./...

# 运行测试
go test ./...

# 运行测试（带覆盖率）
go test ./... -cover

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...

# 运行示例（需要API Key）
go run examples/markets/main.go

# 运行交易示例（需要私钥）
go run examples/trading/main.go
```

## Architecture

### Module Structure

```
opinion-sdk/
├── sdk.go              # 统一 SDK 入口
├── config.go           # 全局配置和常量
├── common/             # 公共模块
│   ├── errors.go       # 统一错误定义
│   ├── http.go         # HTTP 客户端封装
│   └── utils.go        # 工具函数
├── markets/            # Markets API（市场数据）
│   ├── client.go       # Markets 客户端
│   ├── markets.go      # 市场查询方法
│   └── types.go        # 市场数据类型
├── auth/               # 认证模块
│   ├── types.go        # 认证类型定义
│   └── signer.go       # EIP-712 签名器
├── clob/               # CLOB 交易模块
│   ├── client.go       # CLOB 客户端
│   ├── types.go        # 订单/交易类型
│   ├── signing.go      # 订单签名
│   ├── orders.go       # 订单操作
│   ├── account.go      # 账户查询
│   └── trades.go       # 交易历史
├── orderbook/          # WebSocket 订单簿模块
│   ├── types.go        # 消息类型定义
│   ├── orderbook.go    # 订单簿数据结构
│   ├── ws_client.go    # WebSocket 客户端
│   ├── manager.go      # 订单簿管理器
│   └── sdk.go          # 订单簿 SDK 入口
└── examples/           # 示例代码
```

### SDK Initialization

```go
// 公开 SDK（无需私钥，仅查询）
sdk := opinion.NewPublicSDK(nil)

// 完整 SDK（需要私钥，支持交易）
config := opinion.DefaultConfig()
config.APIKey = "your-api-key"
sdk, err := opinion.NewSDK(config, privateKey)
```

### Core Components

#### 1. 统一入口 (sdk.go)
- `SDK` 结构体整合所有子模块
- `Markets` - 市场数据查询
- `Trading` - 交易操作

#### 2. Markets API (markets/)
- 市场列表查询
- 单个市场详情
- 订单簿查询
- 价格历史
- 报价代币列表

#### 3. Auth 模块 (auth/)
- **Signer**: EIP-712 类型数据签名（钱包签名）
- 订单签名

#### 4. CLOB 模块 (clob/)
- 订单创建/取消
- 批量订单操作
- 余额/持仓查询
- 交易历史
- 拆分/合并/赎回操作

#### 5. Orderbook 模块 (orderbook/)
- WebSocket 实时订阅
- 订单簿维护和更新
- BBO（最佳买卖价）查询
- 深度查询
- 价格和交易更新
- 自动重连和心跳

### Orderbook SDK Usage

```go
import "github.com/binary-jerry/opinion-sdk/orderbook"

// 创建配置
config := orderbook.DefaultConfig()
config.APIKey = "your-api-key"

// 创建订单簿 SDK
sdk := orderbook.NewSDK(config)

// 连接并订阅
ctx := context.Background()
sdk.Start(ctx)
defer sdk.Stop()

// 订阅市场
sdk.Subscribe(1274) // marketID

// 获取 BBO
bbo := sdk.GetBBO("token-id")
fmt.Printf("Best Bid: %s, Best Ask: %s\n", bbo.BestBid.Price, bbo.BestAsk.Price)

// 获取深度
depth := sdk.GetDepth("token-id", 10)

// 监听事件
for event := range sdk.Events() {
    switch event.Type {
    case orderbook.EventDepthUpdate:
        // 处理深度更新
    case orderbook.EventPriceUpdate:
        // 处理价格更新
    case orderbook.EventTradeUpdate:
        // 处理交易更新
    }
}
```

### Authentication

Opinion 使用 API Key 认证：
- 请求头中添加 `apikey: your-api-key`
- 需要钱包签名的操作使用 EIP-712

### Key Design Patterns

1. **API Key 认证**: 所有请求都需要在请求头中包含 apikey
2. **EIP-712 签名**: 订单使用 EIP-712 类型数据签名
3. **价格范围**: 价格必须在 0.01 - 0.99 之间
4. **金额精度**: 使用 6 位小数（USDT）

### Network Configuration

| 配置 | 值 |
|------|------|
| Chain ID | 56 (BNB Chain) |
| API Host | https://proxy.opinion.trade:8443 |
| WebSocket | wss://ws.opinion.trade |
| Rate Limit | 15 requests/second |

### API Endpoints

| API | 端点 |
|-----|------|
| Markets | /market, /market/{id} |
| Orderbook | /token/orderbook |
| Latest Price | /token/latest-price |
| Price History | /token/price-history |
| Quote Tokens | /quoteToken |

## Dependencies

- `github.com/ethereum/go-ethereum`: EIP-712 签名
- `github.com/shopspring/decimal`: 精确十进制运算
- `github.com/gorilla/websocket`: WebSocket 连接

## Testing

运行特定模块测试：
```bash
go test ./auth/... -v
go test ./clob/... -v
go test ./common/... -v
go test ./orderbook/... -v
```

### WebSocket Channels

| Channel | 说明 |
|---------|------|
| market.depth.diff | 订单簿增量更新 |
| market.last.price | 最新价格更新 |
| market.last.trade | 最新交易更新 |
| trade.order.update | 用户订单状态更新 |
| trade.record.new | 用户交易记录 |

## Order Types

- **Market Order**: 市价单，立即以最佳价格成交
- **Limit Order**: 限价单，指定价格成交

## Price Range

- 最小价格: 0.01 (1% 概率)
- 最大价格: 0.99 (99% 概率)
- 精度: 最多 4 位小数

## Common Pitfalls

1. **价格范围**: 价格必须在 0.01 - 0.99 之间
2. **API Key**: 所有请求都需要 API Key
3. **金额类型**: 使用 `MakerAmountInQuoteToken`（USDT）或 `MakerAmountInBaseToken`（代币）
4. **限流**: API 限制 15 次/秒

## Smart Contract Operations

- **Split**: 将 USDT 拆分为 YES + NO 代币
- **Merge**: 将 YES + NO 代币合并为 USDT
- **Redeem**: 从已结算市场赎回收益

## OrderBook 数据一致性保障

### 初始化机制
- **initialized 状态**: 订单簿需要收到快照后才标记为已初始化
- **pendingDiffs 缓冲**: 在快照到达前的增量更新会被缓冲，快照后按序列号重放
- **序列号检查**: 只接受比当前序列号更新的增量更新
- **时间戳追踪**: 记录最后更新时间戳

### 重连处理
- 断开连接时自动清空所有订单簿数据（`clearAllOrderBooks`）
- 重连后重新订阅并等待新快照
- pendingDiffs 同时清空，防止混合新旧数据

### 获取订单簿数据
```go
// 获取完整深度（从内存）
depth := sdk.GetDepth(tokenID, 50)

// 获取所有买单/卖单
allBids := sdk.GetAllBids(tokenID)
allAsks := sdk.GetAllAsks(tokenID)

// 扫描特定价格范围
bidsAbove := sdk.ScanBidsAbove(tokenID, minPrice)
asksBelow := sdk.ScanAsksBelow(tokenID, maxPrice)

// 检查初始化状态
if sdk.IsOrderBookInitialized(tokenID) {
    // 订单簿已就绪
}
```

### 重要提醒
- **始终从 SDK 内存获取订单簿**: 使用 `GetDepth`/`GetAllBids`/`GetAllAsks`，不要调用 REST API
- **检查初始化状态**: 使用前调用 `IsOrderBookInitialized(tokenID)` 确认订单簿已就绪
- **处理深度而非仅 BBO**: 套利计算应考虑完整深度，使用 `ScanAsksBelow`/`ScanBidsAbove`
- **ApplySnapshot 设置初始化**: 必须调用 `ApplySnapshot` 后订单簿才会被标记为已初始化
