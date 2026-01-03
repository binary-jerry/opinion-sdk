package clob

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/binary-jerry/opinion-sdk/auth"
)

const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewOrderSigner(t *testing.T) {
	signer, err := auth.NewSigner(testPrivateKey, 56)
	if err != nil {
		t.Fatalf("Failed to create Signer: %v", err)
	}

	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	if orderSigner == nil {
		t.Fatal("NewOrderSigner() returned nil")
	}
}

func TestOrderSignerCreateSignedOrder(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	req := &CreateOrderRequest{
		TokenID:                "12345",
		Side:                   OrderSideBuy,
		Price:                  decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
		Type:                   OrderTypeLimit,
		FeeRateBps:             0,
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder == nil {
		t.Fatal("CreateSignedOrder() returned nil")
	}

	// Verify fields
	if signedOrder.Salt == "" {
		t.Error("Salt should not be empty")
	}
	if signedOrder.Maker == "" {
		t.Error("Maker should not be empty")
	}
	if signedOrder.Signer == "" {
		t.Error("Signer should not be empty")
	}
	if signedOrder.TokenId != "12345" {
		t.Errorf("TokenId = %s, expected 12345", signedOrder.TokenId)
	}
	if signedOrder.Signature == "" {
		t.Error("Signature should not be empty")
	}
	if !strings.HasPrefix(signedOrder.Signature, "0x") {
		t.Error("Signature should start with 0x")
	}
}

func TestOrderSignerCreateSignedOrderSellSide(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	req := &CreateOrderRequest{
		TokenID:                "12345",
		Side:                   OrderSideSell,
		Price:                  decimal.NewFromFloat(0.45),
		MakerAmountInBaseToken: decimal.NewFromInt(50),
		Type:                   OrderTypeLimit,
		FeeRateBps:             0,
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder.Side != "SELL" {
		t.Errorf("Side = %s, expected SELL", signedOrder.Side)
	}
}

func TestOrderSignerCreateSignedOrderPriceValidation(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	// Price too low
	req := &CreateOrderRequest{
		TokenID:                "12345",
		Side:                   OrderSideBuy,
		Price:                  decimal.NewFromFloat(0.001),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
	}

	_, err := orderSigner.CreateSignedOrder(req)
	if err == nil {
		t.Error("CreateSignedOrder() should fail with price < 0.01")
	}

	// Price too high
	req.Price = decimal.NewFromFloat(0.999)
	_, err = orderSigner.CreateSignedOrder(req)
	if err == nil {
		t.Error("CreateSignedOrder() should fail with price > 0.99")
	}
}

func TestOrderSignerCreateSignedOrderWithNonce(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	req := &CreateOrderRequest{
		TokenID:                "12345",
		Side:                   OrderSideBuy,
		Price:                  decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
		Nonce:                  "42",
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder.Nonce != "42" {
		t.Errorf("Nonce = %s, expected 42", signedOrder.Nonce)
	}
}

func TestOrderSignerCreateSignedOrderInvalidNonce(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	req := &CreateOrderRequest{
		TokenID:                "12345",
		Side:                   OrderSideBuy,
		Price:                  decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
		Nonce:                  "invalid",
	}

	_, err := orderSigner.CreateSignedOrder(req)
	if err == nil {
		t.Error("CreateSignedOrder() should fail with invalid nonce")
	}
}

func TestOrderSignerGetExchangeAddress(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	exchangeAddr := "0x0000000000000000000000000000000000000001"
	orderSigner := NewOrderSigner(
		signer,
		56,
		exchangeAddr,
	)

	if orderSigner.GetExchangeAddress() != exchangeAddr {
		t.Errorf("GetExchangeAddress() = %s, expected %s", orderSigner.GetExchangeAddress(), exchangeAddr)
	}
}

func TestOrderSignerGetSignerAddress(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	addr := orderSigner.GetSignerAddress()
	if addr == "" {
		t.Error("GetSignerAddress() should not return empty string")
	}
	if !strings.HasPrefix(addr, "0x") {
		t.Error("GetSignerAddress() should return address starting with 0x")
	}
}

func TestSideToString(t *testing.T) {
	if sideToString(OrderSideBuy) != "BUY" {
		t.Errorf("sideToString(OrderSideBuy) = %s, expected BUY", sideToString(OrderSideBuy))
	}
	if sideToString(OrderSideSell) != "SELL" {
		t.Errorf("sideToString(OrderSideSell) = %s, expected SELL", sideToString(OrderSideSell))
	}
}

func TestCalculateAmountsBuy(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	// BUY with quote amount: price = 0.5, quote = 100 USDT
	// size = 100 / 0.5 = 200
	// makerAmount = 0.5 * 200 * 10^6 = 100000000
	// takerAmount = 200 * 10^6 = 200000000
	makerAmount, takerAmount := orderSigner.calculateAmounts(
		OrderSideBuy,
		decimal.NewFromFloat(0.5),
		decimal.NewFromInt(100),
		decimal.Zero,
	)

	expectedMaker := int64(100000000)
	expectedTaker := int64(200000000)

	if makerAmount.IntPart() != expectedMaker {
		t.Errorf("makerAmount = %d, expected %d", makerAmount.IntPart(), expectedMaker)
	}
	if takerAmount.IntPart() != expectedTaker {
		t.Errorf("takerAmount = %d, expected %d", takerAmount.IntPart(), expectedTaker)
	}
}

func TestCalculateAmountsSell(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	// SELL with base amount: price = 0.5, base = 100 tokens
	// makerAmount = 100 * 10^6 = 100000000 (tokens)
	// takerAmount = 0.5 * 100 * 10^6 = 50000000 (USDT)
	makerAmount, takerAmount := orderSigner.calculateAmounts(
		OrderSideSell,
		decimal.NewFromFloat(0.5),
		decimal.Zero,
		decimal.NewFromInt(100),
	)

	expectedMaker := int64(100000000)
	expectedTaker := int64(50000000)

	if makerAmount.IntPart() != expectedMaker {
		t.Errorf("makerAmount = %d, expected %d", makerAmount.IntPart(), expectedMaker)
	}
	if takerAmount.IntPart() != expectedTaker {
		t.Errorf("takerAmount = %d, expected %d", takerAmount.IntPart(), expectedTaker)
	}
}

func TestOrderSignerDeterministicWithSameNonceAndSalt(t *testing.T) {
	signer, _ := auth.NewSigner(testPrivateKey, 56)
	orderSigner := NewOrderSigner(
		signer,
		56,
		"0x0000000000000000000000000000000000000001",
	)

	req := &CreateOrderRequest{
		TokenID:                "12345",
		Side:                   OrderSideBuy,
		Price:                  decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
	}

	// Create two orders (salt will be different due to random generation)
	order1, _ := orderSigner.CreateSignedOrder(req)
	order2, _ := orderSigner.CreateSignedOrder(req)

	// Salt should be different (random)
	if order1.Salt == order2.Salt {
		t.Error("Each order should have a unique salt")
	}
}
