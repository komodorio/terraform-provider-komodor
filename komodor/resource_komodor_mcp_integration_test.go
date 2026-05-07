package komodor

import (
	"testing"
)

func TestValidateConnectivityBlock_AgentTunnelRequiresBlock(t *testing.T) {
	err := validateConnectivityBlock(map[string]interface{}{
		"mode": "agent-tunnel",
	})
	if err == nil {
		t.Fatalf("expected error when agent_tunnel block is missing")
	}
}

func TestValidateConnectivityBlock_AgentTunnelRequiresProviderCluster(t *testing.T) {
	err := validateConnectivityBlock(map[string]interface{}{
		"mode": "agent-tunnel",
		"agent_tunnel": []interface{}{
			map[string]interface{}{"provider_cluster": ""},
		},
	})
	if err == nil {
		t.Fatalf("expected error when provider_cluster is empty")
	}
}

func TestValidateConnectivityBlock_AgentTunnelOk(t *testing.T) {
	err := validateConnectivityBlock(map[string]interface{}{
		"mode": "agent-tunnel",
		"agent_tunnel": []interface{}{
			map[string]interface{}{"provider_cluster": "hub-cluster"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateConnectivityBlock_PublicMustNotHaveAgentTunnel(t *testing.T) {
	err := validateConnectivityBlock(map[string]interface{}{
		"mode": "public",
		"agent_tunnel": []interface{}{
			map[string]interface{}{"provider_cluster": "hub-cluster"},
		},
	})
	if err == nil {
		t.Fatalf("expected error when agent_tunnel is set with mode=public")
	}
}

func TestValidateAuthMethodBlock_StaticTokenRequiresBlock(t *testing.T) {
	err := validateAuthMethodBlock("static_token", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected missing static_token block error")
	}
}

func TestValidateAuthMethodBlock_OAuth2RequiresBlock(t *testing.T) {
	err := validateAuthMethodBlock("oauth2_client_credentials", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected missing oauth2_client_credentials block error")
	}
}

func TestValidateAuthMethodBlock_TokenExchangeSubjectTokenXOR(t *testing.T) {
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
	err := validateAuthMethodBlock("token_exchange", auth)
	if err == nil {
		t.Fatalf("expected subject_token xor validation error")
	}
}

func TestValidateAuthMethodBlock_TokenExchangeAcceptsValueOnly(t *testing.T) {
	auth := map[string]interface{}{
		"token_exchange": []interface{}{
			map[string]interface{}{
				"subject_token": []interface{}{
					map[string]interface{}{
						"value": "abc",
						"type":  "urn:ietf:params:oauth:token-type:jwt",
					},
				},
			},
		},
	}
	if err := validateAuthMethodBlock("token_exchange", auth); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAuthMethodBlock_CustomRequiresTokenURL(t *testing.T) {
	auth := map[string]interface{}{
		"custom": []interface{}{
			map[string]interface{}{"token_url": ""},
		},
	}
	err := validateAuthMethodBlock("custom", auth)
	if err == nil {
		t.Fatalf("expected missing token_url error")
	}
}

func TestValidateTokenHeaderFormat_KnownPlaceholdersOk(t *testing.T) {
	if err := validateTokenHeaderFormat("{token_type} {access_token}"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTokenHeaderFormat_UnknownPlaceholderRejected(t *testing.T) {
	err := validateTokenHeaderFormat("Bearer {token}")
	if err == nil {
		t.Fatalf("expected unknown placeholder error")
	}
}

// flattenAuth round-trip: redacted secrets returned by the API must NOT leak
// into terraform state.
func TestFlattenAuth_OmitsRedactedSecrets(t *testing.T) {
	auth := flattenAuth(&AuthConfig{
		Method: "token_exchange",
		TokenExchange: &TokenExchangeAuth{
			TokenURL:     "https://auth.example.com/token",
			ClientSecret: redactedSecretPlaceholder,
			SubjectToken: SubjectToken{
				Type:  "urn:ietf:params:oauth:token-type:jwt",
				Value: redactedSecretPlaceholder,
			},
		},
	}, nil)
	if len(auth) != 1 {
		t.Fatalf("expected one auth block, got %d", len(auth))
	}
	te := auth[0]["token_exchange"].([]map[string]interface{})[0]
	if _, ok := te["client_secret"]; ok {
		t.Fatalf("expected client_secret to be omitted when API returns redacted placeholder")
	}
	subj := te["subject_token"].([]map[string]interface{})[0]
	if _, ok := subj["value"]; ok {
		t.Fatalf("expected subject_token.value to be omitted when API returns redacted placeholder")
	}
}

func TestFlattenConnectivity_AgentTunnel(t *testing.T) {
	c := flattenConnectivity(Connectivity{
		Mode:        "agent-tunnel",
		AgentTunnel: &AgentTunnel{ProviderCluster: "hub-1"},
	})
	if len(c) != 1 {
		t.Fatalf("expected one connectivity block")
	}
	if c[0]["mode"] != "agent-tunnel" {
		t.Fatalf("expected mode=agent-tunnel, got %v", c[0]["mode"])
	}
	at := c[0]["agent_tunnel"].([]map[string]interface{})[0]
	if at["provider_cluster"] != "hub-1" {
		t.Fatalf("expected provider_cluster=hub-1, got %v", at["provider_cluster"])
	}
}

func TestFlattenConnectivity_Public(t *testing.T) {
	c := flattenConnectivity(Connectivity{Mode: "public"})
	if c[0]["mode"] != "public" {
		t.Fatalf("expected mode=public, got %v", c[0]["mode"])
	}
	if _, ok := c[0]["agent_tunnel"]; ok {
		t.Fatalf("expected no agent_tunnel block for public mode")
	}
}
