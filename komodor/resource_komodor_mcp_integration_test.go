package komodor

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func TestFlattenAuthFromConfigUsesStateForRedactedSecrets(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceKomodorMCPIntegration().Schema, map[string]interface{}{
		"name":     "integration",
		"skill_id": "skill-1",
		"connectivity": []interface{}{
			map[string]interface{}{
				"mode": "public",
			},
		},
		"mcp_server": []interface{}{
			map[string]interface{}{
				"url":       "http://example.com",
				"transport": "sse",
			},
		},
		"auth": []interface{}{
			map[string]interface{}{
				"method": "token_exchange",
				"token_exchange": []interface{}{
					map[string]interface{}{
						"token_url":            "https://auth.example.com/token",
						"client_secret":        "stored-client-secret",
						"subject_token":        []interface{}{map[string]interface{}{"value": "stored-subject-token", "type": "urn:ietf:params:oauth:token-type:jwt"}},
						"grant_type":           "urn:ietf:params:oauth:grant-type:token-exchange",
						"requested_token_type": "urn:ietf:params:oauth:token-type:access_token",
					},
				},
			},
		},
	})

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

	auth := flattenAuthFromConfig(d, cfg, bearerHeader{})
	te := auth["token_exchange"].([]map[string]interface{})[0]
	if got := te["client_secret"]; got != "stored-client-secret" {
		t.Fatalf("expected stored client secret, got %v", got)
	}
	subjectToken := te["subject_token"].([]map[string]interface{})[0]
	if got := subjectToken["value"]; got != "stored-subject-token" {
		t.Fatalf("expected stored subject token, got %v", got)
	}
}

func TestFlattenConnectivityFromConfig(t *testing.T) {
	connectivity, bearer := flattenConnectivityFromConfig(map[string]interface{}{
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
	if bearer.headerName != "Authorization" || bearer.value != "secret-token" {
		t.Fatalf("unexpected bearer extraction: %+v", bearer)
	}
}
