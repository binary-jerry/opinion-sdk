package opinion

import "time"

// 默认配置常量
const (
	// DefaultAPIHost Opinion OpenAPI 主机
	DefaultAPIHost = "https://openapi.opinion.trade/openapi"

	// DefaultChainID BNB Chain 主网
	DefaultChainID = 56

	// DefaultRPCURL BNB Chain 默认 RPC
	DefaultRPCURL = "https://bsc-dataseed.binance.org"

	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries 默认最大重试次数
	DefaultMaxRetries = 3

	// DefaultRetryDelayMs 默认重试延迟（毫秒）
	DefaultRetryDelayMs = 1000

	// RateLimitPerSecond API 每秒请求限制
	RateLimitPerSecond = 15
)

// 合约地址 (BNB Chain Mainnet)
const (
	// ExchangeAddress Opinion 交易所合约地址 (BSC 主网)
	ExchangeAddress = "0x5f45344126d6488025b0b84a3a8189f2487a7246"

	// USDTAddress USDT 合约地址
	USDTAddress = "0x55d398326f99059fF775485246999027B3197955" // BSC USDT
)

// 价格常量
const (
	// MinPrice 最小价格
	MinPrice = 0.01

	// MaxPrice 最大价格
	MaxPrice = 0.99

	// PriceDecimals 价格小数位数
	PriceDecimals = 4

	// Decimal6 USDT 6位小数
	Decimal6 = 1000000

	// Decimal18 标准 18 位小数
	Decimal18 = 1e18
)

// Config SDK 配置
type Config struct {
	// APIHost API 主机地址
	APIHost string

	// APIKey API 密钥
	APIKey string

	// ChainID 链 ID
	ChainID int

	// RPCURL RPC 地址
	RPCURL string

	// PrivateKey 私钥 (可选，用于签名交易)
	PrivateKey string

	// MultiSigAddress 多签地址 (可选)
	MultiSigAddress string

	// Timeout 请求超时
	Timeout time.Duration

	// MaxRetries 最大重试次数
	MaxRetries int

	// RetryDelayMs 重试延迟（毫秒）
	RetryDelayMs int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		APIHost:      DefaultAPIHost,
		ChainID:      DefaultChainID,
		RPCURL:       DefaultRPCURL,
		Timeout:      DefaultTimeout,
		MaxRetries:   DefaultMaxRetries,
		RetryDelayMs: DefaultRetryDelayMs,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.APIHost == "" {
		c.APIHost = DefaultAPIHost
	}
	if c.ChainID == 0 {
		c.ChainID = DefaultChainID
	}
	if c.RPCURL == "" {
		c.RPCURL = DefaultRPCURL
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = DefaultMaxRetries
	}
	if c.RetryDelayMs == 0 {
		c.RetryDelayMs = DefaultRetryDelayMs
	}
	return nil
}

// WithAPIKey 设置 API Key
func (c *Config) WithAPIKey(apiKey string) *Config {
	c.APIKey = apiKey
	return c
}

// WithPrivateKey 设置私钥
func (c *Config) WithPrivateKey(privateKey string) *Config {
	c.PrivateKey = privateKey
	return c
}

// WithMultiSigAddress 设置多签地址
func (c *Config) WithMultiSigAddress(addr string) *Config {
	c.MultiSigAddress = addr
	return c
}

// WithChainID 设置链 ID
func (c *Config) WithChainID(chainID int) *Config {
	c.ChainID = chainID
	return c
}

// WithRPCURL 设置 RPC URL
func (c *Config) WithRPCURL(rpcURL string) *Config {
	c.RPCURL = rpcURL
	return c
}

// WithAPIHost 设置 API 主机
func (c *Config) WithAPIHost(host string) *Config {
	c.APIHost = host
	return c
}
