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
- `static_token` — A fixed bearer token sent in an HTTP header.
- `oauth2_client_credentials` — OAuth 2.0 client credentials flow.
- `token_exchange` — RFC 8693 token exchange with optional upstream header forwarding.
- `custom` — POST `auth.custom.body` to `auth.custom.token_url` to obtain a token, then send it to the MCP server using `auth.upstream_header` and `auth.response`.

## Example Usage

```terraform
# First create the skill that the integration will attach to.
resource "komodor_klaudia_skill" "example" {
  name         = "my-mcp-skill"
  description  = "Skill for querying an internal MCP server."
  instructions = "Use the MCP tools to query the internal knowledge base when investigating incidents."
  use_cases    = ["chat", "rca"]
  clusters     = ["*"]
  is_enabled   = true
}

# Example 1: Public connectivity with no authentication
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

  auth {
    method = "none"
  }
}

# Example 2: Agent-tunnel connectivity with static token authentication
resource "komodor_mcp_integration" "tunnel_static_token" {
  name     = "my-tunneled-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode             = "agent-tunnel"
    provider_cluster = "my-hub-cluster"
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

# Example 3: OAuth 2.0 client credentials
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

    upstream_header {
      name   = "Authorization"
      format = "{token_type} {access_token}"
    }
  }
}

# Example 4: Token exchange (RFC 8693) with upstream header forwarding
resource "komodor_mcp_integration" "token_exchange" {
  name     = "my-oauth-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "public"
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
        file_path = "/var/run/secrets/kubernetes.io/serviceaccount/token"
        type      = "urn:ietf:params:oauth:token-type:jwt"
      }

      requested_token_type = "urn:ietf:params:oauth:token-type:access_token"
    }

    upstream_header {
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

# Example 5: Custom auth — POST form body to a token URL, then call MCP with the issued token
resource "komodor_mcp_integration" "custom_token" {
  name     = "my-custom-auth-mcp"
  skill_id = komodor_klaudia_skill.example.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = "https://mcp.example.com/mcp"
    transport = "sse"
    # Optional static headers on every MCP request.
    headers = {
      "X-Client-Name" = "klaudia-terraform"
    }
  }

  auth {
    method = "custom"

    custom {
      # Required. Klaudia POSTs x-www-form-urlencoded to this URL using all string fields below.
      token_url = "https://auth.example.com/issue-token"
      body = {
        # Any extra string keys are included in the token request.
        grant_type    = "client_credentials"
        client_id     = "my-client-id"
        client_secret = "my-client-secret" # prefer var.sensitive + Vault; values appear in tfstate
      }
    }

    upstream_header {
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

- `auth` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--auth))
- `connectivity` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--connectivity))
- `mcp_server` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--mcp_server))
- `name` (String) Stable machine-safe name for the integration.
- `skill_id` (String) ID of the Klaudia skill to attach. The skill defines instructions, use_cases, and clusters.

### Optional

- `description` (String) Longer description. Markdown supported.
- `display_name` (String) Human-readable name shown in the UI.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--auth"></a>
### Nested Schema for `auth`

Required:

- `method` (String) Authentication method: `none` | `static_token` | `oauth2_client_credentials` | `token_exchange` | `custom`.

Optional:

- `custom` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--custom))
- `oauth2_client_credentials` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--oauth2_client_credentials))
- `response` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--response))
- `static_token` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--static_token))
- `token_exchange` (Block List, Max: 1) (see [below for nested schema](#nestedblock--auth--token_exchange))
- `upstream_header` (Block List) Headers sent to the upstream MCP server. Each entry must set exactly one of `format` (templated) or `value` (literal). Repeat the block for additional headers. (see [below for nested schema](#nestedblock--auth--upstream_header))

<a id="nestedblock--auth--custom"></a>
### Nested Schema for `auth.custom`

Required:

- `token_url` (String) Custom token endpoint (POST, form-encoded). Required when `auth.method` is `custom`.

Optional:

- `body` (Map of String, Sensitive) Form fields merged into Klaudia `auth_params` and sent to `token_url` (same as `CustomTokenProvider` in ai-investigator).


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
- `token_field` (String) JSON field containing the access token in the exchange response.
- `token_type_field` (String) JSON field containing the token type (e.g., `Bearer`).


<a id="nestedblock--auth--static_token"></a>
### Nested Schema for `auth.static_token`

Optional:

- `env_var` (String) Agent environment variable containing the token.
- `header_name` (String) HTTP header name to send the token in.
- `value` (String, Sensitive) Static token value.


<a id="nestedblock--auth--token_exchange"></a>
### Nested Schema for `auth.token_exchange`

Required:

- `subject_token` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--auth--token_exchange--subject_token))
- `token_url` (String) RFC 8693 token exchange endpoint URL.

Optional:

- `actor_token_type` (String) Actor token type URI, if actor token is required.
- `audience` (String) Target audience for the token exchange.
- `client_id` (String) Client ID if the token endpoint requires client authentication.
- `client_secret` (String, Sensitive) Client secret if the token endpoint requires client authentication.
- `extra_params` (Map of String) Additional form parameters to include in the token exchange request.
- `grant_type` (String) OAuth2 grant type.
- `requested_token_type` (String) Desired response token type.
- `scope` (String) OAuth2 scope.

<a id="nestedblock--auth--token_exchange--subject_token"></a>
### Nested Schema for `auth.token_exchange.subject_token`

Required:

- `type` (String) Subject token type URI (e.g., `urn:ietf:params:oauth:token-type:jwt`).

Optional:

- `file_path` (String) Path to the token file on the agent pod. Mutually exclusive with `value`.
- `value` (String, Sensitive) Direct token value. Mutually exclusive with `file_path`.



<a id="nestedblock--auth--upstream_header"></a>
### Nested Schema for `auth.upstream_header`

Required:

- `name` (String) HTTP header name (e.g., `Authorization`).

Optional:

- `format` (String) Header value template. Allowed placeholders: `{token_type}`, `{access_token}`. Mutually exclusive with `value`.
- `value` (String) Literal header value. Mutually exclusive with `format`.



<a id="nestedblock--connectivity"></a>
### Nested Schema for `connectivity`

Required:

- `mode` (String) How Klaudia reaches the MCP server. `public` — control plane calls directly. `agent-tunnel` — hub agent proxies all traffic.

Optional:

- `provider_cluster` (String) Hub cluster that holds credentials and opens the tunnel. Required when mode is `agent-tunnel`.


<a id="nestedblock--mcp_server"></a>
### Nested Schema for `mcp_server`

Required:

- `url` (String) MCP server URL. May reference an agent env var with ${VAR_NAME}.

Optional:

- `headers` (Map of String) Static HTTP header name → value on every MCP request (Klaudia `configuration.headers`). Merged with `auth.static_token` (same header name is overwritten by the bearer value). For dynamic auth, use with non-auth metadata only; use `auth.upstream_header` for token-backed headers.
- `transport` (String) MCP transport protocol: `sse` | `streamable-http`.
