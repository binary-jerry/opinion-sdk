package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// HTTPClient HTTP 客户端封装
type HTTPClient struct {
	client         *http.Client
	baseURL        string
	maxRetries     int
	retryDelay     time.Duration
	defaultHeaders map[string]string
}

// HTTPClientConfig HTTP 客户端配置
type HTTPClientConfig struct {
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	RetryDelayMs int
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(config *HTTPClientConfig) *HTTPClient {
	if config == nil {
		config = &HTTPClientConfig{
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			RetryDelayMs: 1000,
		}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:        strings.TrimSuffix(config.BaseURL, "/"),
		maxRetries:     config.MaxRetries,
		retryDelay:     time.Duration(config.RetryDelayMs) * time.Millisecond,
		defaultHeaders: make(map[string]string),
	}
}

// SetDefaultHeader 设置默认请求头
func (c *HTTPClient) SetDefaultHeader(key, value string) {
	c.defaultHeaders[key] = value
}

// Get 发送 GET 请求
func (c *HTTPClient) Get(ctx context.Context, path string, params interface{}, result interface{}) error {
	fullURL := c.buildURL(path, params)
	return c.doRequest(ctx, http.MethodGet, fullURL, nil, nil, result)
}

// Post 发送 POST 请求
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	fullURL := c.buildURL(path, nil)
	return c.doRequest(ctx, http.MethodPost, fullURL, body, nil, result)
}

// Delete 发送 DELETE 请求
func (c *HTTPClient) Delete(ctx context.Context, path string, result interface{}) error {
	fullURL := c.buildURL(path, nil)
	return c.doRequest(ctx, http.MethodDelete, fullURL, nil, nil, result)
}

// DeleteWithBody 发送带 body 的 DELETE 请求
func (c *HTTPClient) DeleteWithBody(ctx context.Context, path string, body interface{}, result interface{}) error {
	fullURL := c.buildURL(path, nil)
	return c.doRequest(ctx, http.MethodDelete, fullURL, body, nil, result)
}

// DoWithAuth 发送带认证的请求
func (c *HTTPClient) DoWithAuth(ctx context.Context, method, path string, body interface{}, authHeaders map[string]string, result interface{}) error {
	fullURL := c.buildURL(path, nil)
	return c.doRequest(ctx, method, fullURL, body, authHeaders, result)
}

// DoWithAuthAndParams 发送带认证和查询参数的请求
func (c *HTTPClient) DoWithAuthAndParams(ctx context.Context, method, path string, params interface{}, body interface{}, authHeaders map[string]string, result interface{}) error {
	fullURL := c.buildURL(path, params)
	return c.doRequest(ctx, method, fullURL, body, authHeaders, result)
}

// buildURL 构建完整 URL
func (c *HTTPClient) buildURL(path string, params interface{}) string {
	fullURL := c.baseURL + path

	if params == nil {
		return fullURL
	}

	queryString := structToQueryString(params)
	if queryString != "" {
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + queryString
		} else {
			fullURL += "?" + queryString
		}
	}

	return fullURL
}

// doRequest 执行 HTTP 请求
func (c *HTTPClient) doRequest(ctx context.Context, method, fullURL string, body interface{}, extraHeaders map[string]string, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		err := c.doSingleRequest(ctx, method, fullURL, body, extraHeaders, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// 不重试的错误类型
		if IsUnauthorized(err) || IsNotFound(err) {
			return err
		}

		// 4xx 错误（除了 429）不重试
		if apiErr, ok := err.(*APIError); ok {
			if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
				return err
			}
		}
	}

	return lastErr
}

// doSingleRequest 执行单次 HTTP 请求
func (c *HTTPClient) doSingleRequest(ctx context.Context, method, fullURL string, body interface{}, extraHeaders map[string]string, result interface{}) error {
	var bodyReader io.Reader
	var bodyBytes []byte

	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置默认头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	// 设置额外头（如认证头）
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 处理错误响应
	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if len(respBody) > 0 {
			_ = json.Unmarshal(respBody, apiErr)
		}
		if apiErr.Code == "" {
			apiErr.Code = http.StatusText(resp.StatusCode)
		}
		return apiErr
	}

	// 解析响应
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// structToQueryString 将结构体转换为查询字符串
func structToQueryString(params interface{}) string {
	if params == nil {
		return ""
	}

	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	values := url.Values{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 获取 url 标签
		tag := fieldType.Tag.Get("url")
		if tag == "" || tag == "-" {
			continue
		}

		// 解析标签
		parts := strings.Split(tag, ",")
		key := parts[0]
		omitempty := len(parts) > 1 && parts[1] == "omitempty"

		// 获取字段值
		var strValue string
		switch field.Kind() {
		case reflect.String:
			strValue = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 || !omitempty {
				strValue = fmt.Sprintf("%d", field.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 || !omitempty {
				strValue = fmt.Sprintf("%d", field.Uint())
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() != 0 || !omitempty {
				strValue = fmt.Sprintf("%f", field.Float())
			}
		case reflect.Bool:
			strValue = fmt.Sprintf("%t", field.Bool())
		case reflect.Ptr:
			if !field.IsNil() {
				elem := field.Elem()
				switch elem.Kind() {
				case reflect.Bool:
					strValue = fmt.Sprintf("%t", elem.Bool())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					strValue = fmt.Sprintf("%d", elem.Int())
				case reflect.String:
					strValue = elem.String()
				}
			}
		}

		if strValue != "" || !omitempty {
			if strValue != "" {
				values.Set(key, strValue)
			}
		}
	}

	return values.Encode()
}

// GetBaseURL 获取基础 URL
func (c *HTTPClient) GetBaseURL() string {
	return c.baseURL
}
