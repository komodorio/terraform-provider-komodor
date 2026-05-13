package komodor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var rsTagPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_\-:./]*$`)

const (
	rsPresetSandbox     = "sandbox"
	rsPresetDevelopment = "development"
	rsPresetStaging     = "staging"
	rsPresetProduction  = "production"
	rsPresetCustom      = "custom"

	rsApplyImmediate  = "immediate"
	rsApplyOnCreation = "onCreation"
)

var (
	rsOptimizationPresets = []string{
		rsPresetSandbox, rsPresetDevelopment, rsPresetStaging, rsPresetProduction, rsPresetCustom,
	}
	rsApplyProtocols = []string{rsApplyImmediate, rsApplyOnCreation}

	// presetPercentiles mirrors mono/services/komodor-cost/pkg/endpoints/policies/default_policy.go
	// — keep in sync if upstream preset definitions change.
	presetPercentiles = map[string]RightSizingPolicyPercentile{
		rsPresetSandbox:     N70,
		rsPresetDevelopment: N80,
		rsPresetStaging:     N90,
		rsPresetProduction:  N95,
	}
)

func resourceKomodorCostRightSizingPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Komodor cost right-sizing policy.",
		CreateContext: resourceKomodorCostRightSizingPolicyCreate,
		ReadContext:   resourceKomodorCostRightSizingPolicyRead,
		UpdateContext: resourceKomodorCostRightSizingPolicyUpdate,
		DeleteContext: resourceKomodorCostRightSizingPolicyDelete,
		CustomizeDiff: resourceKomodorCostRightSizingPolicyCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKomodorCostRightSizingPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "Unique policy name within the account.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Free-text description of the policy.",
			},
			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Policy evaluation priority. Higher value wins when multiple policies match the same workload.",
			},

			"scope": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "One or more scope blocks. Multiple scopes are evaluated with OR.",
				Elem:        costRSPScopeResource(),
			},

			"apply_protocol": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateUnsupportedString("apply_protocol", rsApplyProtocols),
				Description:      `When to apply right-sizing changes. One of: "immediate", "onCreation".`,
			},
			"allow_restart": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `Whether Komodor may restart pods to apply right-sizing. Effective only when apply_protocol = "onCreation".`,
			},
			"allow_hpa_right_sizing": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether HPA-managed workloads are subject to right-sizing.",
			},

			"optimization_preset": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateUnsupportedString("optimization_preset", rsOptimizationPresets),
				Description:      `Optimization preset. "custom" requires an explicit guardrails block; named presets (sandbox/development/staging/production) are resolved to guardrail values server-side and exposed as Computed read-only attributes. Updates to a preset's definition do not affect existing policies.`,
			},
			"guardrails": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: `Right-sizing guardrails. Required when optimization_preset = "custom"; must be omitted otherwise (resolved server-side from the named preset and exposed as Computed).`,
				Elem:        costRSPGuardRailsResource(),
			},

			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    20,
				Description: "Optional client-managed tags for categorization. Each tag must be lowercase, start with a letter or digit, and contain only letters, digits, and the characters `_ - : . /`. Max 200 characters per tag; max 20 tags per policy.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 200),
						validation.StringMatch(rsTagPattern, "must be lowercase, start with a letter or digit, and contain only letters, digits, and `_ - : . /`"),
					),
				},
			},

			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, cascade-deletes any active workload overrides on destroy. Has no effect on create/update.",
			},

			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server-generated unique identifier.",
			},
			"policy_source": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The source channel that last mutated this policy. One of: "terraform", "public-api", "webapp-ui".`,
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user who created the policy.",
			},
			"last_modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user who last modified the policy.",
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

func costRSPScopeResource() *schema.Resource {
	stringList := func(desc string) *schema.Schema {
		return &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: desc,
		}
	}
	patternBlock := func(desc string) *schema.Schema {
		return &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        costRSPPatternResource(),
			Description: desc,
		}
	}
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"clusters":                stringList("Exact cluster names. The string `\"*\"` is treated literally — to match all clusters, use `clusters_patterns { include = \"*\" }` instead. Mutually exclusive with clusters_patterns."),
			"clusters_patterns":       patternBlock("Glob pattern for cluster names (`include = \"*\"` matches all). Mutually exclusive with clusters."),
			"namespaces":              stringList("Exact namespace names. The string `\"*\"` is treated literally — to match all namespaces, use `namespaces_patterns { include = \"*\" }` instead. Mutually exclusive with namespaces_patterns."),
			"namespaces_patterns":     patternBlock("Glob pattern for namespace names (`include = \"*\"` matches all). Mutually exclusive with namespaces."),
			"resource_types":          stringList("Workload kinds (e.g., Deployment, StatefulSet). The string `\"*\"` is treated literally — to match all kinds, use `resource_types_patterns { include = \"*\" }` instead. Mutually exclusive with resource_types_patterns."),
			"resource_types_patterns": patternBlock("Glob pattern for workload kinds (`include = \"*\"` matches all). Mutually exclusive with resource_types."),
			"workload_names":          stringList("Exact workload names. The string `\"*\"` is treated literally — to match all workloads, use `workload_names_patterns { include = \"*\" }` instead. Mutually exclusive with workload_names_patterns."),
			"workload_names_patterns": patternBlock("Glob pattern for workload names (`include = \"*\"` matches all). Mutually exclusive with workload_names."),
		},
	}
}

func costRSPPatternResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"include": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Glob pattern for matching (e.g., "prod-*").`,
			},
			"exclude": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional glob pattern to exclude within the include set.",
			},
		},
	}
}

func costRSPGuardRailsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"percentile": {
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validateUnsupportedInt("percentile", validPercentiles),
				Description:      "Usage percentile to base recommendations on. One of: 70, 80, 90, 95, 99.",
			},
			"managed_resources": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Which resource fields right-sizing may modify. At least one must be true.",
				Elem:        costRSPManagedResourcesResource(),
			},
			"allow_right_sizing_up": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether right-sizing may scale resources up.",
			},
			"allow_qos_upgrade": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Allow to Increase QoS (Support reliability). e.g. BestEffort → Burstable → Guarantee.",
			},
			"allow_qos_downgrade": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Allow to Decrease QoS (Support savings). e.g. Guarantee → Burstable.",
			},
			"constraints": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Per-cycle scaling constraints expressed as percentages.",
				Elem:        costRSPConstraintsResource(),
			},
			"absolute_constraints": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Absolute floor/ceiling values for CPU (millicores) and memory (bytes).",
				Elem:        costRSPAbsoluteConstraintsResource(),
			},
			"buffer": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Headroom percentage on top of recommended request values.",
				Elem:        costRSPBufferResource(),
			},
		},
	}
}

func costRSPManagedResourcesResource() *schema.Resource {
	flag := func(desc string) *schema.Schema {
		return &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: desc,
		}
	}
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cpu_requests":    flag("Manage CPU requests."),
			"cpu_limits":      flag("Manage CPU limits."),
			"memory_requests": flag("Manage memory requests."),
			"memory_limits":   flag("Manage memory limits."),
		},
	}
}

func costRSPConstraintsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"increase_cpu_by":    costRSPPercentageToggleableSchema("Max percent to increase CPU per cycle."),
			"decrease_cpu_by":    costRSPPercentageToggleableSchema("Max percent to decrease CPU per cycle."),
			"increase_memory_by": costRSPPercentageToggleableSchema("Max percent to increase memory per cycle."),
			"decrease_memory_by": costRSPPercentageToggleableSchema("Max percent to decrease memory per cycle."),
		},
	}
}

func costRSPAbsoluteConstraintsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cpu_request_millicores_min": costRSPAbsoluteToggleableSchema("Floor for CPU request in millicores."),
			"cpu_request_millicores_max": costRSPAbsoluteToggleableSchema("Ceiling for CPU request in millicores."),
			"cpu_limits_millicores_min":  costRSPAbsoluteToggleableSchema("Floor for CPU limits in millicores."),
			"cpu_limits_millicores_max":  costRSPAbsoluteToggleableSchema("Ceiling for CPU limits in millicores."),
			"memory_request_bytes_min":   costRSPAbsoluteToggleableSchema("Floor for memory request in bytes."),
			"memory_request_bytes_max":   costRSPAbsoluteToggleableSchema("Ceiling for memory request in bytes."),
			"memory_limits_bytes_min":    costRSPAbsoluteToggleableSchema("Floor for memory limits in bytes."),
			"memory_limits_bytes_max":    costRSPAbsoluteToggleableSchema("Ceiling for memory limits in bytes."),
		},
	}
}

func costRSPBufferResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cpu":    costRSPPercentageToggleableSchema("CPU buffer percentage."),
			"memory": costRSPPercentageToggleableSchema("Memory buffer percentage."),
		},
	}
}

func costRSPPercentageToggleableSchema(desc string) *schema.Schema {
	return costRSPToggleableSchemaWith(desc, validation.ToDiagFunc(validation.IntBetween(0, 100)))
}

func costRSPAbsoluteToggleableSchema(desc string) *schema.Schema {
	return costRSPToggleableSchemaWith(desc, validation.ToDiagFunc(validation.IntAtLeast(0)))
}

func costRSPToggleableSchemaWith(desc string, valueValidator schema.SchemaValidateDiagFunc) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: desc,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Whether the value is applied.",
				},
				"value": {
					Type:             schema.TypeInt,
					Optional:         true,
					Default:          0,
					ValidateDiagFunc: valueValidator,
					Description:      "The numeric value (effective only when enabled).",
				},
			},
		},
	}
}

func resourceKomodorCostRightSizingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := newRightSizingClientFromMeta(meta)
	api := tfToAPIRightSizingPolicy(expandRightSizingPolicy(d))

	resp, err := client.Create(ctx, api)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Id)
	return resourceKomodorCostRightSizingPolicyRead(ctx, d, meta)
}

func resourceKomodorCostRightSizingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := newRightSizingClientFromMeta(meta)
	resp, status, err := client.GetByID(ctx, d.Id())
	if status == http.StatusNotFound {
		log.Printf("[DEBUG] right-sizing policy (%s) not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	tf := apiToTFRightSizingPolicy(resp.Policy)
	if err := flattenRightSizingPolicy(d, tf); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceKomodorCostRightSizingPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := newRightSizingClientFromMeta(meta)
	api := tfToAPIRightSizingPolicy(expandRightSizingPolicy(d))

	if _, err := client.Update(ctx, d.Id(), api); err != nil {
		return diag.FromErr(err)
	}
	return resourceKomodorCostRightSizingPolicyRead(ctx, d, meta)
}

func resourceKomodorCostRightSizingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := newRightSizingClientFromMeta(meta)
	force := d.Get("force_delete").(bool)
	if err := client.Delete(ctx, d.Id(), force); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func newRightSizingClientFromMeta(meta interface{}) *rightSizingPoliciesClient {
	c := meta.(*Client)
	return newRightSizingPoliciesClient(newRightSizingHTTP(c.BaseURL, c.ApiKey))
}

func resourceKomodorCostRightSizingPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	identifier := d.Id()
	client := newRightSizingClientFromMeta(meta)

	if _, _, err := client.GetByID(ctx, identifier); err == nil {
		return []*schema.ResourceData{d}, nil
	}

	resp, status, err := client.GetByName(ctx, identifier)
	if status == http.StatusNotFound {
		return nil, fmt.Errorf("right-sizing policy %q not found (tried as ID and name)", identifier)
	}
	if err != nil {
		return nil, fmt.Errorf("import right-sizing policy %q: %w", identifier, err)
	}
	d.SetId(resp.Id)
	return []*schema.ResourceData{d}, nil
}
