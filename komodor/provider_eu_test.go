package komodor

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestProviderConfigure_WithEUEndpoint verifies that provider correctly configures client with EU endpoint
func TestProviderConfigure_WithEUEndpoint(t *testing.T) {
	provider := Provider()

	d := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"api_key": "test-api-key-123456789012345678901234",
		"api_url": "https://api.eu.komodor.com",
	})

	ctx := context.Background()
	meta, diags := provider.ConfigureContextFunc(ctx, d)

	if diags != nil && diags.HasError() {
		t.Fatalf("Provider configuration failed: %v", diags)
	}

	client, ok := meta.(*Client)
	if !ok {
		t.Fatalf("Expected *Client, got %T", meta)
	}

	if client.BaseURL != "https://api.eu.komodor.com" {
		t.Errorf("Expected BaseURL 'https://api.eu.komodor.com', got '%s'", client.BaseURL)
	}
}

// TestProviderConfigure_DefaultToUS verifies backward compatibility - defaults to US when api_url not set
func TestProviderConfigure_DefaultToUS(t *testing.T) {
	provider := Provider()

	d := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"api_key": "test-api-key-123456789012345678901234",
		// api_url not set - should default to US
	})

	ctx := context.Background()
	meta, diags := provider.ConfigureContextFunc(ctx, d)

	if diags != nil && diags.HasError() {
		t.Fatalf("Provider configuration failed: %v", diags)
	}

	client, ok := meta.(*Client)
	if !ok {
		t.Fatalf("Expected *Client, got %T", meta)
	}

	if client.BaseURL != DefaultAPIBaseURL {
		t.Errorf("Expected BaseURL '%s' (default US), got '%s'", DefaultAPIBaseURL, client.BaseURL)
	}
}
