package komodor

import (
	"testing"
)

// TestGetV2Endpoint_EU verifies that EU endpoint construction works correctly
func TestGetV2Endpoint_EU(t *testing.T) {
	client := NewClient("test-key", "https://api.eu.komodor.com")
	
	expected := "https://api.eu.komodor.com/api/v2"
	actual := client.GetV2Endpoint()
	
	if actual != expected {
		t.Errorf("Expected GetV2Endpoint() to return %s, got %s", expected, actual)
	}
}

// TestGetDefaultEndpoint_EU verifies that EU v1 endpoint construction works correctly
func TestGetDefaultEndpoint_EU(t *testing.T) {
	client := NewClient("test-key", "https://api.eu.komodor.com")
	
	expected := "https://api.eu.komodor.com/mgmt/v1"
	actual := client.GetDefaultEndpoint()
	
	if actual != expected {
		t.Errorf("Expected GetDefaultEndpoint() to return %s, got %s", expected, actual)
	}
}
