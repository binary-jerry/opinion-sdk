package opinion

import (
	"fmt"

	"github.com/binary-jerry/opinion-sdk/clob"
	"github.com/binary-jerry/opinion-sdk/markets"
)

// SDK Opinion SDK 统一入口
type SDK struct {
	config  *Config
	Markets *markets.Client
	Trading *clob.Client
}

// NewSDK 创建完整 SDK（需要私钥，支持交易）
// 使用 EOA 模式，maker 和 signer 使用同一地址
func NewSDK(config *Config, privateKey string) (*SDK, error) {
	return NewSDKWithMaker(config, privateKey, "")
}

// NewSDKWithMaker 创建完整 SDK（支持智能钱包）
// privateKey: EOA 私钥，用于签名
// makerAddress: 智能钱包地址（Gnosis Safe），为空时使用 EOA 模式
func NewSDKWithMaker(config *Config, privateKey string, makerAddress string) (*SDK, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if privateKey == "" {
		return nil, fmt.Errorf("private key is required")
	}

	// 创建市场客户端
	marketsClient := markets.NewClient(&markets.ClientConfig{
		Host:         config.APIHost,
		APIKey:       config.APIKey,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	})

	// 创建交易客户端
	tradingClient, err := clob.NewClient(&clob.ClientConfig{
		Host:            config.APIHost,
		APIKey:          config.APIKey,
		PrivateKey:      privateKey,
		ChainID:         config.ChainID,
		ExchangeAddress: ExchangeAddress,
		MakerAddress:    makerAddress, // 智能钱包地址
		Timeout:         config.Timeout,
		MaxRetries:      config.MaxRetries,
		RetryDelayMs:    config.RetryDelayMs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create trading client: %w", err)
	}

	return &SDK{
		config:  config,
		Markets: marketsClient,
		Trading: tradingClient,
	}, nil
}

// NewPublicSDK 创建公开 SDK（无需私钥，仅查询）
func NewPublicSDK(config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}
	_ = config.Validate()

	// 创建市场客户端
	marketsClient := markets.NewClient(&markets.ClientConfig{
		Host:         config.APIHost,
		APIKey:       config.APIKey,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	})

	return &SDK{
		config:  config,
		Markets: marketsClient,
	}
}

// NewTradingSDK 创建交易 SDK
func NewTradingSDK(config *Config, privateKey string) (*SDK, error) {
	return NewSDK(config, privateKey)
}

// GetConfig 获取配置
func (s *SDK) GetConfig() *Config {
	return s.config
}

// GetAddress 获取钱包地址
func (s *SDK) GetAddress() string {
	if s.Trading == nil {
		return ""
	}
	return s.Trading.GetAddress()
}

// SetAPIKey 设置 API Key
func (s *SDK) SetAPIKey(apiKey string) {
	s.config.APIKey = apiKey
	if s.Markets != nil {
		s.Markets.SetAPIKey(apiKey)
	}
	if s.Trading != nil {
		s.Trading.SetAPIKey(apiKey)
	}
}
