package komodor

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorCostRightSizingPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Looks up an existing komodor_cost_right_sizing_policy by name. Exposes top-level scalar attributes; the nested scope and guardrails blocks are not surfaced here — manage the policy as a resource if you need them.",
		ReadContext: dataSourceKomodorCostRightSizingPolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Policy name to look up. Resolved via list+iterate (the public API has no get-by-name endpoint).",
			},
			"id":                     {Type: schema.TypeString, Computed: true, Description: "Server-generated unique identifier."},
			"description":            {Type: schema.TypeString, Computed: true, Description: "Free-text description of the policy."},
			"priority":               {Type: schema.TypeInt, Computed: true, Description: "Policy evaluation priority."},
			"optimization_preset":    {Type: schema.TypeString, Computed: true, Description: "Optimization preset."},
			"percentile":             {Type: schema.TypeInt, Computed: true, Description: "Usage percentile."},
			"apply_protocol":         {Type: schema.TypeString, Computed: true, Description: "When right-sizing changes apply."},
			"allow_restart":          {Type: schema.TypeBool, Computed: true, Description: "Whether Komodor may restart pods."},
			"allow_hpa_right_sizing": {Type: schema.TypeBool, Computed: true, Description: "Whether HPA-managed workloads are subject to right-sizing."},
			"allow_qos_upgrade":      {Type: schema.TypeBool, Computed: true, Description: "Allow QoS upgrade."},
			"allow_qos_downgrade":    {Type: schema.TypeBool, Computed: true, Description: "Allow QoS downgrade."},
			"policy_source":          {Type: schema.TypeString, Computed: true, Description: "Source channel that last mutated this policy."},
			"created_by":             {Type: schema.TypeString, Computed: true, Description: "Email of the user who created the policy."},
			"last_modified_by":       {Type: schema.TypeString, Computed: true, Description: "Email of the user who last modified the policy."},
			"created_at":             {Type: schema.TypeString, Computed: true, Description: "Creation timestamp."},
			"updated_at":             {Type: schema.TypeString, Computed: true, Description: "Last-update timestamp."},
		},
	}
}

func dataSourceKomodorCostRightSizingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := newRightSizingClientFromMeta(meta)
	name := d.Get("name").(string)

	resp, status, err := client.GetByName(ctx, name)
	if status == http.StatusNotFound {
		return diag.Errorf("right-sizing policy with name %q not found", name)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	p := resp.Policy
	d.SetId(resp.Id)
	for k, v := range map[string]interface{}{
		"description":            stringValue(p.Description),
		"priority":               int(p.Priority),
		"optimization_preset":    p.OptimizationPreset,
		"percentile":             int(p.Percentile),
		"apply_protocol":         p.ApplyProtocol,
		"allow_restart":          boolValue(p.AllowRestart),
		"allow_hpa_right_sizing": boolValue(p.AllowHpaRightSizing),
		"allow_qos_upgrade":      boolValue(p.AllowQoSUpgradeV2),
		"allow_qos_downgrade":    boolValue(p.AllowQoSDowngrade),
		"policy_source":          stringValue(p.PolicySource),
		"created_by":             stringValue(p.CreatedBy),
		"last_modified_by":       stringValue(p.LastModifiedBy),
		"created_at":             stringValue(p.CreatedAt),
		"updated_at":             stringValue(p.UpdatedAt),
	} {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
