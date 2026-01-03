package auth

import (
	"strings"
	"testing"
)

// 测试用私钥
const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewSigner(t *testing.T) {
	// Test with valid private key
	signer, err := NewSigner(testPrivateKey, 56)
	if err != nil {
		t.Fatalf("NewSigner() error: %v", err)
	}
	if signer == nil {
		t.Fatal("NewSigner() returned nil")
	}

	// Test with 0x prefix
	signer2, err := NewSigner("0x"+testPrivateKey, 56)
	if err != nil {
		t.Fatalf("NewSigner() with 0x prefix error: %v", err)
	}
	if signer.GetAddress() != signer2.GetAddress() {
		t.Error("Same key with/without 0x prefix should produce same address")
	}

	// Test with invalid key
	_, err = NewSigner("invalid", 56)
	if err == nil {
		t.Error("NewSigner() should fail with invalid key")
	}

	// Test with empty key
	_, err = NewSigner("", 56)
	if err == nil {
		t.Error("NewSigner() should fail with empty key")
	}
}

func TestNewSignerFromKey(t *testing.T) {
	// Test with nil key
	_, err := NewSignerFromKey(nil, 56)
	if err == nil {
		t.Error("NewSignerFromKey() should fail with nil key")
	}
}

func TestSignerGetAddress(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)

	address := signer.GetAddress()
	if !strings.HasPrefix(address, "0x") {
		t.Errorf("Address should start with 0x, got %s", address)
	}
	if len(address) != 42 {
		t.Errorf("Address should be 42 chars, got %d", len(address))
	}
	// Address should be lowercase
	if address != strings.ToLower(address) {
		t.Error("GetAddress() should return lowercase address")
	}
}

func TestSignerGetAddressChecksum(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)

	address := signer.GetAddressChecksum()
	if !strings.HasPrefix(address, "0x") {
		t.Errorf("Checksum address should start with 0x, got %s", address)
	}
	if len(address) != 42 {
		t.Errorf("Checksum address should be 42 chars, got %d", len(address))
	}
}

func TestSignerGetChainID(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)
	if signer.GetChainID() != 56 {
		t.Errorf("GetChainID() = %d, expected 56", signer.GetChainID())
	}

	signer2, _ := NewSigner(testPrivateKey, 1)
	if signer2.GetChainID() != 1 {
		t.Errorf("GetChainID() = %d, expected 1", signer2.GetChainID())
	}
}

func TestSignerSignMessage(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)

	message := []byte("Hello, World!")
	signature, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	// Signature should be 65 bytes (r: 32, s: 32, v: 1)
	if len(signature) != 65 {
		t.Errorf("Signature should be 65 bytes, got %d", len(signature))
	}

	// v should be 27 or 28
	v := signature[64]
	if v != 27 && v != 28 {
		t.Errorf("v should be 27 or 28, got %d", v)
	}
}

func TestSignerSignOrder(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)

	order := &OrderPayload{
		Salt:          "12345",
		Maker:         signer.GetAddress(),
		Signer:        signer.GetAddress(),
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       "100",
		MakerAmount:   "1000000",
		TakerAmount:   "500000",
		Expiration:    "0",
		Nonce:         "0",
		FeeRateBps:    "0",
		Side:          0,
		SignatureType: 0,
	}

	exchangeAddr := "0x0000000000000000000000000000000000000001"
	signature, err := signer.SignOrder(order, exchangeAddr)
	if err != nil {
		t.Fatalf("SignOrder() error: %v", err)
	}

	if !strings.HasPrefix(signature, "0x") {
		t.Errorf("Signature should start with 0x, got %s", signature)
	}
}

func TestSignerSignOrderInvalidParams(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)
	exchangeAddr := "0x0000000000000000000000000000000000000001"

	tests := []struct {
		name  string
		order *OrderPayload
	}{
		{
			name: "invalid salt",
			order: &OrderPayload{
				Salt: "invalid", Maker: signer.GetAddress(), Signer: signer.GetAddress(),
				Taker: "0x0000000000000000000000000000000000000000", TokenID: "100",
				MakerAmount: "1000000", TakerAmount: "500000", Expiration: "0",
				Nonce: "0", FeeRateBps: "0", Side: 0, SignatureType: 0,
			},
		},
		{
			name: "invalid tokenID",
			order: &OrderPayload{
				Salt: "12345", Maker: signer.GetAddress(), Signer: signer.GetAddress(),
				Taker: "0x0000000000000000000000000000000000000000", TokenID: "invalid",
				MakerAmount: "1000000", TakerAmount: "500000", Expiration: "0",
				Nonce: "0", FeeRateBps: "0", Side: 0, SignatureType: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := signer.SignOrder(tt.order, exchangeAddr)
			if err == nil {
				t.Errorf("%s: expected error but got none", tt.name)
			}
		})
	}
}

func TestSignerDeterministicSignature(t *testing.T) {
	signer, _ := NewSigner(testPrivateKey, 56)

	message := []byte("Test message for deterministic signing")

	sig1, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("First SignMessage() error: %v", err)
	}

	sig2, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("Second SignMessage() error: %v", err)
	}

	// go-ethereum crypto.Sign is deterministic
	if string(sig1) != string(sig2) {
		t.Error("Signatures should be deterministic for the same message")
	}
}
