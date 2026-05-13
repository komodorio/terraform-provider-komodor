package komodor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorCostRightSizingPolicyPresets() *schema.Resource {
	return &schema.Resource{
		Description: "Returns the list of supported optimization preset names for komodor_cost_right_sizing_policy.",
		ReadContext: dataSourceKomodorCostRightSizingPolicyPresetsRead,
		Schema: map[string]*schema.Schema{
			"names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `Supported values for optimization_preset (includes "custom").`,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceKomodorCostRightSizingPolicyPresetsRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("komodor_cost_right_sizing_policy_presets")
	if err := d.Set("names", rsOptimizationPresets); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
