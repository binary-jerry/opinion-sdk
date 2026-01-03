package opinion

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.APIHost != DefaultAPIHost {
		t.Errorf("APIHost = %s, want %s", config.APIHost, DefaultAPIHost)
	}
	if config.ChainID != DefaultChainID {
		t.Errorf("ChainID = %d, want %d", config.ChainID, DefaultChainID)
	}
	if config.RPCURL != DefaultRPCURL {
		t.Errorf("RPCURL = %s, want %s", config.RPCURL, DefaultRPCURL)
	}
	if config.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", config.Timeout, DefaultTimeout)
	}
}

func TestConfigValidate(t *testing.T) {
	config := &Config{}
	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	// Should fill in defaults
	if config.APIHost != DefaultAPIHost {
		t.Errorf("After Validate(), APIHost = %s, want %s", config.APIHost, DefaultAPIHost)
	}
	if config.ChainID != DefaultChainID {
		t.Errorf("After Validate(), ChainID = %d, want %d", config.ChainID, DefaultChainID)
	}
}

func TestConfigWithMethods(t *testing.T) {
	config := DefaultConfig()

	config.WithAPIKey("test-key")
	if config.APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", config.APIKey)
	}

	config.WithPrivateKey("test-private-key")
	if config.PrivateKey != "test-private-key" {
		t.Errorf("PrivateKey = %s, want test-private-key", config.PrivateKey)
	}

	config.WithMultiSigAddress("0x123")
	if config.MultiSigAddress != "0x123" {
		t.Errorf("MultiSigAddress = %s, want 0x123", config.MultiSigAddress)
	}

	config.WithChainID(1)
	if config.ChainID != 1 {
		t.Errorf("ChainID = %d, want 1", config.ChainID)
	}

	config.WithRPCURL("https://custom.rpc")
	if config.RPCURL != "https://custom.rpc" {
		t.Errorf("RPCURL = %s, want https://custom.rpc", config.RPCURL)
	}

	config.WithAPIHost("https://custom.host")
	if config.APIHost != "https://custom.host" {
		t.Errorf("APIHost = %s, want https://custom.host", config.APIHost)
	}
}

func TestConstants(t *testing.T) {
	if DefaultChainID != 56 {
		t.Errorf("DefaultChainID = %d, want 56", DefaultChainID)
	}
	if MinPrice != 0.01 {
		t.Errorf("MinPrice = %f, want 0.01", MinPrice)
	}
	if MaxPrice != 0.99 {
		t.Errorf("MaxPrice = %f, want 0.99", MaxPrice)
	}
	if PriceDecimals != 4 {
		t.Errorf("PriceDecimals = %d, want 4", PriceDecimals)
	}
	if RateLimitPerSecond != 15 {
		t.Errorf("RateLimitPerSecond = %d, want 15", RateLimitPerSecond)
	}
	if DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v, want 30s", DefaultTimeout)
	}
}
