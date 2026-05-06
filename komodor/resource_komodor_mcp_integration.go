package komodor

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var allowedTokenHeaderPlaceholders = map[string]struct{}{
	"{access_token}": {},
	"{token_type}":   {},
}

var tokenHeaderPlaceholderRE = regexp.MustCompile(`\{[^{}]+\}`)

const redactedSecretPlaceholder = "••••••••"

func isRedactedAPIValue(v string) bool {
	return v == redactedSecretPlaceholder
}

func resourceKomodorMCPIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Klaudia MCP integration — connects Klaudia to an external MCP server for AI-powered investigations.",
		CreateContext: resourceMCPIntegrationCreate,
		ReadContext:   resourceMCPIntegrationRead,
		UpdateContext: resourceMCPIntegrationUpdate,
		DeleteContext: resourceMCPIntegrationDelete,
		CustomizeDiff: validateMCPIntegrationDiff,
		Schema: map[string]*schema.Schema{
			// ── Identity ──
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Stable machine-safe name for the integration.",
			},

			// ── Connectivity ──
			"connectivity": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "How Klaudia reaches the MCP server. `public` — control plane calls directly. `agent-tunnel` — hub agent proxies all traffic.",
							ValidateFunc: validation.StringInSlice([]string{"public", "agent-tunnel"}, false),
						},
						"agent_tunnel": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Agent-tunnel options. Required when `mode` is `agent-tunnel`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"provider_cluster": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Hub cluster that holds credentials and opens the tunnel.",
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},

			// ── MCP Server ──
			"mcp_server": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "MCP server URL.",
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"transport": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "sse",
							Description:  "MCP transport protocol: `sse` | `streamable-http`.",
							ValidateFunc: validation.StringInSlice([]string{"sse", "streamable-http"}, false),
						},
						"headers": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "Static HTTP headers sent on every MCP request. " +
								"For static-token auth, do not put the bearer header here — use `auth.static_token` instead. " +
								"For dynamic auth, do not put token-bearing headers here — use `auth.token_header`.",
						},
					},
				},
			},

			// ── Authentication (optional — absence means no auth) ──
			"auth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Authentication method: `static_token` | `oauth2_client_credentials` | `token_exchange` | `custom`.",
							ValidateFunc: validation.StringInSlice([]string{"static_token", "oauth2_client_credentials", "token_exchange", "custom"}, false),
						},

						// --- Static token ---
						"static_token": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "Static token value (raw — the server applies the prefix).",
									},
									"header_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "Authorization",
										Description: "HTTP header name. The server emits `Bearer <token>` when this is `Authorization`, raw `<token>` otherwise.",
									},
								},
							},
						},

						// --- RFC 8693 token exchange ---
						"token_exchange": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token_url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "RFC 8693 token exchange endpoint URL.",
									},
									"grant_type": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "OAuth2 grant type. Defaults server-side to `urn:ietf:params:oauth:grant-type:token-exchange`.",
									},
									"subject_token": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"value": {
													Type:        schema.TypeString,
													Optional:    true,
													Sensitive:   true,
													Description: "Direct token value. Mutually exclusive with `file_path`.",
												},
												"file_path": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Path to the token file on the agent pod. Mutually exclusive with `value`. Requires `connectivity.mode = \"agent-tunnel\"`.",
												},
												"type": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Subject token type URI (e.g., `urn:ietf:params:oauth:token-type:jwt`).",
												},
											},
										},
									},
									"audience": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Target audience for the token exchange.",
									},
									"requested_token_type": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Desired response token type.",
									},
									"scope": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "OAuth2 scope.",
									},
									"actor_token": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "Actor token, if delegation chain is required.",
									},
									"actor_token_type": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Actor token type URI.",
									},
									"client_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Client ID if the token endpoint requires client authentication.",
									},
									"client_secret": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "Client secret if the token endpoint requires client authentication.",
									},
									"extra_params": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Additional form parameters to include in the token exchange request.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},

						// --- OAuth2 client credentials ---
						"oauth2_client_credentials": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token_url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "OAuth2 token endpoint URL.",
									},
									"client_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "OAuth2 client ID.",
									},
									"client_secret": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "OAuth2 client secret.",
									},
									"scope": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "OAuth2 scope.",
									},
									"audience": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Target audience.",
									},
								},
							},
						},

						// --- Custom auth ---
						"custom": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token_url": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Custom token endpoint (POST, form-encoded).",
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"body": {
										Type:        schema.TypeMap,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Sensitive:   true,
										Description: "Form fields posted to `token_url`.",
									},
								},
							},
						},

						// --- Token header (singular, dynamic methods only) ---
						"token_header": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Header that receives the acquired token, with a templated `format` using `{token_type}` / `{access_token}` placeholders. Not valid when `method = \"static_token\"`. When omitted, the server defaults to `Authorization: {token_type} {access_token}`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "Authorization",
										Description: "HTTP header name.",
									},
									"format": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Header value template. Allowed placeholders: `{token_type}`, `{access_token}`.",
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},

						// --- Response field mapping ---
						"response": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token_field": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "access_token",
										Description: "JSON field containing the access token.",
									},
									"token_type_field": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "token_type",
										Description: "JSON field containing the token type (e.g., `Bearer`).",
									},
									"expires_in_field": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "expires_in",
										Description: "JSON field containing the TTL in seconds.",
									},
								},
							},
						},
					},
				},
			},

			// ── Skill (required) ──
			"skill_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Klaudia skill to attach. The skill defines instructions and clusters.",
			},
		},
	}
}

func resourceMCPIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	req := buildMCPRequest(d)
	integration, err := c.CreateMCPIntegration(req)
	if err != nil {
		return diag.Errorf("error creating MCP integration: %s", err)
	}
	d.SetId(integration.ID)
	return resourceMCPIntegrationRead(ctx, d, meta)
}

func resourceMCPIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	integration, statusCode, err := c.GetMCPIntegration(d.Id())
	if err != nil {
		if statusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading MCP integration %s: %s", d.Id(), err)
	}

	_ = d.Set("name", integration.Name)
	skillID := ""
	if integration.SkillID != nil {
		skillID = *integration.SkillID
	}
	_ = d.Set("skill_id", skillID)

	cfg := integration.Configuration
	_ = d.Set("connectivity", flattenConnectivity(cfg.Connectivity))
	_ = d.Set("mcp_server", flattenMCPServer(cfg.MCPServer))
	_ = d.Set("auth", flattenAuth(cfg.Auth))

	return nil
}

func resourceMCPIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	req := buildMCPRequest(d)
	if err := c.UpdateMCPIntegration(d.Id(), req); err != nil {
		return diag.Errorf("error updating MCP integration %s: %s", d.Id(), err)
	}
	return resourceMCPIntegrationRead(ctx, d, meta)
}

func resourceMCPIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	if err := c.DeleteMCPIntegration(d.Id()); err != nil {
		return diag.Errorf("error deleting MCP integration %s: %s", d.Id(), err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// buildMCPRequest — straight 1:1 marshal from the schema blocks. No flattening.
// ─────────────────────────────────────────────────────────────────────────────

func buildMCPRequest(d *schema.ResourceData) *MCPIntegrationRequest {
	mcpServerBlock := d.Get("mcp_server").([]interface{})[0].(map[string]interface{})
	connectivityBlock := d.Get("connectivity").([]interface{})[0].(map[string]interface{})

	cfg := MCPConfiguration{
		MCPServer:    buildMCPServer(mcpServerBlock),
		Connectivity: buildConnectivity(connectivityBlock),
	}

	if authList, ok := d.Get("auth").([]interface{}); ok && len(authList) > 0 {
		if authBlock, ok := authList[0].(map[string]interface{}); ok {
			cfg.Auth = buildAuth(authBlock)
		}
	}

	var skillID *string
	if v := d.Get("skill_id").(string); v != "" {
		skillID = &v
	}

	return &MCPIntegrationRequest{
		Name:          d.Get("name").(string),
		Configuration: cfg,
		UseCases:      []string{"chat", "rca"},
		Clusters:      []string{"*"},
		SkillID:       skillID,
	}
}

func buildMCPServer(block map[string]interface{}) MCPServer {
	return MCPServer{
		URL:       getString(block, "url"),
		Transport: getString(block, "transport"),
		Headers:   stringifyMap(block["headers"]),
	}
}

func buildConnectivity(block map[string]interface{}) Connectivity {
	mode := getString(block, "mode")
	c := Connectivity{Mode: mode}
	if mode == "agent-tunnel" {
		if at, ok := getSingleBlock(block, "agent_tunnel"); ok {
			c.AgentTunnel = &AgentTunnel{
				ProviderCluster: getString(at, "provider_cluster"),
			}
		}
	}
	return c
}

func buildAuth(block map[string]interface{}) *AuthConfig {
	auth := &AuthConfig{Method: getString(block, "method")}

	switch auth.Method {
	case "static_token":
		if st, ok := getSingleBlock(block, "static_token"); ok {
			auth.StaticToken = &StaticTokenAuth{
				Value:      getString(st, "value"),
				HeaderName: getString(st, "header_name"),
			}
		}
	case "token_exchange":
		if te, ok := getSingleBlock(block, "token_exchange"); ok {
			auth.TokenExchange = buildTokenExchange(te)
		}
	case "oauth2_client_credentials":
		if cc, ok := getSingleBlock(block, "oauth2_client_credentials"); ok {
			auth.OAuth2ClientCredentials = &OAuth2Auth{
				TokenURL:     getString(cc, "token_url"),
				ClientID:     getString(cc, "client_id"),
				ClientSecret: getString(cc, "client_secret"),
				Scope:        getString(cc, "scope"),
				Audience:     getString(cc, "audience"),
			}
		}
	case "custom":
		if cu, ok := getSingleBlock(block, "custom"); ok {
			auth.Custom = &CustomAuth{
				TokenURL: strings.TrimSpace(getString(cu, "token_url")),
				Body:     stringifyMap(cu["body"]),
			}
		}
	}

	if th, ok := getSingleBlock(block, "token_header"); ok {
		auth.TokenHeader = &TokenHeader{
			Name:   getString(th, "name"),
			Format: getString(th, "format"),
		}
	}

	if r, ok := getSingleBlock(block, "response"); ok {
		auth.Response = &ResponseConfig{
			TokenField:     getString(r, "token_field"),
			TokenTypeField: getString(r, "token_type_field"),
			ExpiresInField: getString(r, "expires_in_field"),
		}
	}

	return auth
}

func buildTokenExchange(te map[string]interface{}) *TokenExchangeAuth {
	out := &TokenExchangeAuth{
		TokenURL:           getString(te, "token_url"),
		GrantType:          getString(te, "grant_type"),
		Audience:           getString(te, "audience"),
		Scope:              getString(te, "scope"),
		RequestedTokenType: getString(te, "requested_token_type"),
		ActorToken:         getString(te, "actor_token"),
		ActorTokenType:     getString(te, "actor_token_type"),
		ClientID:           getString(te, "client_id"),
		ClientSecret:       getString(te, "client_secret"),
		ExtraParams:        stringifyMap(te["extra_params"]),
	}
	if st, ok := getSingleBlock(te, "subject_token"); ok {
		out.SubjectToken = SubjectToken{
			Type:     getString(st, "type"),
			Value:    getString(st, "value"),
			FilePath: getString(st, "file_path"),
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// flatten* — convert API response back into terraform-state shape (a list of
// maps with the same fields as the schema). Skips redacted secret placeholders
// so they don't leak into state.
// ─────────────────────────────────────────────────────────────────────────────

func flattenConnectivity(c Connectivity) []map[string]interface{} {
	mode := c.Mode
	if mode == "" {
		mode = "public"
	}
	out := map[string]interface{}{"mode": mode}
	if c.AgentTunnel != nil {
		out["agent_tunnel"] = []map[string]interface{}{
			{"provider_cluster": c.AgentTunnel.ProviderCluster},
		}
	}
	return []map[string]interface{}{out}
}

func flattenMCPServer(s MCPServer) []map[string]interface{} {
	out := map[string]interface{}{
		"url":       s.URL,
		"transport": s.Transport,
	}
	if len(s.Headers) > 0 {
		out["headers"] = s.Headers
	}
	return []map[string]interface{}{out}
}

func flattenAuth(a *AuthConfig) []map[string]interface{} {
	if a == nil {
		return nil
	}
	out := map[string]interface{}{"method": a.Method}

	if a.StaticToken != nil {
		st := map[string]interface{}{
			"header_name": a.StaticToken.HeaderName,
		}
		if v := a.StaticToken.Value; v != "" && !isRedactedAPIValue(v) {
			st["value"] = v
		}
		out["static_token"] = []map[string]interface{}{st}
	}
	if a.TokenExchange != nil {
		out["token_exchange"] = flattenTokenExchange(a.TokenExchange)
	}
	if a.OAuth2ClientCredentials != nil {
		cc := map[string]interface{}{
			"token_url": a.OAuth2ClientCredentials.TokenURL,
			"client_id": a.OAuth2ClientCredentials.ClientID,
			"scope":     a.OAuth2ClientCredentials.Scope,
			"audience":  a.OAuth2ClientCredentials.Audience,
		}
		if v := a.OAuth2ClientCredentials.ClientSecret; v != "" && !isRedactedAPIValue(v) {
			cc["client_secret"] = v
		}
		out["oauth2_client_credentials"] = []map[string]interface{}{cc}
	}
	if a.Custom != nil {
		body := map[string]string{}
		for k, v := range a.Custom.Body {
			if !isRedactedAPIValue(v) {
				body[k] = v
			}
		}
		cu := map[string]interface{}{"token_url": a.Custom.TokenURL}
		if len(body) > 0 {
			cu["body"] = body
		}
		out["custom"] = []map[string]interface{}{cu}
	}
	if a.TokenHeader != nil {
		th := map[string]interface{}{
			"name":   a.TokenHeader.Name,
			"format": a.TokenHeader.Format,
		}
		out["token_header"] = []map[string]interface{}{th}
	}
	if a.Response != nil {
		out["response"] = []map[string]interface{}{
			{
				"token_field":      a.Response.TokenField,
				"token_type_field": a.Response.TokenTypeField,
				"expires_in_field": a.Response.ExpiresInField,
			},
		}
	}

	return []map[string]interface{}{out}
}

func flattenTokenExchange(te *TokenExchangeAuth) []map[string]interface{} {
	out := map[string]interface{}{
		"token_url":            te.TokenURL,
		"grant_type":           te.GrantType,
		"audience":             te.Audience,
		"scope":                te.Scope,
		"requested_token_type": te.RequestedTokenType,
		"actor_token_type":     te.ActorTokenType,
		"client_id":            te.ClientID,
	}
	if v := te.ClientSecret; v != "" && !isRedactedAPIValue(v) {
		out["client_secret"] = v
	}
	if v := te.ActorToken; v != "" && !isRedactedAPIValue(v) {
		out["actor_token"] = v
	}
	if len(te.ExtraParams) > 0 {
		out["extra_params"] = te.ExtraParams
	}
	subj := map[string]interface{}{"type": te.SubjectToken.Type}
	if v := te.SubjectToken.Value; v != "" && !isRedactedAPIValue(v) {
		subj["value"] = v
	}
	if v := te.SubjectToken.FilePath; v != "" {
		subj["file_path"] = v
	}
	out["subject_token"] = []map[string]interface{}{subj}
	return []map[string]interface{}{out}
}

// ─────────────────────────────────────────────────────────────────────────────
// Generic helpers
// ─────────────────────────────────────────────────────────────────────────────

func getString(src map[string]interface{}, key string) string {
	raw, ok := src[key]
	if !ok || raw == nil {
		return ""
	}
	if s, ok := raw.(string); ok {
		return s
	}
	return fmt.Sprint(raw)
}

func getSingleBlock(parent map[string]interface{}, key string) (map[string]interface{}, bool) {
	raw, ok := parent[key]
	if !ok {
		return nil, false
	}
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil, false
	}
	block, ok := list[0].(map[string]interface{})
	return block, ok
}

func stringifyMap(raw interface{}) map[string]string {
	switch v := raw.(type) {
	case map[string]string:
		out := make(map[string]string, len(v))
		for k, val := range v {
			out[k] = val
		}
		return out
	case map[string]interface{}:
		out := make(map[string]string, len(v))
		for k, val := range v {
			if val != nil {
				out[k] = fmt.Sprint(val)
			}
		}
		return out
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// validateMCPIntegrationDiff — plan-time checks. The schema enforces required
// fields structurally; this function adds cross-block constraints that the
// schema can't express.
// ─────────────────────────────────────────────────────────────────────────────

func validateMCPIntegrationDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if connectivityList, ok := d.Get("connectivity").([]interface{}); ok && len(connectivityList) > 0 {
		if connectivity, ok := connectivityList[0].(map[string]interface{}); ok {
			if err := validateConnectivityBlock(connectivity); err != nil {
				return err
			}
		}
	}

	auths, ok := d.Get("auth").([]interface{})
	if !ok || len(auths) == 0 {
		return nil
	}
	auth, ok := auths[0].(map[string]interface{})
	if !ok {
		return nil
	}
	method := getString(auth, "method")
	if err := validateAuthMethodBlock(method, auth); err != nil {
		return err
	}

	if th, ok := getSingleBlock(auth, "token_header"); ok {
		if method == "static_token" {
			return fmt.Errorf("auth.token_header is not valid when auth.method is \"static_token\"; use auth.static_token.header_name instead")
		}
		if err := validateTokenHeaderFormat(getString(th, "format")); err != nil {
			return err
		}
	}

	return nil
}

func validateConnectivityBlock(connectivity map[string]interface{}) error {
	mode := getString(connectivity, "mode")
	at, hasTunnel := getSingleBlock(connectivity, "agent_tunnel")
	if mode == "agent-tunnel" {
		if !hasTunnel {
			return fmt.Errorf("connectivity.agent_tunnel block is required when connectivity.mode is \"agent-tunnel\"")
		}
		if strings.TrimSpace(getString(at, "provider_cluster")) == "" {
			return fmt.Errorf("connectivity.agent_tunnel.provider_cluster is required when connectivity.mode is \"agent-tunnel\"")
		}
	} else if hasTunnel {
		return fmt.Errorf("connectivity.agent_tunnel must not be set when connectivity.mode is %q", mode)
	}
	return nil
}

func validateAuthMethodBlock(method string, auth map[string]interface{}) error {
	switch method {
	case "static_token":
		if _, ok := getSingleBlock(auth, "static_token"); !ok {
			return fmt.Errorf("auth.static_token block is required when auth.method is \"static_token\"")
		}
	case "token_exchange":
		te, ok := getSingleBlock(auth, "token_exchange")
		if !ok {
			return fmt.Errorf("auth.token_exchange block is required when auth.method is \"token_exchange\"")
		}
		st, ok := getSingleBlock(te, "subject_token")
		if !ok {
			return fmt.Errorf("auth.token_exchange.subject_token block is required when auth.method is \"token_exchange\"")
		}
		hasValue := strings.TrimSpace(getString(st, "value")) != ""
		hasFilePath := strings.TrimSpace(getString(st, "file_path")) != ""
		if hasValue == hasFilePath {
			return fmt.Errorf("auth.token_exchange.subject_token must set exactly one of `value` or `file_path`")
		}
	case "oauth2_client_credentials":
		if _, ok := getSingleBlock(auth, "oauth2_client_credentials"); !ok {
			return fmt.Errorf("auth.oauth2_client_credentials block is required when auth.method is \"oauth2_client_credentials\"")
		}
	case "custom":
		cu, ok := getSingleBlock(auth, "custom")
		if !ok {
			return fmt.Errorf("auth.custom block is required when auth.method is \"custom\"")
		}
		if strings.TrimSpace(getString(cu, "token_url")) == "" {
			return fmt.Errorf("auth.custom.token_url is required when auth.method is \"custom\"")
		}
	}
	return nil
}

func validateTokenHeaderFormat(template string) error {
	for _, placeholder := range tokenHeaderPlaceholderRE.FindAllString(template, -1) {
		if _, ok := allowedTokenHeaderPlaceholders[placeholder]; !ok {
			return fmt.Errorf("auth.token_header.format: unknown placeholder %q (allowed: {access_token}, {token_type})", placeholder)
		}
	}
	return nil
}
