package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// TimestampMs 获取当前时间戳（毫秒）
func TimestampMs() int64 {
	return time.Now().UnixMilli()
}

// TimestampSec 获取当前时间戳（秒）
func TimestampSec() int64 {
	return time.Now().Unix()
}

// TimestampSecStr 获取当前时间戳字符串（秒）
func TimestampSecStr() string {
	return strconv.FormatInt(TimestampSec(), 10)
}

// TimestampMsStr 获取当前时间戳字符串（毫秒）
func TimestampMsStr() string {
	return strconv.FormatInt(TimestampMs(), 10)
}

// GenerateRandomSalt 生成随机 salt
func GenerateRandomSalt() (*big.Int, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	return new(big.Int).SetBytes(bytes), nil
}

// GenerateRandomNonce 生成随机 nonce
func GenerateRandomNonce() (int64, error) {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random nonce: %w", err)
	}
	nonce := new(big.Int).SetBytes(bytes)
	return nonce.Int64() & 0x7FFFFFFFFFFFFFFF, nil // 确保为正数
}

// GenerateRandomHex 生成随机十六进制字符串
func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random hex: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}

// IntPtr 返回 int 指针
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr 返回 int64 指针
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr 返回 bool 指针
func BoolPtr(b bool) *bool {
	return &b
}

// ParseBigInt 解析大整数
func ParseBigInt(s string) (*big.Int, bool) {
	n := new(big.Int)
	_, ok := n.SetString(s, 10)
	return n, ok
}

// FormatBigInt 格式化大整数
func FormatBigInt(n *big.Int) string {
	if n == nil {
		return "0"
	}
	return n.String()
}

// Min 返回两个 int64 中的较小值
func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Max 返回两个 int64 中的较大值
func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
