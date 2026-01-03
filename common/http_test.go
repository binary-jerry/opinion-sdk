package common

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	// Test with nil config
	client := NewHTTPClient(nil)
	if client == nil {
		t.Fatal("NewHTTPClient(nil) returned nil")
	}
	if client.client.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want 30s", client.client.Timeout)
	}

	// Test with custom config
	config := &HTTPClientConfig{
		BaseURL:      "https://api.example.com/",
		Timeout:      10 * time.Second,
		MaxRetries:   5,
		RetryDelayMs: 500,
	}
	client = NewHTTPClient(config)
	if client.baseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %s, want https://api.example.com", client.baseURL)
	}
	if client.maxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", client.maxRetries)
	}
}

func TestHTTPClientSetDefaultHeader(t *testing.T) {
	client := NewHTTPClient(nil)
	client.SetDefaultHeader("X-Custom", "value")

	if client.defaultHeaders["X-Custom"] != "value" {
		t.Errorf("defaultHeaders[X-Custom] = %s, want value", client.defaultHeaders["X-Custom"])
	}
}

func TestHTTPClientGetBaseURL(t *testing.T) {
	config := &HTTPClientConfig{
		BaseURL: "https://api.example.com",
	}
	client := NewHTTPClient(config)

	if client.GetBaseURL() != "https://api.example.com" {
		t.Errorf("GetBaseURL() = %s, want https://api.example.com", client.GetBaseURL())
	}
}

func TestHTTPClientGet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("Expected /test, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("result[status] = %s, want ok", result["status"])
	}
}

func TestHTTPClientGetWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	params := struct {
		Page  int `url:"page"`
		Limit int `url:"limit"`
	}{
		Page:  1,
		Limit: 10,
	}

	var result map[string]string
	err := client.Get(context.Background(), "/test", &params, &result)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestHTTPClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("Expected body.name=test, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	body := map[string]string{"name": "test"}
	var result map[string]string
	err := client.Post(context.Background(), "/test", body, &result)
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("result[id] = %s, want 123", result["id"])
	}
}

func TestHTTPClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"deleted": "true"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	var result map[string]string
	err := client.Delete(context.Background(), "/test/123", &result)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
}

func TestHTTPClientDeleteWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["reason"] != "test" {
			t.Errorf("Expected body.reason=test, got %s", body["reason"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"deleted": "true"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	body := map[string]string{"reason": "test"}
	var result map[string]string
	err := client.DeleteWithBody(context.Background(), "/test/123", body, &result)
	if err != nil {
		t.Fatalf("DeleteWithBody() error: %v", err)
	}
}

func TestHTTPClientDoWithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("apikey") != "test-key" {
			t.Errorf("Expected apikey=test-key, got %s", r.Header.Get("apikey"))
		}
		if r.Header.Get("address") != "0x123" {
			t.Errorf("Expected address=0x123, got %s", r.Header.Get("address"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	authHeaders := map[string]string{
		"apikey":  "test-key",
		"address": "0x123",
	}

	var result map[string]string
	err := client.DoWithAuth(context.Background(), "GET", "/test", nil, authHeaders, &result)
	if err != nil {
		t.Fatalf("DoWithAuth() error: %v", err)
	}
}

func TestHTTPClientDoWithAuthAndParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", r.URL.Query().Get("page"))
		}
		if r.Header.Get("apikey") != "test-key" {
			t.Errorf("Expected apikey=test-key, got %s", r.Header.Get("apikey"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	params := struct {
		Page int `url:"page"`
	}{
		Page: 1,
	}

	authHeaders := map[string]string{
		"apikey": "test-key",
	}

	var result map[string]string
	err := client.DoWithAuthAndParams(context.Background(), "GET", "/test", &params, nil, authHeaders, &result)
	if err != nil {
		t.Fatalf("DoWithAuthAndParams() error: %v", err)
	}
}

func TestHTTPClientError4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"code": "BAD_REQUEST",
			"msg":  "Invalid input",
		})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:    server.URL,
		MaxRetries: 0,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func TestHTTPClientError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:    server.URL,
		MaxRetries: 3,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Error("Expected IsUnauthorized to return true")
	}
}

func TestHTTPClientError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"code": "NOT_FOUND"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:    server.URL,
		MaxRetries: 3,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Error("Expected IsNotFound to return true")
	}
}

func TestHTTPClientError429Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"code": "RATE_LIMITED"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:      server.URL,
		MaxRetries:   3,
		RetryDelayMs: 10,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestHTTPClientError5xxRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:      server.URL,
		MaxRetries:   3,
		RetryDelayMs: 10,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestHTTPClientContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:      server.URL,
		MaxRetries:   3,
		RetryDelayMs: 50,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result map[string]string
	err := client.Get(ctx, "/test", nil, &result)
	if err == nil {
		t.Fatal("Expected error due to cancelled context")
	}
}

func TestStructToQueryString(t *testing.T) {
	tests := []struct {
		name     string
		params   interface{}
		contains []string
	}{
		{
			name: "simple struct",
			params: struct {
				Page  int    `url:"page"`
				Limit int    `url:"limit"`
				Name  string `url:"name"`
			}{
				Page:  1,
				Limit: 10,
				Name:  "test",
			},
			contains: []string{"page=1", "limit=10", "name=test"},
		},
		{
			name: "with omitempty",
			params: struct {
				Page  int    `url:"page,omitempty"`
				Limit int    `url:"limit,omitempty"`
				Name  string `url:"name,omitempty"`
			}{
				Page:  1,
				Limit: 0,
				Name:  "",
			},
			contains: []string{"page=1"},
		},
		{
			name: "with pointer",
			params: struct {
				Active *bool `url:"active,omitempty"`
			}{
				Active: BoolPtr(true),
			},
			contains: []string{"active=true"},
		},
		{
			name:     "nil params",
			params:   nil,
			contains: []string{},
		},
		{
			name: "ignored field",
			params: struct {
				Page    int `url:"page"`
				Ignored int `url:"-"`
			}{
				Page:    1,
				Ignored: 100,
			},
			contains: []string{"page=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := structToQueryString(tt.params)
			for _, expected := range tt.contains {
				if expected != "" && !containsSubstring(result, expected) {
					t.Errorf("Expected %s to contain %s", result, expected)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHTTPClientDefaultHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Errorf("Expected X-Custom=value, got %s", r.Header.Get("X-Custom"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})
	client.SetDefaultHeader("X-Custom", "value")

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestHTTPClientEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
	})

	err := client.Delete(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
}

func TestBuildURLWithExistingQuery(t *testing.T) {
	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: "https://api.example.com",
	})

	params := struct {
		Page int `url:"page"`
	}{
		Page: 1,
	}

	url := client.buildURL("/test?foo=bar", &params)
	if url != "https://api.example.com/test?foo=bar&page=1" {
		t.Errorf("buildURL() = %s, want https://api.example.com/test?foo=bar&page=1", url)
	}
}
