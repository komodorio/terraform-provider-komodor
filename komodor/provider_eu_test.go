package komodor

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProviderConfigure_WithEUEndpoint(t *testing.T) {
	provider := Provider()
	
	// Create a resource data with EU endpoint
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
	
	if client.ApiKey != "test-api-key-123456789012345678901234" {
		t.Errorf("Expected ApiKey 'test-api-key-123456789012345678901234', got '%s'", client.ApiKey)
	}
}

func TestProviderConfigure_WithUSEndpoint(t *testing.T) {
	provider := Provider()
	
	// Create a resource data with US endpoint (explicit)
	d := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"api_key": "test-api-key-123456789012345678901234",
		"api_url": "https://api.komodor.com",
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
	
	if client.BaseURL != "https://api.komodor.com" {
		t.Errorf("Expected BaseURL 'https://api.komodor.com', got '%s'", client.BaseURL)
	}
}

func TestProviderConfigure_DefaultToUS(t *testing.T) {
	provider := Provider()
	
	// Create a resource data without api_url (should default to US)
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
	
	// Should default to US endpoint
	expected := DefaultAPIBaseURL
	if client.BaseURL != expected {
		t.Errorf("Expected BaseURL '%s' (default US), got '%s'", expected, client.BaseURL)
	}
}

func TestProviderConfigure_WithEnvVar(t *testing.T) {
	// Set environment variable
	originalValue := os.Getenv(KomodorAPIURLEnvName)
	defer func() {
		if originalValue != "" {
			os.Setenv(KomodorAPIURLEnvName, originalValue)
		} else {
			os.Unsetenv(KomodorAPIURLEnvName)
		}
	}()
	
	os.Setenv(KomodorAPIURLEnvName, "https://api.eu.komodor.com")
	
	provider := Provider()
	
	// Create a resource data without api_url (should use env var)
	d := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"api_key": "test-api-key-123456789012345678901234",
		// api_url not set - should use KOMODOR_API_URL env var
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
		t.Errorf("Expected BaseURL 'https://api.eu.komodor.com' (from env var), got '%s'", client.BaseURL)
	}
}

func TestProviderSchema_ApiURLField(t *testing.T) {
	provider := Provider()
	
	// Verify api_url field exists in schema
	apiURLSchema, exists := provider.Schema["api_url"]
	if !exists {
		t.Fatal("Expected 'api_url' field in provider schema, but it doesn't exist")
	}
	
	if apiURLSchema.Type != schema.TypeString {
		t.Errorf("Expected api_url Type to be TypeString, got %v", apiURLSchema.Type)
	}
	
	if !apiURLSchema.Optional {
		t.Error("Expected api_url to be Optional")
	}
	
	if apiURLSchema.Description == "" {
		t.Error("Expected api_url to have a description")
	}
}

func TestProviderConfigure_EmptyApiURLDefaultsToUS(t *testing.T) {
	provider := Provider()
	
	// Create a resource data with empty api_url string
	d := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"api_key": "test-api-key-123456789012345678901234",
		"api_url": "", // Empty string should default to US
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
	
	// Should default to US endpoint
	expected := DefaultAPIBaseURL
	if client.BaseURL != expected {
		t.Errorf("Expected BaseURL '%s' (default US), got '%s'", expected, client.BaseURL)
	}
}
