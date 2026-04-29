package komodor

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Default US API base URL
const DefaultAPIBaseURL = "https://api.komodor.com"

// KomodorAPIKeyEnvName name of env var for API key
const KomodorAPIKeyEnvName = "KOMODOR_API_KEY"

// KomodorTokenEnvName name of env var for API key
const KomodorTokenEnvName = "KOMODOR_TOKEN"

// KomodorAPIURLEnvName name of env var for API URL
const KomodorAPIURLEnvName = "KOMODOR_API_URL"

// APIKeyEnvVars names of env var for API key
var APIKeyEnvVars = []string{KomodorAPIKeyEnvName, KomodorTokenEnvName}

// Provider returns a schema.Provider for Komodor.
func Provider() *schema.Provider {
	// Some Provider
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc(APIKeyEnvVars, nil),
				Description:  "The API key for operations. Alternatively, can be configured using the `KOMODOR_API_KEY` or `KOMODOR_TOKEN` environment variables.",
				ValidateFunc: validation.StringMatch(regexp.MustCompile("[0-9a-f-]{36}"), "API key must be 36 characters long and only contain characters 0-9 and a-f and '-' character(all lowercased)"),
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(KomodorAPIURLEnvName, DefaultAPIBaseURL),
				Description: "The base URL for the Komodor API. Defaults to `https://api.komodor.com` for US region. For EU region, use `https://api.eu.komodor.com`. Alternatively, can be configured using the `KOMODOR_API_URL` environment variable.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"komodor_policy_v2":              resourceKomodorPolicyV2(),
			"komodor_role":                   resourceKomodorRole(),
			"komodor_policy_role_attachment": resourcePolicyRoleAttachment(),
			"komodor_user_role_binding":      resourceUserRoleBinding(),
			"komodor_monitor":                resourceKomodorMonitor(),
			"komodor_action":                 resourceKomodorCustomK8sAction(),
			"komodor_kubernetes":             resourceKomodorKubernetes(),
			"komodor_workspace":              resourceKomodorWorkspace(),
			"komodor_user":                   resourceKomodorUser(),
			"komodor_klaudia_skill":          resourceKomodorKlaudiaSkill(),
			"komodor_mcp_integration":        resourceKomodorMCPIntegration(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"komodor_role":       dataSourceKomodorRole(),
			"komodor_policy_v2":  dataSourceKomodorPolicyV2(),
			"komodor_kubernetes": dataSourceKomodorKubernetes(),
			"komodor_user":       dataSourceKomodorUser(),
			"komodor_workspace":  dataSourceKomodorWorkspace(),
		},
		ConfigureContextFunc: providerConfigure,
	}

	return provider
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	if apiKey == "" {
		return nil, diag.Errorf("[ERROR] api_key must be set, can't continue")
	}
	apiURL := d.Get("api_url").(string)
	client := NewClient(apiKey, apiURL)
	return client, nil
}
