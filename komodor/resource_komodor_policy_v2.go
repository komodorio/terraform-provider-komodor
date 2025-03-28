package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"
)

func resourceKomodorPolicyV2() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a Komodor RBAC Policy",
		CreateContext: resourceKomodorPolicyV2Create,
		ReadContext:   resourceKomodorPolicyV2Read,
		UpdateContext: resourceKomodorPolicyV2Update,
		DeleteContext: resourceKomodorPolicyV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"statements": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resources_scope": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"clusters": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"namespaces": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"clusters_patterns":   patternListSchema(),
									"namespaces_patterns": patternListSchema(),
									"selectors": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     selectorSchema(),
									},
									"selectors_patterns": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     selectorPatternSchema(),
									},
								},
							},
						},
					},
				},
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "v2",
				ValidateFunc: validation.StringInSlice([]string{"v2"}, false),
			},
		},
	}
}

func patternListSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"include": {Type: schema.TypeString, Required: true},
				"exclude": {Type: schema.TypeString, Required: true},
			},
		},
	}
}

func selectorSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key":   {Type: schema.TypeString, Required: true},
			"type":  {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"label", "annotation"}, false)},
			"value": {Type: schema.TypeString, Required: true},
		},
	}
}

func selectorPatternSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key":  {Type: schema.TypeString, Required: true},
			"type": {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"label", "annotation"}, false)},
			"value": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include": {Type: schema.TypeString, Required: true},
						"exclude": {Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	}
}

// Expand (from TF -> GO)

func expandPolicy(d *schema.ResourceData) *NewPolicy {
	return &NewPolicy{
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Statements: expandStatements(d.Get("statements").([]interface{})),
	}
}

func expandStatements(list []interface{}) []Statement {
	statements := make([]Statement, 0, len(list))
	for _, item := range list {
		data := item.(map[string]interface{})
		statements = append(statements, Statement{
			Actions:        toStringList(data["actions"].([]interface{})),
			ResourcesScope: expandResourcesScope(data["resources_scope"].([]interface{})),
		})
	}
	return statements
}

func expandResources(list []interface{}) *[]Resource {
	if len(list) == 0 {
		return nil
	}
	result := make([]Resource, 0, len(list))
	for _, item := range list {
		data := item.(map[string]interface{})
		result = append(result, Resource{
			Cluster:          data["cluster"].(string),
			Namespaces:       toStringList(data["namespaces"].([]interface{})),
			NamespacePattern: data["namespace_pattern"].(string),
		})
	}
	return &result
}

func toStringList(raw []interface{}) []string {
	return lo.Map(raw, func(i interface{}, _ int) string {
		return i.(string)
	})
}

func expandResourcesScope(list []interface{}) *ResourcesScope {
	if len(list) == 0 {
		return nil
	}
	data := list[0].(map[string]interface{})

	return &ResourcesScope{
		Clusters:           toStringList(data["clusters"].([]interface{})),
		Namespaces:         toStringList(data["namespaces"].([]interface{})),
		ClustersPatterns:   expandPatterns(data["clusters_patterns"].([]interface{})),
		NamespacesPatterns: expandPatterns(data["namespaces_patterns"].([]interface{})),
		Selectors:          expandSelectors(data["selectors"].([]interface{})),
		SelectorsPatterns:  expandSelectorPatterns(data["selectors_patterns"].([]interface{})),
	}
}

func expandPatterns(list []interface{}) []Pattern {
	return lo.Map(list, func(item interface{}, _ int) Pattern {
		p := item.(map[string]interface{})
		return Pattern{
			Include: p["include"].(string),
			Exclude: p["exclude"].(string),
		}
	})
}

func expandSelectors(list []interface{}) []Selector {
	return lo.Map(list, func(item interface{}, _ int) Selector {
		s := item.(map[string]interface{})
		return Selector{
			Key:   s["key"].(string),
			Type:  SelectorType(s["type"].(string)),
			Value: s["value"].(string),
		}
	})
}

func expandSelectorPatterns(list []interface{}) []SelectorPattern {
	return lo.Map(list, func(item interface{}, _ int) SelectorPattern {
		sp := item.(map[string]interface{})
		valueList := sp["value"].([]interface{})
		value := valueList[0].(map[string]interface{})
		return SelectorPattern{
			Key:  sp["key"].(string),
			Type: SelectorType(sp["type"].(string)),
			Value: Pattern{
				Include: value["include"].(string),
				Exclude: value["exclude"].(string),
			},
		}
	})
}

// END Expand (from TF -> GO)

// Flatten (from GO -> TF)

func flattenPolicy(policy *Policy, d *schema.ResourceData) error {
	d.Set("name", policy.Name)
	d.Set("type", policy.Type)
	d.Set("statements", flattenStatements(policy.Statements))
	return nil
}

func flattenStatements(statements []Statement) []interface{} {
	return lo.Map(statements, func(s Statement, _ int) interface{} {
		m := map[string]interface{}{
			"actions": toInterfaceList(s.Actions),
		}
		if s.ResourcesScope != nil {
			m["resources_scope"] = []interface{}{flattenResourcesScope(s.ResourcesScope)}
		}
		return m
	})
}

func flattenResources(resources []Resource) []interface{} {
	return lo.Map(resources, func(r Resource, _ int) interface{} {
		return map[string]interface{}{
			"cluster":           r.Cluster,
			"namespaces":        toInterfaceList(r.Namespaces),
			"namespace_pattern": r.NamespacePattern,
		}
	})
}

func toInterfaceList(strs []string) []interface{} {
	return lo.Map(strs, func(s string, _ int) interface{} {
		return s
	})
}

func flattenResourcesScope(scope *ResourcesScope) map[string]interface{} {
	return map[string]interface{}{
		"clusters":            toInterfaceList(scope.Clusters),
		"namespaces":          toInterfaceList(scope.Namespaces),
		"clusters_patterns":   flattenPatterns(scope.ClustersPatterns),
		"namespaces_patterns": flattenPatterns(scope.NamespacesPatterns),
		"selectors":           flattenSelectors(scope.Selectors),
		"selectors_patterns":  flattenSelectorPatterns(scope.SelectorsPatterns),
	}
}

func flattenPatterns(patterns []Pattern) []interface{} {
	return lo.Map(patterns, func(p Pattern, _ int) interface{} {
		return map[string]interface{}{
			"include": p.Include,
			"exclude": p.Exclude,
		}
	})
}

func flattenSelectors(selectors []Selector) []interface{} {
	return lo.Map(selectors, func(s Selector, _ int) interface{} {
		return map[string]interface{}{
			"key":   s.Key,
			"type":  string(s.Type),
			"value": s.Value,
		}
	})
}

func flattenSelectorPattern(sp SelectorPattern) interface{} {
	return map[string]interface{}{
		"key":  sp.Key,
		"type": string(sp.Type),
		"value": []interface{}{
			map[string]interface{}{
				"include": sp.Value.Include,
				"exclude": sp.Value.Exclude,
			},
		},
	}
}

func flattenSelectorPatterns(list []SelectorPattern) []interface{} {
	return lo.Map(list, func(sp SelectorPattern, _ int) interface{} {
		return flattenSelectorPattern(sp)
	})
}

// END Flatten (from GO -> TF)

func resourceKomodorPolicyV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	newPolicy := expandPolicy(d)

	policy, err := client.CreatePolicyV2(newPolicy)
	if err != nil {
		return diag.Errorf("Error creating policy: %s", err)
	}

	d.SetId(policy.Id)

	log.Printf("[INFO] Policy created successfully. Policy Id: %s", policy.Id)

	return resourceKomodorPolicyV2Read(ctx, d, meta)
}

func resourceKomodorPolicyV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	policy, statusCode, err := client.GetPolicy(d.Id())
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Policy (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Policy: %s", err)
	}

	if err := flattenPolicy(policy, d); err != nil {
		return diag.Errorf("Error flattening policy: %s", err)
	}

	return nil
}

func resourceKomodorPolicyV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	newPolicy := expandPolicy(d)

	_, err := client.UpdatePolicyV2(d.Id(), newPolicy)
	if err != nil {
		return diag.Errorf("Error updating policy: %s", err)
	}

	log.Printf("[INFO] Policy %s successfully updated", d.Id())
	return resourceKomodorPolicyV2Read(ctx, d, meta)
}

func resourceKomodorPolicyV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Policy: %s", id)
	if err := client.DeletePolicyV2(id); err != nil {
		return diag.Errorf("Error deleting policy: %s", err)
	}

	d.SetId("")
	return nil
}
