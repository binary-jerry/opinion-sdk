package opinion

import (
	"testing"
)

const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewPublicSDK(t *testing.T) {
	sdk := NewPublicSDK(nil)
	if sdk == nil {
		t.Fatal("NewPublicSDK() returned nil")
	}
	if sdk.Markets == nil {
		t.Error("Markets client should not be nil")
	}
	if sdk.Trading != nil {
		t.Error("Trading client should be nil for public SDK")
	}
}

func TestNewPublicSDKWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.APIKey = "test-key"

	sdk := NewPublicSDK(config)
	if sdk == nil {
		t.Fatal("NewPublicSDK() returned nil")
	}
	if sdk.GetConfig().APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", sdk.GetConfig().APIKey)
	}
}

func TestNewSDK(t *testing.T) {
	config := DefaultConfig()
	sdk, err := NewSDK(config, testPrivateKey)
	if err != nil {
		t.Fatalf("NewSDK() error: %v", err)
	}
	if sdk == nil {
		t.Fatal("NewSDK() returned nil")
	}
	if sdk.Markets == nil {
		t.Error("Markets client should not be nil")
	}
	if sdk.Trading == nil {
		t.Error("Trading client should not be nil")
	}
}

func TestNewSDKWithoutPrivateKey(t *testing.T) {
	config := DefaultConfig()
	_, err := NewSDK(config, "")
	if err == nil {
		t.Error("NewSDK() should fail without private key")
	}
}

func TestNewTradingSDK(t *testing.T) {
	config := DefaultConfig()
	sdk, err := NewTradingSDK(config, testPrivateKey)
	if err != nil {
		t.Fatalf("NewTradingSDK() error: %v", err)
	}
	if sdk == nil {
		t.Fatal("NewTradingSDK() returned nil")
	}
}

func TestSDKGetAddress(t *testing.T) {
	config := DefaultConfig()
	sdk, _ := NewSDK(config, testPrivateKey)

	address := sdk.GetAddress()
	if address == "" {
		t.Error("GetAddress() should not return empty string")
	}
}

func TestSDKGetAddressPublic(t *testing.T) {
	sdk := NewPublicSDK(nil)

	address := sdk.GetAddress()
	if address != "" {
		t.Errorf("GetAddress() for public SDK should return empty string, got %s", address)
	}
}

func TestSDKSetAPIKey(t *testing.T) {
	sdk := NewPublicSDK(nil)
	sdk.SetAPIKey("new-key")

	if sdk.GetConfig().APIKey != "new-key" {
		t.Errorf("APIKey = %s, want new-key", sdk.GetConfig().APIKey)
	}
}

func TestSDKGetConfig(t *testing.T) {
	config := DefaultConfig()
	config.APIKey = "test-key"

	sdk := NewPublicSDK(config)
	gotConfig := sdk.GetConfig()

	if gotConfig.APIKey != "test-key" {
		t.Errorf("GetConfig().APIKey = %s, want test-key", gotConfig.APIKey)
	}
}
