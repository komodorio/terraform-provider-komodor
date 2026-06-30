# First create the skill that the integration will attach to.
resource "komodor_klaudia_skill" "example" {
  name         = "my-mcp-skill"
  description  = "Skill for querying an internal MCP server."
  instructions = "Use the MCP tools to query the internal knowledge base when investigating incidents."
  clusters     = ["*"]
  is_enabled   = true
}

# Example 1: Public connectivity with no authentication.
# Omit the `auth` block entirely for unauthenticated MCP servers.
resource "komodor_mcp_integration" "public_no_auth" {
  name     = "my-public-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = "https://mcp.example.com/mcp"
    transport = "sse"
  }
}

# Example 2: Agent-tunnel connectivity with static token authentication.
# Note: connectivity uses a discriminated nested block (`agent_tunnel { provider_cluster }`).
# The static token value is the raw token — Klaudia wraps it in `Bearer <token>` for the
# `Authorization` header (or in the format you specify).
resource "komodor_mcp_integration" "tunnel_static_token" {
  name     = "my-tunneled-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "agent-tunnel"
    agent_tunnel {
      provider_cluster = "my-hub-cluster"
    }
  }

  mcp_server {
    url       = "http://internal-mcp.svc.cluster.local:8080/mcp"
    transport = "streamable-http"
  }

  auth {
    method = "static_token"

    static_token {
      value       = "my-secret-token"
      header_name = "Authorization"
    }
  }
}

# Example 3: OAuth 2.0 client credentials.
resource "komodor_mcp_integration" "oauth2" {
  name     = "my-oauth2-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = "https://secure-mcp.example.com/mcp"
    transport = "sse"
  }

  auth {
    method = "oauth2_client_credentials"

    oauth2_client_credentials {
      token_url     = "https://auth.example.com/oauth2/token"
      client_id     = "my-client-id"
      client_secret = "my-client-secret"
      scope         = "mcp:read mcp:write"
      audience      = "secure-mcp"
    }
  }
}

# Example 4: Token exchange (RFC 8693).
resource "komodor_mcp_integration" "token_exchange" {
  name     = "my-oauth-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "agent-tunnel"
    agent_tunnel {
      provider_cluster = "my-hub-cluster"
    }
  }

  mcp_server {
    url       = "https://secure-mcp.example.com/mcp"
    transport = "sse"
  }

  auth {
    method = "token_exchange"

    token_exchange {
      token_url  = "https://auth.example.com/token"
      grant_type = "urn:ietf:params:oauth:grant-type:token-exchange"
      audience   = "secure-mcp"

      subject_token {
        # subject_token.file_path requires connectivity.mode = "agent-tunnel"
        # — the agent reads the token from this file inside the cluster.
        file_path = "/var/run/secrets/kubernetes.io/serviceaccount/token"
        type      = "urn:ietf:params:oauth:token-type:jwt"
      }

      requested_token_type = "urn:ietf:params:oauth:token-type:access_token"
    }

    response {
      token_field      = "access_token"
      token_type_field = "token_type"
      expires_in_field = "expires_in"
    }
  }
}
