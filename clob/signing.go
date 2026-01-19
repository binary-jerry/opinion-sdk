package clob

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/binary-jerry/opinion-sdk/auth"
	"github.com/binary-jerry/opinion-sdk/common"
)

// Decimal18 USDT 精度 (BSC 上的 USDT 使用 18 位小数)
var Decimal18 = decimal.RequireFromString("1000000000000000000")

// SignatureType 签名类型
const (
	SignatureTypeEOA        = 0 // EOA 钱包签名
	SignatureTypeGnosisSafe = 2 // Gnosis Safe 智能钱包签名
)

// OrderSigner 订单签名器
type OrderSigner struct {
	signer          *auth.Signer
	chainID         int
	exchangeAddress string
	makerAddress    string // 智能钱包地址（maker），为空时使用 signer 地址
	signatureType   int    // 签名类型：0=EOA, 2=Gnosis Safe
}

// NewOrderSigner 创建订单签名器（EOA 模式）
func NewOrderSigner(signer *auth.Signer, chainID int, exchangeAddress string) *OrderSigner {
	return &OrderSigner{
		signer:          signer,
		chainID:         chainID,
		exchangeAddress: exchangeAddress,
		makerAddress:    "", // 默认使用 signer 地址
		signatureType:   SignatureTypeEOA,
	}
}

// NewOrderSignerWithMaker 创建订单签名器（Gnosis Safe 模式）
// makerAddress: 智能钱包地址，作为资产持有者和订单 maker
// signer: EOA 私钥，用于签名
func NewOrderSignerWithMaker(signer *auth.Signer, chainID int, exchangeAddress string, makerAddress string) *OrderSigner {
	signatureType := SignatureTypeEOA
	if makerAddress != "" {
		signatureType = SignatureTypeGnosisSafe
	}
	return &OrderSigner{
		signer:          signer,
		chainID:         chainID,
		exchangeAddress: exchangeAddress,
		makerAddress:    makerAddress,
		signatureType:   signatureType,
	}
}

// GetMakerAddress 获取 maker 地址（智能钱包或 EOA）
func (s *OrderSigner) GetMakerAddress() string {
	if s.makerAddress != "" {
		return s.makerAddress
	}
	return s.signer.GetAddressChecksum()
}

// CreateSignedOrder 创建签名订单
func (s *OrderSigner) CreateSignedOrder(req *CreateOrderRequest) (*SignedOrder, error) {
	// 验证价格范围
	if req.Price.LessThan(decimal.NewFromFloat(0.01)) || req.Price.GreaterThan(decimal.NewFromFloat(0.99)) {
		return nil, fmt.Errorf("price must be between 0.01 and 0.99")
	}

	// 生成随机 salt
	salt, err := common.GenerateRandomSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// 处理 nonce
	var nonce int64
	if req.Nonce != "" {
		n, err := strconv.ParseInt(req.Nonce, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid nonce: %w", err)
		}
		nonce = n
	} else {
		nonce, err = common.GenerateRandomNonce()
		if err != nil {
			return nil, fmt.Errorf("failed to generate nonce: %w", err)
		}
	}

	// 计算金额
	makerAmount, takerAmount := s.calculateAmounts(req.Side, req.Price, req.MakerAmountInQuoteToken, req.MakerAmountInBaseToken)

	// 处理过期时间
	expiration := int64(0)
	if req.ExpiresAt > 0 {
		expiration = req.ExpiresAt
	}

	// 获取 maker 地址：智能钱包地址（如果设置）或签名者地址
	makerAddr := s.GetMakerAddress()
	signerAddr := s.signer.GetAddressChecksum()

	// 构建订单负载
	payload := &auth.OrderPayload{
		Salt:          strconv.FormatInt(salt, 10), // 签名时使用字符串形式
		Maker:         makerAddr,                   // 智能钱包地址或 EOA 地址
		Signer:        signerAddr,                  // 签名者地址（始终是 EOA）
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       req.TokenID,
		MakerAmount:   makerAmount.String(),
		TakerAmount:   takerAmount.String(),
		Expiration:    strconv.FormatInt(expiration, 10),
		Nonce:         strconv.FormatInt(nonce, 10),
		FeeRateBps:    strconv.Itoa(req.FeeRateBps),
		Side:          int(req.Side),
		SignatureType: s.signatureType, // 0=EOA, 2=Gnosis Safe
	}

	// 签名订单
	signature, err := s.signer.SignOrder(payload, s.exchangeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	return &SignedOrder{
		Salt:          salt, // JSON 序列化时使用整数
		Maker:         payload.Maker,
		Signer:        payload.Signer,
		Taker:         payload.Taker,
		TokenId:       payload.TokenID,
		MakerAmount:   payload.MakerAmount,
		TakerAmount:   payload.TakerAmount,
		Expiration:    payload.Expiration,
		Nonce:         payload.Nonce,
		FeeRateBps:    payload.FeeRateBps,
		Side:          sideToString(req.Side),
		SignatureType: s.signatureType,
		Signature:     signature,
	}, nil
}

// calculateAmounts 计算订单金额
func (s *OrderSigner) calculateAmounts(side OrderSide, price, quoteAmount, baseAmount decimal.Decimal) (makerAmount, takerAmount decimal.Decimal) {
	var size decimal.Decimal

	if !quoteAmount.IsZero() {
		// 使用 USDT 金额计算
		if side == OrderSideBuy {
			// BUY: size = quoteAmount / price
			size = quoteAmount.Div(price)
		} else {
			// SELL: size = quoteAmount / price
			size = quoteAmount.Div(price)
		}
	} else if !baseAmount.IsZero() {
		size = baseAmount
	} else {
		return decimal.Zero, decimal.Zero
	}

	if side == OrderSideBuy {
		// BUY: maker 出 USDT, taker 出代币
		makerAmount = price.Mul(size).Mul(Decimal18).Round(0)
		takerAmount = size.Mul(Decimal18).Round(0)
	} else {
		// SELL: maker 出代币, taker 出 USDT
		makerAmount = size.Mul(Decimal18).Round(0)
		takerAmount = price.Mul(size).Mul(Decimal18).Round(0)
	}

	return makerAmount, takerAmount
}

// GetExchangeAddress 获取交易所地址
func (s *OrderSigner) GetExchangeAddress() string {
	return s.exchangeAddress
}

// GetSignerAddress 获取签名者地址
func (s *OrderSigner) GetSignerAddress() string {
	return s.signer.GetAddressChecksum()
}

// sideToString 订单方向转字符串
func sideToString(side OrderSide) string {
	switch side {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return "BUY"
	}
}
