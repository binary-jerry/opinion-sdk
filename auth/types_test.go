package auth

import (
	"math/big"
	"testing"
)

func TestAuthHeadersToMap(t *testing.T) {
	headers := &AuthHeaders{
		APIKey:    "test-key",
		Address:   "0x123",
		Timestamp: "1234567890",
		Signature: "0xabc",
	}

	m := headers.ToMap()

	if m["apikey"] != "test-key" {
		t.Errorf("apikey = %s, want test-key", m["apikey"])
	}
	if m["address"] != "0x123" {
		t.Errorf("address = %s, want 0x123", m["address"])
	}
	if m["timestamp"] != "1234567890" {
		t.Errorf("timestamp = %s, want 1234567890", m["timestamp"])
	}
	if m["signature"] != "0xabc" {
		t.Errorf("signature = %s, want 0xabc", m["signature"])
	}
}

func TestAuthHeadersToMapPartial(t *testing.T) {
	headers := &AuthHeaders{
		APIKey: "test-key",
	}

	m := headers.ToMap()

	if m["apikey"] != "test-key" {
		t.Errorf("apikey = %s, want test-key", m["apikey"])
	}
	if _, ok := m["address"]; ok {
		t.Error("address should not be present")
	}
	if _, ok := m["timestamp"]; ok {
		t.Error("timestamp should not be present")
	}
	if _, ok := m["signature"]; ok {
		t.Error("signature should not be present")
	}
}

func TestCredentials(t *testing.T) {
	creds := &Credentials{
		APIKey:  "api-key-123",
		Address: "0x123",
	}

	if creds.APIKey != "api-key-123" {
		t.Errorf("APIKey = %s, want api-key-123", creds.APIKey)
	}
	if creds.Address != "0x123" {
		t.Errorf("Address = %s, want 0x123", creds.Address)
	}
}

func TestTypedDataField(t *testing.T) {
	field := TypedDataField{
		Name: "salt",
		Type: "uint256",
	}

	if field.Name != "salt" {
		t.Errorf("Name = %s, want salt", field.Name)
	}
	if field.Type != "uint256" {
		t.Errorf("Type = %s, want uint256", field.Type)
	}
}

func TestTypedDataDomain(t *testing.T) {
	domain := TypedDataDomain{
		Name:              "Test Domain",
		Version:           "1",
		ChainId:           big.NewInt(56),
		VerifyingContract: "0x123",
	}

	if domain.Name != "Test Domain" {
		t.Errorf("Name = %s, want Test Domain", domain.Name)
	}
	if domain.Version != "1" {
		t.Errorf("Version = %s, want 1", domain.Version)
	}
	if domain.ChainId.Int64() != 56 {
		t.Errorf("ChainId = %d, want 56", domain.ChainId.Int64())
	}
	if domain.VerifyingContract != "0x123" {
		t.Errorf("VerifyingContract = %s, want 0x123", domain.VerifyingContract)
	}
}

func TestTypedData(t *testing.T) {
	typedData := &TypedData{
		Types:       OrderTypes,
		PrimaryType: "Order",
		Domain: TypedDataDomain{
			Name:    "Test",
			Version: "1",
			ChainId: big.NewInt(56),
		},
		Message: map[string]interface{}{
			"salt": big.NewInt(12345),
		},
	}

	if typedData.PrimaryType != "Order" {
		t.Errorf("PrimaryType = %s, want Order", typedData.PrimaryType)
	}
	if len(typedData.Types) == 0 {
		t.Error("Types should not be empty")
	}
}

func TestOrderPayload(t *testing.T) {
	payload := &OrderPayload{
		Salt:          "12345",
		Maker:         "0x123",
		Signer:        "0x123",
		Taker:         "0x000",
		TokenID:       "67890",
		MakerAmount:   "1000000",
		TakerAmount:   "500000",
		Expiration:    "0",
		Nonce:         "1",
		FeeRateBps:    "0",
		Side:          0,
		SignatureType: 0,
	}

	if payload.Salt != "12345" {
		t.Errorf("Salt = %s, want 12345", payload.Salt)
	}
	if payload.Side != 0 {
		t.Errorf("Side = %d, want 0", payload.Side)
	}
	if payload.SignatureType != 0 {
		t.Errorf("SignatureType = %d, want 0", payload.SignatureType)
	}
}

func TestOrderTypes(t *testing.T) {
	if _, ok := OrderTypes["Order"]; !ok {
		t.Error("OrderTypes should have 'Order' type")
	}

	orderFields := OrderTypes["Order"]
	if len(orderFields) != 12 {
		t.Errorf("Order should have 12 fields, got %d", len(orderFields))
	}

	// Check some fields
	fieldNames := make(map[string]string)
	for _, f := range orderFields {
		fieldNames[f.Name] = f.Type
	}

	if fieldNames["salt"] != "uint256" {
		t.Errorf("salt type = %s, want uint256", fieldNames["salt"])
	}
	if fieldNames["maker"] != "address" {
		t.Errorf("maker type = %s, want address", fieldNames["maker"])
	}
	if fieldNames["side"] != "uint8" {
		t.Errorf("side type = %s, want uint8", fieldNames["side"])
	}
}

func TestOpinionExchangeDomain(t *testing.T) {
	domain := OpinionExchangeDomain(56, "0x123456")

	if domain.Name != "OPINION CTF Exchange" {
		t.Errorf("Name = %s, want OPINION CTF Exchange", domain.Name)
	}
	if domain.Version != "1" {
		t.Errorf("Version = %s, want 1", domain.Version)
	}
	if domain.ChainId.Int64() != 56 {
		t.Errorf("ChainId = %d, want 56", domain.ChainId.Int64())
	}
	if domain.VerifyingContract != "0x123456" {
		t.Errorf("VerifyingContract = %s, want 0x123456", domain.VerifyingContract)
	}
}

func TestOpinionExchangeDomainDifferentChains(t *testing.T) {
	tests := []struct {
		chainID  int
		exchange string
	}{
		{56, "0xabc"},
		{97, "0xdef"},
		{1, "0x123"},
	}

	for _, tt := range tests {
		domain := OpinionExchangeDomain(tt.chainID, tt.exchange)
		if domain.ChainId.Int64() != int64(tt.chainID) {
			t.Errorf("ChainId = %d, want %d", domain.ChainId.Int64(), tt.chainID)
		}
		if domain.VerifyingContract != tt.exchange {
			t.Errorf("VerifyingContract = %s, want %s", domain.VerifyingContract, tt.exchange)
		}
	}
}

func TestWallet(t *testing.T) {
	wallet := &Wallet{}
	// Wallet is just a struct to hold address and private key
	// We can't test much without actually creating a key
	if wallet.PrivateKey != nil {
		t.Error("Empty wallet should have nil PrivateKey")
	}
}
