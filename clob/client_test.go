package clob

import (
	"testing"
	"time"
)

const testPrivateKeyClient = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewClient(t *testing.T) {
	// Test with nil config
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient(nil) error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient(nil) returned nil")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}

	// Test with custom config without private key
	config := &ClientConfig{
		Host:            "https://custom.api.com",
		APIKey:          "test-key",
		ChainID:         56,
		ExchangeAddress: "0x123",
		Timeout:         10 * time.Second,
		MaxRetries:      5,
		RetryDelayMs:    500,
	}
	client, err = NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client.apiKey != "test-key" {
		t.Errorf("apiKey = %s, want test-key", client.apiKey)
	}
	if client.signer != nil {
		t.Error("signer should be nil without private key")
	}

	// Test with private key
	config.PrivateKey = testPrivateKeyClient
	client, err = NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() with private key error: %v", err)
	}
	if client.signer == nil {
		t.Error("signer should not be nil with private key")
	}
	if client.orderSigner == nil {
		t.Error("orderSigner should not be nil with private key")
	}
}

func TestClientSetAPIKey(t *testing.T) {
	client, _ := NewClient(nil)
	client.SetAPIKey("new-key")

	if client.apiKey != "new-key" {
		t.Errorf("apiKey = %s, want new-key", client.apiKey)
	}
}

func TestClientSetPrivateKey(t *testing.T) {
	client, _ := NewClient(nil)

	err := client.SetPrivateKey(testPrivateKeyClient)
	if err != nil {
		t.Fatalf("SetPrivateKey() error: %v", err)
	}

	if client.signer == nil {
		t.Error("signer should not be nil after SetPrivateKey")
	}
	if client.orderSigner == nil {
		t.Error("orderSigner should not be nil after SetPrivateKey")
	}

	// Test with invalid private key
	err = client.SetPrivateKey("invalid")
	if err == nil {
		t.Error("SetPrivateKey() should fail with invalid key")
	}
}

func TestClientGetAddress(t *testing.T) {
	// Without private key
	client, _ := NewClient(nil)
	if client.GetAddress() != "" {
		t.Errorf("GetAddress() without signer should return empty, got %s", client.GetAddress())
	}

	// With private key
	config := &ClientConfig{
		PrivateKey: testPrivateKeyClient,
		ChainID:    56,
	}
	client, _ = NewClient(config)
	addr := client.GetAddress()
	if addr == "" {
		t.Error("GetAddress() with signer should not return empty")
	}
}

func TestClientGetOrderSigner(t *testing.T) {
	// Without private key
	client, _ := NewClient(nil)
	if client.GetOrderSigner() != nil {
		t.Error("GetOrderSigner() without signer should return nil")
	}

	// With private key
	config := &ClientConfig{
		PrivateKey:      testPrivateKeyClient,
		ChainID:         56,
		ExchangeAddress: "0x123",
	}
	client, _ = NewClient(config)
	if client.GetOrderSigner() == nil {
		t.Error("GetOrderSigner() with signer should not return nil")
	}
}
