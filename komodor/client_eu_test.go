package komodor

import (
	"testing"
)

func TestNewClient_WithUSEndpoint(t *testing.T) {
	apiKey := "test-api-key-123456789012345678901234"
	baseURL := "https://api.komodor.com"
	
	client := NewClient(apiKey, baseURL)
	
	if client.ApiKey != apiKey {
		t.Errorf("Expected ApiKey %s, got %s", apiKey, client.ApiKey)
	}
	
	if client.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}
}

func TestNewClient_WithEUEndpoint(t *testing.T) {
	apiKey := "test-api-key-123456789012345678901234"
	baseURL := "https://api.eu.komodor.com"
	
	client := NewClient(apiKey, baseURL)
	
	if client.ApiKey != apiKey {
		t.Errorf("Expected ApiKey %s, got %s", apiKey, client.ApiKey)
	}
	
	if client.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}
}

func TestGetV2Endpoint_US(t *testing.T) {
	client := NewClient("test-key", "https://api.komodor.com")
	
	expected := "https://api.komodor.com/api/v2"
	actual := client.GetV2Endpoint()
	
	if actual != expected {
		t.Errorf("Expected GetV2Endpoint() to return %s, got %s", expected, actual)
	}
}

func TestGetV2Endpoint_EU(t *testing.T) {
	client := NewClient("test-key", "https://api.eu.komodor.com")
	
	expected := "https://api.eu.komodor.com/api/v2"
	actual := client.GetV2Endpoint()
	
	if actual != expected {
		t.Errorf("Expected GetV2Endpoint() to return %s, got %s", expected, actual)
	}
}

func TestGetDefaultEndpoint_US(t *testing.T) {
	client := NewClient("test-key", "https://api.komodor.com")
	
	expected := "https://api.komodor.com/mgmt/v1"
	actual := client.GetDefaultEndpoint()
	
	if actual != expected {
		t.Errorf("Expected GetDefaultEndpoint() to return %s, got %s", expected, actual)
	}
}

func TestGetDefaultEndpoint_EU(t *testing.T) {
	client := NewClient("test-key", "https://api.eu.komodor.com")
	
	expected := "https://api.eu.komodor.com/mgmt/v1"
	actual := client.GetDefaultEndpoint()
	
	if actual != expected {
		t.Errorf("Expected GetDefaultEndpoint() to return %s, got %s", expected, actual)
	}
}

func TestGetV2Endpoint_CustomBaseURL(t *testing.T) {
	// Test with a custom base URL to ensure the method works with any base URL
	customBaseURL := "https://custom.komodor.com"
	client := NewClient("test-key", customBaseURL)
	
	expected := "https://custom.komodor.com/api/v2"
	actual := client.GetV2Endpoint()
	
	if actual != expected {
		t.Errorf("Expected GetV2Endpoint() to return %s, got %s", expected, actual)
	}
}

func TestGetDefaultEndpoint_CustomBaseURL(t *testing.T) {
	// Test with a custom base URL to ensure the method works with any base URL
	customBaseURL := "https://custom.komodor.com"
	client := NewClient("test-key", customBaseURL)
	
	expected := "https://custom.komodor.com/mgmt/v1"
	actual := client.GetDefaultEndpoint()
	
	if actual != expected {
		t.Errorf("Expected GetDefaultEndpoint() to return %s, got %s", expected, actual)
	}
}
