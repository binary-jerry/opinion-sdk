package clob

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/binary-jerry/opinion-sdk/auth"
	"github.com/binary-jerry/opinion-sdk/common"
)

// Decimal6 USDT 精度
var Decimal6 = decimal.NewFromInt(1000000)

// OrderSigner 订单签名器
type OrderSigner struct {
	signer          *auth.Signer
	chainID         int
	exchangeAddress string
}

// NewOrderSigner 创建订单签名器
func NewOrderSigner(signer *auth.Signer, chainID int, exchangeAddress string) *OrderSigner {
	return &OrderSigner{
		signer:          signer,
		chainID:         chainID,
		exchangeAddress: exchangeAddress,
	}
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

	// 构建订单负载
	payload := &auth.OrderPayload{
		Salt:          salt.String(),
		Maker:         s.signer.GetAddressChecksum(),
		Signer:        s.signer.GetAddressChecksum(),
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       req.TokenID,
		MakerAmount:   makerAmount.String(),
		TakerAmount:   takerAmount.String(),
		Expiration:    strconv.FormatInt(expiration, 10),
		Nonce:         strconv.FormatInt(nonce, 10),
		FeeRateBps:    strconv.Itoa(req.FeeRateBps),
		Side:          int(req.Side),
		SignatureType: 0, // EOA
	}

	// 签名订单
	signature, err := s.signer.SignOrder(payload, s.exchangeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	return &SignedOrder{
		Salt:          payload.Salt,
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
		SignatureType: "0",
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
		makerAmount = price.Mul(size).Mul(Decimal6).Round(0)
		takerAmount = size.Mul(Decimal6).Round(0)
	} else {
		// SELL: maker 出代币, taker 出 USDT
		makerAmount = size.Mul(Decimal6).Round(0)
		takerAmount = price.Mul(size).Mul(Decimal6).Round(0)
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
