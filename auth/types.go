package auth

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Wallet 钱包信息
type Wallet struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

// Credentials API 凭证
type Credentials struct {
	APIKey   string `json:"apiKey"`
	Address  string `json:"address"`
}

// AuthHeaders 认证请求头
type AuthHeaders struct {
	APIKey    string
	Address   string
	Timestamp string
	Signature string
}

// ToMap 转换为 map
func (h *AuthHeaders) ToMap() map[string]string {
	m := map[string]string{
		"apikey": h.APIKey,
	}
	if h.Address != "" {
		m["address"] = h.Address
	}
	if h.Timestamp != "" {
		m["timestamp"] = h.Timestamp
	}
	if h.Signature != "" {
		m["signature"] = h.Signature
	}
	return m
}

// TypedDataField EIP-712 类型字段
type TypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TypedDataDomain EIP-712 域
type TypedDataDomain struct {
	Name              string   `json:"name"`
	Version           string   `json:"version"`
	ChainId           *big.Int `json:"chainId"`
	VerifyingContract string   `json:"verifyingContract,omitempty"`
}

// TypedData EIP-712 类型数据
type TypedData struct {
	Types       map[string][]TypedDataField `json:"types"`
	PrimaryType string                      `json:"primaryType"`
	Domain      TypedDataDomain            `json:"domain"`
	Message     map[string]interface{}     `json:"message"`
}

// OrderPayload 订单签名负载
type OrderPayload struct {
	Salt          string `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          int    `json:"side"`          // 0 = BUY, 1 = SELL
	SignatureType int    `json:"signatureType"` // 0 = EOA
}

// EIP-712 类型定义
var (
	// OrderTypes 订单类型定义
	OrderTypes = map[string][]TypedDataField{
		"Order": {
			{Name: "salt", Type: "uint256"},
			{Name: "maker", Type: "address"},
			{Name: "signer", Type: "address"},
			{Name: "taker", Type: "address"},
			{Name: "tokenId", Type: "uint256"},
			{Name: "makerAmount", Type: "uint256"},
			{Name: "takerAmount", Type: "uint256"},
			{Name: "expiration", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "feeRateBps", Type: "uint256"},
			{Name: "side", Type: "uint8"},
			{Name: "signatureType", Type: "uint8"},
		},
	}
)

// OpinionExchangeDomain 返回 Opinion 交易所域
func OpinionExchangeDomain(chainID int, exchangeAddress string) TypedDataDomain {
	return TypedDataDomain{
		Name:              "OPINION CTF Exchange",
		Version:           "1",
		ChainId:           big.NewInt(int64(chainID)),
		VerifyingContract: exchangeAddress,
	}
}
