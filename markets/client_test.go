package markets

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Test with nil config
	client := NewClient(nil)
	if client == nil {
		t.Fatal("NewClient(nil) returned nil")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}

	// Test with custom config
	config := &ClientConfig{
		Host:         "https://custom.api.com",
		APIKey:       "test-key",
		Timeout:      10 * time.Second,
		MaxRetries:   5,
		RetryDelayMs: 500,
	}
	client = NewClient(config)
	if client.apiKey != "test-key" {
		t.Errorf("apiKey = %s, want test-key", client.apiKey)
	}
}

func TestClientSetAPIKey(t *testing.T) {
	client := NewClient(nil)
	client.SetAPIKey("new-key")

	if client.apiKey != "new-key" {
		t.Errorf("apiKey = %s, want new-key", client.apiKey)
	}
}
