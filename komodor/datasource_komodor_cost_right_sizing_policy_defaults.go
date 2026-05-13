package komodor

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorCostRightSizingPolicyDefaults() *schema.Resource {
	return &schema.Resource{
		Description: "Returns the default right-sizing policy and per-preset values from the Komodor API (GET /api/v2/cost/right-sizing/policies/defaults). Per-preset percentiles are exposed as Computed scalars; the full structure is also available as a JSON string for jsondecode().",
		ReadContext: dataSourceKomodorCostRightSizingPolicyDefaultsRead,
		Schema: map[string]*schema.Schema{
			"sandbox_percentile":     {Type: schema.TypeInt, Computed: true, Description: "Sandbox preset's default percentile."},
			"development_percentile": {Type: schema.TypeInt, Computed: true, Description: "Development preset's default percentile."},
			"staging_percentile":     {Type: schema.TypeInt, Computed: true, Description: "Staging preset's default percentile."},
			"production_percentile":  {Type: schema.TypeInt, Computed: true, Description: "Production preset's default percentile."},
			"raw_json": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Full defaults payload as a JSON string. Use jsondecode() to access individual fields (e.g., per-preset guardrails).",
			},
		},
	}
}

func dataSourceKomodorCostRightSizingPolicyDefaultsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := newRightSizingClientFromMeta(meta)
	defaults, err := client.GetDefaults(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("komodor_cost_right_sizing_policy_defaults")

	p := defaults.OptimizationPresets
	for k, v := range map[string]interface{}{
		"sandbox_percentile":     int(p.Sandbox.Percentile),
		"development_percentile": int(p.Development.Percentile),
		"staging_percentile":     int(p.Staging.Percentile),
		"production_percentile":  int(p.Production.Percentile),
	} {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

	raw, err := json.Marshal(defaults)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("raw_json", string(raw)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
