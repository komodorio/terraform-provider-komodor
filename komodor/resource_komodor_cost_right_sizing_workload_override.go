package komodor

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var errPolicyIDRequiredForInclude = errors.New(`policy_id is required when action = "include"`)

func resourceKomodorCostRightSizingWorkloadOverride() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Komodor cost right-sizing workload override — a per-workload escape hatch on top of the glob-based scope of `komodor_cost_right_sizing_policy`. `include` pins a workload to a specific policy; `exclude` removes it from right-sizing automation entirely.",
		CreateContext: resourceKomodorCostRightSizingWorkloadOverrideCreate,
		ReadContext:   resourceKomodorCostRightSizingWorkloadOverrideRead,
		UpdateContext: resourceKomodorCostRightSizingWorkloadOverrideUpdate,
		DeleteContext: resourceKomodorCostRightSizingWorkloadOverrideDelete,
		CustomizeDiff: resourceKomodorCostRightSizingWorkloadOverrideCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateUnsupportedString("action", workloadOverrideActions),
				Description:      "Override action. One of: `include` (pin the workload to `policy_id`) or `exclude` (remove the workload from right-sizing automation).",
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "Cluster of the target workload. Part of the workload's immutable identity — changing it forces a new resource.",
			},
			"namespace": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "Namespace of the target workload. Part of the workload's immutable identity — changing it forces a new resource.",
			},
			"kind": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "Workload kind (e.g., `Deployment`, `StatefulSet`, `DaemonSet`). Part of the workload's immutable identity — changing it forces a new resource.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "Workload name. Part of the workload's immutable identity — changing it forces a new resource.",
			},
			"policy_id": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsUUID),
				Description:      "Right-sizing policy to pin the workload to. Required when `action = \"include\"`; optional (and typically omitted) when `action = \"exclude\"`.",
			},

			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server-generated unique identifier.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user who created the override.",
			},
			"updated_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user who last modified the override.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last-update timestamp.",
			},
		},
	}
}

func resourceKomodorCostRightSizingWorkloadOverrideCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	return validateWorkloadOverridePolicyID(d.Get("action").(string), d.Get("policy_id").(string))
}

func validateWorkloadOverridePolicyID(action, policyID string) error {
	if action == workloadOverrideActionInclude && policyID == "" {
		return errPolicyIDRequiredForInclude
	}
	return nil
}

func resourceKomodorCostRightSizingWorkloadOverrideCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	api := tfToAPIWorkloadOverride(expandWorkloadOverride(d))

	resp, err := client.CreateOverride(api)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.PolicyScopeExceptionWorkload.Id)

	diags := warningsToDiagnostics(resp.Meta)
	return append(diags, resourceKomodorCostRightSizingWorkloadOverrideRead(ctx, d, meta)...)
}

func resourceKomodorCostRightSizingWorkloadOverrideRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	wl, found, err := client.GetOverrideByID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}
	if err := flattenWorkloadOverride(d, apiToTFWorkloadOverride(*wl)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceKomodorCostRightSizingWorkloadOverrideUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	api := tfToAPIWorkloadOverride(expandWorkloadOverride(d))

	resp, err := client.UpdateOverride(d.Id(), api)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := warningsToDiagnostics(resp.Meta)
	return append(diags, resourceKomodorCostRightSizingWorkloadOverrideRead(ctx, d, meta)...)
}

func resourceKomodorCostRightSizingWorkloadOverrideDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	if err := client.DeleteOverride(d.Id()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func warningsToDiagnostics(meta *WorkloadOverrideItemMeta) diag.Diagnostics {
	if meta == nil || len(meta.Warnings) == 0 {
		return nil
	}
	diags := make(diag.Diagnostics, 0, len(meta.Warnings))
	for _, w := range meta.Warnings {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  w.Code,
			Detail:   w.Message,
		})
	}
	return diags
}
