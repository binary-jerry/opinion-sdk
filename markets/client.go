package markets

import (
	"time"

	"github.com/binary-jerry/opinion-sdk/common"
)

// Client 市场数据客户端
type Client struct {
	httpClient *common.HTTPClient
	apiKey     string
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Host         string
	APIKey       string
	Timeout      time.Duration
	MaxRetries   int
	RetryDelayMs int
}

// NewClient 创建市场数据客户端
func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = &ClientConfig{
			Host:         "https://proxy.opinion.trade:8443",
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

	return &Client{
		httpClient: httpClient,
		apiKey:     config.APIKey,
	}
}

// SetAPIKey 设置 API Key
func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
	c.httpClient.SetDefaultHeader("apikey", apiKey)
}
