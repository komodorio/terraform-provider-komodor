package komodor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorMCPIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Klaudia MCP integration — connects Klaudia to an external MCP server for AI-powered investigations.",
		CreateContext: resourceMCPIntegrationCreate,
		ReadContext:   resourceMCPIntegrationRead,
		UpdateContext: resourceMCPIntegrationUpdate,
		DeleteContext: resourceMCPIntegrationDelete,
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

						// --- Upstream header ---
						"upstream_header": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "HTTP header name (e.g., `Authorization`).",
									},
									"format": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Header value template. Use `{token_type}` and `{access_token}` as placeholders.",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Static header value (no placeholders).",
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
				Description: "ID of the Klaudia skill to attach. The skill defines instructions, use_cases, and clusters.",
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
	_ = d.Set("mcp_server", []map[string]interface{}{
		{"url": cfg["url"], "transport": cfg["transport"]},
	})

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
		if st, ok := getSingleBlock(auth, "static_token"); ok {
			headerName := st["header_name"].(string)
			if headerName == "" {
				headerName = "Authorization"
			}
			if v := st["value"].(string); v != "" {
				headers := map[string]string{headerName: "Bearer " + v}
				configuration["headers"] = headers
			}
		}

	case "custom":
		configuration["auth_method"] = "custom"

	case "none":
		configuration["auth_method"] = "static_token"
	}

	if len(authParams) > 0 {
		configuration["auth_params"] = authParams
	}

	// Upstream header — nested object in auth_params
	if headers, ok := auth["upstream_header"].([]interface{}); ok && len(headers) > 0 {
		hdr := headers[0].(map[string]interface{})
		upstreamHeader := map[string]string{}
		if v, _ := hdr["name"].(string); v != "" {
			upstreamHeader["name"] = v
		}
		if v, _ := hdr["format"].(string); v != "" {
			upstreamHeader["format"] = v
		}
		if len(upstreamHeader) > 0 {
			authParams["upstream_header"] = upstreamHeader
		}
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
