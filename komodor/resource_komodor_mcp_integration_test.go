package komodor

import (
	"testing"
)

func TestValidateConnectivityBlock(t *testing.T) {
	err := validateConnectivityBlock(map[string]interface{}{
		"mode": "agent-tunnel",
	})
	if err == nil {
		t.Fatalf("expected error when provider_cluster is missing")
	}

	err = validateConnectivityBlock(map[string]interface{}{
		"mode":             "agent-tunnel",
		"provider_cluster": "hub-cluster",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAuthMethodBlock(t *testing.T) {
	err := validateAuthMethodBlock("static_token", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected missing static_token block error")
	}

	err = validateAuthMethodBlock("oauth2_client_credentials", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected missing oauth2_client_credentials block error")
	}

	auth := map[string]interface{}{
		"token_exchange": []interface{}{
			map[string]interface{}{
				"subject_token": []interface{}{
					map[string]interface{}{
						"value":     "abc",
						"file_path": "/tmp/token",
					},
				},
			},
		},
	}
	err = validateAuthMethodBlock("token_exchange", auth)
	if err == nil {
		t.Fatalf("expected subject_token xor validation error")
	}
}

func TestFlattenAuthFromConfigOmitsRedactedSecrets(t *testing.T) {
	cfg := map[string]interface{}{
		"auth_method": "rfc8693_token_exchange",
		"auth_params": map[string]interface{}{
			"token_url":            "https://auth.example.com/token",
			"subject_token_type":   "urn:ietf:params:oauth:token-type:jwt",
			"subject_token":        redactedSecretPlaceholder,
			"client_secret":        redactedSecretPlaceholder,
			"requested_token_type": "urn:ietf:params:oauth:token-type:access_token",
		},
	}

	auth, stripKey := flattenAuthFromConfig(cfg)
	if stripKey != "" {
		t.Fatalf("expected no header strip key for token_exchange, got %q", stripKey)
	}
	te := auth["token_exchange"].([]map[string]interface{})[0]
	if _, ok := te["client_secret"]; ok {
		t.Fatalf("expected client_secret to be omitted when API returns redacted placeholder")
	}
	subjectToken := te["subject_token"].([]map[string]interface{})[0]
	if _, ok := subjectToken["value"]; ok {
		t.Fatalf("expected subject_token.value to be omitted when API returns redacted placeholder")
	}
}

func TestFlattenConnectivityFromConfig(t *testing.T) {
	connectivity := flattenConnectivityFromConfig(map[string]interface{}{
		"use_tunnel":      true,
		"tunnel_cluster":  "hub-1",
		"headers":         map[string]interface{}{"Authorization": "Bearer secret-token", "X-Tenant": "acme"},
		"auth_method":     "static_token",
		"configurationId": "ignored",
	})

	if got := connectivity["mode"]; got != "agent-tunnel" {
		t.Fatalf("expected mode=agent-tunnel, got %v", got)
	}
	if got := connectivity["provider_cluster"]; got != "hub-1" {
		t.Fatalf("expected provider_cluster=hub-1, got %v", got)
	}
}

func TestFlattenAuthFromConfigStaticTokenReturnsStripKey(t *testing.T) {
	cfg := map[string]interface{}{
		"auth_method": "static_token",
		"headers": map[string]interface{}{
			"X-Custom-Auth": "Bearer my-token",
			"X-Tenant":      "acme",
		},
	}
	auth, stripKey := flattenAuthFromConfig(cfg)
	if stripKey != "X-Custom-Auth" {
		t.Fatalf("expected strip key X-Custom-Auth, got %q", stripKey)
	}
	st := auth["static_token"].([]map[string]interface{})[0]
	if got := st["header_name"]; got != "X-Custom-Auth" {
		t.Fatalf("expected header_name X-Custom-Auth, got %v", got)
	}
	if got := st["value"]; got != "my-token" {
		t.Fatalf("expected value my-token, got %v", got)
	}
}

func TestFlattenAuthFromConfigNonStaticAuthLeavesNoStripKey(t *testing.T) {
	cfg := map[string]interface{}{
		"auth_method": "oauth2_client_credentials",
		"headers": map[string]interface{}{
			"Authorization": "Bearer should-not-strip",
		},
		"auth_params": map[string]interface{}{
			"token_url":     "https://token.example.com",
			"client_id":     "id",
			"client_secret": "secret",
		},
	}
	auth, stripKey := flattenAuthFromConfig(cfg)
	if stripKey != "" {
		t.Fatalf("expected no strip key for oauth2, got %q", stripKey)
	}
	if _, ok := auth["static_token"]; ok {
		t.Fatalf("expected no static_token block for oauth2")
	}
}
