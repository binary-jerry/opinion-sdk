package common

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	ErrUnauthorized     = errors.New("unauthorized")
	ErrNotFound         = errors.New("not found")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrRateLimited      = errors.New("rate limited")
	ErrInternalServer   = errors.New("internal server error")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidPrivateKey = errors.New("invalid private key")
	ErrInvalidAddress   = errors.New("invalid address")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrInvalidPrice     = errors.New("invalid price")
	ErrInvalidOrderType = errors.New("invalid order type")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrOrderNotFound    = errors.New("order not found")
	ErrMarketNotFound   = errors.New("market not found")
	ErrTokenNotFound    = errors.New("token not found")
)

// APIError API 错误
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"msg"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error [%d] %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error [%d] %s", e.StatusCode, e.Code)
}

// NewAPIError 创建 API 错误
func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// IsUnauthorized 检查是否是未授权错误
func IsUnauthorized(err error) bool {
	if errors.Is(err, ErrUnauthorized) {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}
	return false
}

// IsNotFound 检查是否是未找到错误
func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsRateLimited 检查是否是限流错误
func IsRateLimited(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 429
	}
	return false
}
