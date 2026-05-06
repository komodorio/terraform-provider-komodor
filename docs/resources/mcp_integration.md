---
page_title: "komodor_mcp_integration Resource - komodor"
subcategory: ""
description: |-
  Manages a Klaudia MCP integration — connects Klaudia to an external MCP server for AI-powered investigations.
---

# komodor_mcp_integration (Resource)

Manages a Klaudia MCP integration — connects Klaudia to an external MCP server for AI-powered investigations.

An MCP integration connects Klaudia to an external MCP server so that Klaudia can call its tools during AI-powered investigations. Each integration must reference a `komodor_klaudia_skill` via `skill_id`.

**Connectivity modes:**
- `public` — Komodor's control plane calls the MCP server URL directly.
- `agent-tunnel` — A hub agent in the specified `provider_cluster` proxies all traffic to the MCP server.

**Authentication methods:**
- `none` — No authentication.
- `static_token` — A fixed token sent in an HTTP header.
- `oauth2_client_credentials` — OAuth 2.0 client credentials.
- `token_exchange` — RFC 8693 token exchange with optional upstream header forwarding.
- `custom` — POST `auth.custom.body` to `auth.custom.token_url` to obtain a token, then send it to the MCP server using `auth.upstream_header` and `auth.response`.

## Example Usage

```terraform
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
# `token_header` is singular and contains a single template injecting the acquired
# token into the request header(s).
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

    token_header {
      name   = "Authorization"
      format = "{token_type} {access_token}"
    }
  }
}

# Example 4: Token exchange (RFC 8693) with the token injected into a custom header.
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

    token_header {
      name   = "Authorization"
      format = "{token_type} {access_token}"
    }

    response {
      token_field      = "access_token"
      token_type_field = "token_type"
      expires_in_field = "expires_in"
    }
  }
}

# Example 5: Custom auth — POST form body to a token URL, then call MCP with the issued token.
resource "komodor_mcp_integration" "custom_token" {
  name     = "my-custom-auth-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = "https://mcp.example.com/mcp"
    transport = "sse"
    # Optional static (non-auth) headers on every MCP request.
    headers = {
      "X-Client-Name" = "klaudia-terraform"
    }
  }

  auth {
    method = "custom"

    custom {
      # Klaudia POSTs x-www-form-urlencoded to this URL using the body fields below.
      token_url = "https://auth.example.com/issue-token"
      body = {
        grant_type    = "client_credentials"
        client_id     = "my-client-id"
        client_secret = "my-client-secret" # prefer var.sensitive + Vault; values appear in tfstate
      }
    }

    token_header {
      name   = "Authorization"
      format = "{token_type} {access_token}"
    }

    response {
      token_field      = "access_token"
      token_type_field = "token_type"
      expires_in_field = "expires_in"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connectivity` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--connectivity))
- `mcp_server` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--mcp_server))
- `name` (String) Stable machine-safe name for the integration.
- `skill_id` (String) ID of the Klaudia skill to attach. The skill defines instructions and clusters.

### Optional

- `auth` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--connectivity"></a>
### Nested Schema for `connectivity`

Required:

- `mode` (String) How Klaudia reaches the MCP server. `public` — control plane calls directly. `agent-tunnel` — hub agent proxies all traffic.

Optional:

- `agent_tunnel` (Block List, Max: 1) Agent-tunnel options. Required when `mode` is `agent-tunnel`. (see [below for nested schema](#nestedblock--connectivity--agent_tunnel))

<a id="nestedblock--connectivity--agent_tunnel"></a>
### Nested Schema for `connectivity.agent_tunnel`

Required:

- `provider_cluster` (String) Hub cluster that holds credentials and opens the tunnel.



<a id="nestedblock--mcp_server"></a>
### Nested Schema for `mcp_server`

Required:

- `url` (String) MCP server URL.

Optional:

- `headers` (Map of String) Static HTTP headers sent on every MCP request. For static-token auth, do not put the bearer header here — use `auth.static_token` instead. For dynamic auth, do not put token-bearing headers here — use `auth.token_header`.
- `transport` (String) MCP transport protocol: `sse` | `streamable-http`.


<a id="nestedblock--auth"></a>
### Nested Schema for `auth`

Required:

- `method` (String) Authentication method: `static_token` | `oauth2_client_credentials` | `token_exchange` | `custom`.

Optional:

- `custom` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--custom))
- `oauth2_client_credentials` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--oauth2_client_credentials))
- `response` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--response))
- `static_token` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--static_token))
- `token_exchange` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--token_exchange))
- `token_header` (Block List, Max: 1) Header that receives the acquired token, with a templated `format` using `{token_type}` / `{access_token}` placeholders. Not valid when `method = "static_token"`. When omitted, the server defaults to `Authorization: {token_type} {access_token}`. (see [below for nested schema](#nestedblock--auth--token_header))

<a id="nestedblock--auth--custom"></a>
### Nested Schema for `auth.custom`

Required:

- `token_url` (String) Custom token endpoint (POST, form-encoded).

Optional:

- `body` (Map of String, Sensitive) Form fields posted to `token_url`.


<a id="nestedblock--auth--oauth2_client_credentials"></a>
### Nested Schema for `auth.oauth2_client_credentials`

Required:

- `client_id` (String) OAuth2 client ID.
- `client_secret` (String, Sensitive) OAuth2 client secret.
- `token_url` (String) OAuth2 token endpoint URL.

Optional:

- `audience` (String) Target audience.
- `scope` (String) OAuth2 scope.


<a id="nestedblock--auth--response"></a>
### Nested Schema for `auth.response`

Optional:

- `expires_in_field` (String) JSON field containing the TTL in seconds.
- `token_field` (String) JSON field containing the access token.
- `token_type_field` (String) JSON field containing the token type (e.g., `Bearer`).


<a id="nestedblock--auth--static_token"></a>
### Nested Schema for `auth.static_token`

Required:

- `value` (String, Sensitive) Static token value (raw — the server applies the prefix).

Optional:

- `header_name` (String) HTTP header name. The server emits `Bearer <token>` when this is `Authorization`, raw `<token>` otherwise.


<a id="nestedblock--auth--token_exchange"></a>
### Nested Schema for `auth.token_exchange`

Required:

- `subject_token` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--auth--token_exchange--subject_token))
- `token_url` (String) RFC 8693 token exchange endpoint URL.

Optional:

- `actor_token` (String, Sensitive) Actor token, if delegation chain is required.
- `actor_token_type` (String) Actor token type URI.
- `audience` (String) Target audience for the token exchange.
- `client_id` (String) Client ID if the token endpoint requires client authentication.
- `client_secret` (String, Sensitive) Client secret if the token endpoint requires client authentication.
- `extra_params` (Map of String) Additional form parameters to include in the token exchange request.
- `grant_type` (String) OAuth2 grant type. Defaults server-side to `urn:ietf:params:oauth:grant-type:token-exchange`.
- `requested_token_type` (String) Desired response token type.
- `scope` (String) OAuth2 scope.

<a id="nestedblock--auth--token_exchange--subject_token"></a>
### Nested Schema for `auth.token_exchange.subject_token`

Required:

- `type` (String) Subject token type URI (e.g., `urn:ietf:params:oauth:token-type:jwt`).

Optional:

- `file_path` (String) Path to the token file on the agent pod. Mutually exclusive with `value`. Requires `connectivity.mode = "agent-tunnel"`.
- `value` (String, Sensitive) Direct token value. Mutually exclusive with `file_path`.



<a id="nestedblock--auth--token_header"></a>
### Nested Schema for `auth.token_header`

Required:

- `format` (String) Header value template. Allowed placeholders: `{token_type}`, `{access_token}`.

Optional:

- `name` (String) HTTP header name.
