package komodor

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const DefaultEndpoint = "https://api.komodor.com/mgmt/v1"
const V2Endpoint = "https://api.komodor.com/api/v2"

// KomodorAPIKeyEnvName name of env var for API key
const KomodorAPIKeyEnvName = "KOMODOR_API_KEY"

// KomodorTokenEnvName name of env var for API key
const KomodorTokenEnvName = "KOMODOR_TOKEN"

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
		},

		ResourcesMap: map[string]*schema.Resource{
			"komodor_policy":                 resourceKomodorPolicy(),
			"komodor_role":                   resourceKomodorRole(),
			"komodor_policy_role_attachment": resourcePolicyRoleAttachment(),
			"komodor_monitor":                resourceKomodorMonitor(),
			"komodor_action":                 resourceKomodorCustomK8sAction(),
			"komodor_kubernetes":             resourceKomodorKubernetes(),
			"komodor_user":                   resourceKomodorUser(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"komodor_role":       dataSourceKomodorRole(),
			"komodor_policy":     dataSourceKomodorPolicy(),
			"komodor_kubernetes": dataSourceKomodorKubernetes(),
			"komodor_user":       dataSourceKomodorUser(),
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
	client := NewClient(apiKey)
	return client, nil
}
