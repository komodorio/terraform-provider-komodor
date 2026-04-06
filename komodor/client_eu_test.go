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
