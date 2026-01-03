package clob

import (
	"time"

	"github.com/binary-jerry/opinion-sdk/auth"
	"github.com/binary-jerry/opinion-sdk/common"
)

// Client CLOB 交易客户端
type Client struct {
	httpClient      *common.HTTPClient
	signer          *auth.Signer
	orderSigner     *OrderSigner
	apiKey          string
	chainID         int
	exchangeAddress string
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Host            string
	APIKey          string
	PrivateKey      string
	ChainID         int
	ExchangeAddress string
	Timeout         time.Duration
	MaxRetries      int
	RetryDelayMs    int
}

// NewClient 创建 CLOB 客户端
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		config = &ClientConfig{
			Host:         "https://proxy.opinion.trade:8443",
			ChainID:      56,
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			RetryDelayMs: 1000,
		}
	}

	httpClient := common.NewHTTPClient(&common.HTTPClientConfig{
		BaseURL:      config.Host,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	})

	// 设置 API Key 请求头
	if config.APIKey != "" {
		httpClient.SetDefaultHeader("apikey", config.APIKey)
	}

	client := &Client{
		httpClient:      httpClient,
		apiKey:          config.APIKey,
		chainID:         config.ChainID,
		exchangeAddress: config.ExchangeAddress,
	}

	// 如果提供了私钥，创建签名器
	if config.PrivateKey != "" {
		signer, err := auth.NewSigner(config.PrivateKey, config.ChainID)
		if err != nil {
			return nil, err
		}
		client.signer = signer
		client.orderSigner = NewOrderSigner(signer, config.ChainID, config.ExchangeAddress)
		httpClient.SetDefaultHeader("address", signer.GetAddress())
	}

	return client, nil
}

// SetAPIKey 设置 API Key
func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
	c.httpClient.SetDefaultHeader("apikey", apiKey)
}

// SetPrivateKey 设置私钥
func (c *Client) SetPrivateKey(privateKey string) error {
	signer, err := auth.NewSigner(privateKey, c.chainID)
	if err != nil {
		return err
	}
	c.signer = signer
	c.orderSigner = NewOrderSigner(signer, c.chainID, c.exchangeAddress)
	c.httpClient.SetDefaultHeader("address", signer.GetAddress())
	return nil
}

// GetAddress 获取钱包地址
func (c *Client) GetAddress() string {
	if c.signer == nil {
		return ""
	}
	return c.signer.GetAddress()
}

// GetOrderSigner 获取订单签名器
func (c *Client) GetOrderSigner() *OrderSigner {
	return c.orderSigner
}
