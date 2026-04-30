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
	"github.com/samber/lo"
)

var allowedUpstreamHeaderPlaceholders = map[string]struct{}{
	"{access_token}": {},
	"{token_type}":   {},
}

var upstreamHeaderPlaceholderRE = regexp.MustCompile(`\{[^{}]+\}`)

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
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Human-readable name shown in the UI.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Longer description. Markdown supported.",
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
						"provider_cluster": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Hub cluster that holds credentials and opens the tunnel. Required when mode is `agent-tunnel`.",
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
							Description:  "MCP server URL. May reference an agent env var with ${VAR_NAME}.",
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
							Description: "Static HTTP header name → value on every MCP request (Klaudia `configuration.headers`). " +
								"Merged with `auth.static_token` (same header name is overwritten by the bearer value). " +
								"For dynamic auth, use with non-auth metadata only; use `auth.upstream_header` for token-backed headers.",
						},
					},
				},
			},

			// ── Authentication ──
			"auth": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Authentication method: `none` | `static_token` | `oauth2_client_credentials` | `token_exchange` | `custom`.",
							ValidateFunc: validation.StringInSlice([]string{"none", "static_token", "oauth2_client_credentials", "token_exchange", "custom"}, false),
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
										Default:     "urn:ietf:params:oauth:grant-type:token-exchange",
										Description: "OAuth2 grant type.",
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
													Description: "Path to the token file on the agent pod. Mutually exclusive with `value`.",
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
									"actor_token_type": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Actor token type URI, if actor token is required.",
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

						// --- Static token ---
						"static_token": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "Static token value.",
									},
									"env_var": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Agent environment variable containing the token.",
									},
									"header_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "Authorization",
										Description: "HTTP header name to send the token in.",
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

						"custom": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token_url": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Custom token endpoint (POST, form-encoded). Required when `auth.method` is `custom`.",
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"body": {
										Type:        schema.TypeMap,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Sensitive:   true,
										Description: "Form fields merged into Klaudia `auth_params` and sent to `token_url` (same as `CustomTokenProvider` in ai-investigator).",
									},
								},
							},
						},

						// --- Upstream headers ---
						// Repeatable. Each entry is either templated (`format` with
						// `{token_type}`/`{access_token}` placeholders, expanded per
						// request) or static (`value`).
						"upstream_header": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Headers sent to the upstream MCP server. Each entry must set exactly one of `format` (templated) or `value` (literal). Repeat the block for additional headers.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "HTTP header name (e.g., `Authorization`).",
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"format": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Header value template. Allowed placeholders: `{token_type}`, `{access_token}`. Mutually exclusive with `value`.",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Literal header value. Mutually exclusive with `format`.",
									},
								},
							},
						},

						// --- Response parsing ---
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
										Description: "JSON field containing the access token in the exchange response.",
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
				Description: "ID of the Klaudia skill to attach. The skill defines instructions, and clusters.",
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
	connectivity, extractedBearer := flattenConnectivityFromConfig(cfg)
	_ = d.Set("connectivity", []map[string]interface{}{connectivity})

	mcp := map[string]interface{}{
		"url":       cfg["url"],
		"transport": cfg["transport"],
	}
	if h, ok := cfg["headers"]; ok && h != nil {
		if hm := stringifyMap(h); len(hm) > 0 {
			if extractedBearer.headerName != "" {
				delete(hm, extractedBearer.headerName)
			}
			if len(hm) > 0 {
				mcp["headers"] = hm
			}
		}
	}
	_ = d.Set("mcp_server", []map[string]interface{}{mcp})
	auth := flattenAuthFromConfig(d, cfg, extractedBearer)
	_ = d.Set("auth", []map[string]interface{}{auth})

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

// buildMCPRequest maps the CRD-style Terraform hierarchy into the backend's
// flat MCPConfiguration format.
func buildMCPRequest(d *schema.ResourceData) *MCPIntegrationRequest {
	// MCP Server
	mcpServer := d.Get("mcp_server").([]interface{})[0].(map[string]interface{})
	configuration := map[string]interface{}{
		"url":       mcpServer["url"],
		"transport": mcpServer["transport"],
	}
	baseHeaders := mcpServerStringHeaders(mcpServer)

	// Connectivity
	connectivity := d.Get("connectivity").([]interface{})[0].(map[string]interface{})
	if connectivity["mode"].(string) == "agent-tunnel" {
		configuration["use_tunnel"] = true
		if v := connectivity["provider_cluster"].(string); v != "" {
			configuration["tunnel_cluster"] = v
		}
	}

	// Auth
	auth := d.Get("auth").([]interface{})[0].(map[string]interface{})
	method := auth["method"].(string)

	authParams := map[string]interface{}{}

	switch method {
	case "token_exchange":
		configuration["auth_method"] = "rfc8693_token_exchange"
		if te, ok := getSingleBlock(auth, "token_exchange"); ok {
			authParams["token_url"] = te["token_url"].(string)
			setOptionalParam(authParams, te, "grant_type")
			setOptionalParam(authParams, te, "audience")
			setOptionalParam(authParams, te, "requested_token_type")
			setOptionalParam(authParams, te, "scope")
			setOptionalParam(authParams, te, "actor_token_type")
			setOptionalParam(authParams, te, "client_id")
			setOptionalParam(authParams, te, "client_secret")

			if st, ok := getSingleBlock(te, "subject_token"); ok {
				authParams["subject_token_type"] = st["type"].(string)
				if v := st["value"].(string); v != "" {
					authParams["subject_token"] = v
				}
				if v := st["file_path"].(string); v != "" {
					authParams["subject_token_path"] = v
				}
			}

			if extra, ok := te["extra_params"].(map[string]interface{}); ok {
				for k, v := range extra {
					authParams[k] = fmt.Sprintf("%v", v)
				}
			}
		}

	case "oauth2_client_credentials":
		configuration["auth_method"] = "oauth2_client_credentials"
		if cc, ok := getSingleBlock(auth, "oauth2_client_credentials"); ok {
			authParams["token_url"] = cc["token_url"].(string)
			authParams["client_id"] = cc["client_id"].(string)
			authParams["client_secret"] = cc["client_secret"].(string)
			setOptionalParam(authParams, cc, "scope")
			setOptionalParam(authParams, cc, "audience")
		}

	case "static_token":
		configuration["auth_method"] = "static_token"
		headers := map[string]string{}
		for k, v := range baseHeaders {
			headers[k] = v
		}
		if st, ok := getSingleBlock(auth, "static_token"); ok {
			headerName := st["header_name"].(string)
			if headerName == "" {
				headerName = "Authorization"
			}
			if v := st["value"].(string); v != "" {
				headers[headerName] = "Bearer " + v
			}
		}
		if len(headers) > 0 {
			configuration["headers"] = headers
		}

	case "custom":
		configuration["auth_method"] = "custom"
		if cu, ok := getSingleBlock(auth, "custom"); ok {
			if body, ok := cu["body"].(map[string]interface{}); ok {
				for k, v := range body {
					if k == "token_url" {
						continue
					}
					authParams[k] = fmt.Sprintf("%v", v)
				}
			}
			if tu, ok := cu["token_url"].(string); ok {
				authParams["token_url"] = strings.TrimSpace(tu)
			}
		}

	case "none":
		configuration["auth_method"] = "static_token"
	}

	if headers := buildUpstreamHeaders(auth); len(headers) > 0 {
		authParams["upstream_headers"] = headers
	}

	// Response field mapping — nested object in auth_params
	if resp, ok := getSingleBlock(auth, "response"); ok {
		responseMap := map[string]string{}
		if v, _ := resp["token_field"].(string); v != "" {
			responseMap["token_field"] = v
		}
		if v, _ := resp["token_type_field"].(string); v != "" {
			responseMap["token_type_field"] = v
		}
		if v, _ := resp["expires_in_field"].(string); v != "" {
			responseMap["expires_in_field"] = v
		}
		if len(responseMap) > 0 {
			authParams["response"] = responseMap
		}
	}

	if len(authParams) > 0 {
		configuration["auth_params"] = authParams
	}

	// `configuration.headers` (static) for dynamic auth: not merged into static_token branch above
	if method != "static_token" && len(baseHeaders) > 0 {
		if _, has := configuration["headers"]; !has {
			configuration["headers"] = baseHeaders
		}
	}

	// Skill
	var skillID *string
	if v := d.Get("skill_id").(string); v != "" {
		skillID = &v
	}

	// Use cases + clusters come from the skill, but the API still needs them
	// as fallback fields. Set sensible defaults.
	return &MCPIntegrationRequest{
		Name:          d.Get("name").(string),
		Configuration: configuration,
		UseCases:      []string{"chat", "rca"},
		Clusters:      []string{"*"},
		SkillID:       skillID,
	}
}

func mcpServerStringHeaders(mcpServer map[string]interface{}) map[string]string {
	raw, ok := mcpServer["headers"]
	if !ok || raw == nil {
		return nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok || len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if k == "" {
			continue
		}
		out[k] = fmt.Sprint(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

const redactedSecretPlaceholder = "••••••••"

type bearerHeader struct {
	headerName string
	value      string
}

func flattenConnectivityFromConfig(cfg map[string]interface{}) (map[string]interface{}, bearerHeader) {
	connectivity := map[string]interface{}{
		"mode": "public",
	}
	if useTunnel, _ := cfg["use_tunnel"].(bool); useTunnel {
		connectivity["mode"] = "agent-tunnel"
		if cluster, _ := cfg["tunnel_cluster"].(string); cluster != "" {
			connectivity["provider_cluster"] = cluster
		}
	}

	headers := stringifyMap(cfg["headers"])
	bearer := bearerHeader{}
	for name, value := range headers {
		if strings.HasPrefix(value, "Bearer ") {
			bearer.headerName = name
			bearer.value = strings.TrimPrefix(value, "Bearer ")
			break
		}
	}

	return connectivity, bearer
}

func flattenAuthFromConfig(d *schema.ResourceData, cfg map[string]interface{}, bearer bearerHeader) map[string]interface{} {
	auth := map[string]interface{}{}
	authParams := mapFromInterface(cfg["auth_params"])
	method, _ := cfg["auth_method"].(string)
	switch method {
	case "rfc8693_token_exchange":
		auth["method"] = "token_exchange"
		tokenExchange := map[string]interface{}{}
		setStringIfPresent(tokenExchange, "token_url", authParams, "token_url")
		setStringIfPresent(tokenExchange, "grant_type", authParams, "grant_type")
		setStringIfPresent(tokenExchange, "audience", authParams, "audience")
		setStringIfPresent(tokenExchange, "requested_token_type", authParams, "requested_token_type")
		setStringIfPresent(tokenExchange, "scope", authParams, "scope")
		setStringIfPresent(tokenExchange, "actor_token_type", authParams, "actor_token_type")
		setStringIfPresent(tokenExchange, "client_id", authParams, "client_id")
		clientSecret := getString(authParams, "client_secret")
		if clientSecret == redactedSecretPlaceholder {
			clientSecret = getStringFromState(d, "auth.0.token_exchange.0.client_secret")
		}
		if clientSecret != "" {
			tokenExchange["client_secret"] = clientSecret
		}

		subjectToken := map[string]interface{}{}
		setStringIfPresent(subjectToken, "type", authParams, "subject_token_type")
		if st := getString(authParams, "subject_token"); st != "" {
			if st == redactedSecretPlaceholder {
				st = getStringFromState(d, "auth.0.token_exchange.0.subject_token.0.value")
			}
			if st != "" {
				subjectToken["value"] = st
			}
		}
		setStringIfPresent(subjectToken, "file_path", authParams, "subject_token_path")
		if len(subjectToken) > 0 {
			tokenExchange["subject_token"] = []map[string]interface{}{subjectToken}
		}

		extraParams := map[string]string{}
		for key, value := range authParams {
			switch key {
			case "token_url", "grant_type", "audience", "requested_token_type", "scope",
				"actor_token_type", "client_id", "client_secret", "subject_token_type",
				"subject_token", "subject_token_path", "upstream_headers", "response":
				continue
			default:
				if value != nil {
					extraParams[key] = fmt.Sprint(value)
				}
			}
		}
		if len(extraParams) > 0 {
			tokenExchange["extra_params"] = extraParams
		}
		auth["token_exchange"] = []map[string]interface{}{tokenExchange}
	case "oauth2_client_credentials":
		auth["method"] = "oauth2_client_credentials"
		oauth := map[string]interface{}{}
		setStringIfPresent(oauth, "token_url", authParams, "token_url")
		setStringIfPresent(oauth, "client_id", authParams, "client_id")
		clientSecret := getString(authParams, "client_secret")
		if clientSecret == redactedSecretPlaceholder {
			clientSecret = getStringFromState(d, "auth.0.oauth2_client_credentials.0.client_secret")
		}
		if clientSecret != "" {
			oauth["client_secret"] = clientSecret
		}
		setStringIfPresent(oauth, "scope", authParams, "scope")
		setStringIfPresent(oauth, "audience", authParams, "audience")
		auth["oauth2_client_credentials"] = []map[string]interface{}{oauth}
	case "custom":
		auth["method"] = "custom"
		custom := map[string]interface{}{}
		setStringIfPresent(custom, "token_url", authParams, "token_url")
		body := map[string]string{}
		for key, value := range authParams {
			switch key {
			case "token_url", "upstream_headers", "response":
				continue
			default:
				if value != nil {
					body[key] = fmt.Sprint(value)
				}
			}
		}
		if len(body) > 0 {
			custom["body"] = body
		}
		auth["custom"] = []map[string]interface{}{custom}
	case "static_token":
		staticToken := map[string]interface{}{}
		if bearer.headerName != "" {
			staticToken["header_name"] = bearer.headerName
			if bearer.value != "" {
				staticToken["value"] = bearer.value
			}
		}
		if len(staticToken) > 0 {
			auth["method"] = "static_token"
			auth["static_token"] = []map[string]interface{}{staticToken}
		} else {
			auth["method"] = "none"
		}
	default:
		auth["method"] = "none"
	}

	if upstreamHeaders := flattenUpstreamHeaders(authParams["upstream_headers"]); len(upstreamHeaders) > 0 {
		auth["upstream_header"] = upstreamHeaders
	}
	if response := flattenResponseFields(authParams["response"]); len(response) > 0 {
		auth["response"] = []map[string]interface{}{response}
	}

	return auth
}

func setStringIfPresent(dst map[string]interface{}, dstKey string, src map[string]interface{}, srcKey string) {
	if v := getString(src, srcKey); v != "" {
		dst[dstKey] = v
	}
}

func stringifyMap(raw interface{}) map[string]string {
	switch v := raw.(type) {
	case map[string]string:
		out := make(map[string]string, len(v))
		for key, value := range v {
			out[key] = value
		}
		return out
	case map[string]interface{}:
		return lo.MapEntries(v, func(key string, value interface{}) (string, string) {
			return key, fmt.Sprint(value)
		})
	default:
		return map[string]string{}
	}
}

func flattenUpstreamHeaders(raw interface{}) []map[string]interface{} {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil
	}
	return lo.FilterMap(items, func(item interface{}, _ int) (map[string]interface{}, bool) {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, false
		}
		name := getString(m, "name")
		if name == "" {
			return nil, false
		}
		entry := map[string]interface{}{"name": name}
		if format := getString(m, "format"); format != "" {
			entry["format"] = format
		} else if value := getString(m, "value"); value != "" {
			entry["value"] = value
		} else {
			return nil, false
		}
		return entry, true
	})
}

func flattenResponseFields(raw interface{}) map[string]interface{} {
	m := mapFromInterface(raw)
	response := map[string]interface{}{}
	setStringIfPresent(response, "token_field", m, "token_field")
	setStringIfPresent(response, "token_type_field", m, "token_type_field")
	setStringIfPresent(response, "expires_in_field", m, "expires_in_field")
	return response
}

func mapFromInterface(raw interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	if raw == nil {
		return out
	}
	switch v := raw.(type) {
	case map[string]interface{}:
		for key, value := range v {
			out[key] = value
		}
	case map[string]string:
		for key, value := range v {
			out[key] = value
		}
	}
	return out
}

func getString(src map[string]interface{}, key string) string {
	raw, ok := src[key]
	if !ok || raw == nil {
		return ""
	}
	return fmt.Sprint(raw)
}

func getStringFromState(d *schema.ResourceData, path string) string {
	raw := d.Get(path)
	if raw == nil {
		return ""
	}
	s, _ := raw.(string)
	return s
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

func setOptionalParam(params map[string]interface{}, src map[string]interface{}, key string) {
	if v, ok := src[key].(string); ok && v != "" {
		params[key] = v
	}
}

// buildUpstreamHeaders converts the auth.upstream_header[] blocks into the wire
// format expected by the backend: a list under auth_params["upstream_headers"]
// where every entry has a name plus exactly one of format / value.
func buildUpstreamHeaders(auth map[string]interface{}) []map[string]string {
	raw, ok := auth["upstream_header"].([]interface{})
	if !ok || len(raw) == 0 {
		return nil
	}
	return lo.FilterMap(raw, func(item interface{}, _ int) (map[string]string, bool) {
		hdr, ok := item.(map[string]interface{})
		if !ok {
			return nil, false
		}
		name, _ := hdr["name"].(string)
		if name == "" {
			return nil, false
		}
		format, _ := hdr["format"].(string)
		value, _ := hdr["value"].(string)
		entry := map[string]string{"name": name}
		switch {
		case format != "":
			entry["format"] = format
		case value != "":
			entry["value"] = value
		default:
			return nil, false
		}
		return entry, true
	})
}

// validateMCPIntegrationDiff enforces format-xor-value per upstream_header entry
// and rejects unknown placeholders in `format`. Surfacing these at plan time
// gives users actionable diagnostics with the offending block index.
func validateMCPIntegrationDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	connectivityList, ok := d.Get("connectivity").([]interface{})
	if ok && len(connectivityList) > 0 {
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
	method, _ := auth["method"].(string)
	if err := validateAuthMethodBlock(method, auth); err != nil {
		return err
	}

	headers, ok := auth["upstream_header"].([]interface{})
	if !ok {
		return nil
	}
	for i, raw := range headers {
		hdr, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		format, _ := hdr["format"].(string)
		value, _ := hdr["value"].(string)
		hasFormat := format != ""
		hasValue := value != ""
		if hasFormat == hasValue {
			return fmt.Errorf("auth.upstream_header[%d]: must set exactly one of `format` or `value`", i)
		}
		if hasFormat {
			if err := validateUpstreamHeaderFormat(format, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateConnectivityBlock(connectivity map[string]interface{}) error {
	mode, _ := connectivity["mode"].(string)
	if mode != "agent-tunnel" {
		return nil
	}
	providerCluster, _ := connectivity["provider_cluster"].(string)
	if strings.TrimSpace(providerCluster) == "" {
		return fmt.Errorf("connectivity.provider_cluster is required when connectivity.mode is \"agent-tunnel\"")
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
		value, _ := st["value"].(string)
		filePath, _ := st["file_path"].(string)
		hasValue := strings.TrimSpace(value) != ""
		hasFilePath := strings.TrimSpace(filePath) != ""
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
		tu, _ := cu["token_url"].(string)
		if strings.TrimSpace(tu) == "" {
			return fmt.Errorf("auth.custom.token_url is required when auth.method is \"custom\"")
		}
	}
	return nil
}

func validateUpstreamHeaderFormat(template string, index int) error {
	for _, placeholder := range upstreamHeaderPlaceholderRE.FindAllString(template, -1) {
		if _, ok := allowedUpstreamHeaderPlaceholders[placeholder]; !ok {
			return fmt.Errorf(
				"auth.upstream_header[%d].format: unknown placeholder %q (allowed: {access_token}, {token_type})",
				index, placeholder,
			)
		}
	}
	return nil
}
