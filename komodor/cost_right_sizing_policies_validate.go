package komodor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKomodorCostRightSizingPolicyCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	for _, check := range []func(*schema.ResourceDiff) error{
		applyPresetDefaults,
		validatePresetGuardRailsCombination,
		validateApplyProtocolWithRestart,
		validateScopes,
		validateGuardRailsBlock,
	} {
		if err := check(d); err != nil {
			return err
		}
	}
	return nil
}

func applyPresetDefaults(d *schema.ResourceDiff) error {
	raw := d.GetRawConfig()
	if !raw.IsKnown() || raw.IsNull() {
		return nil
	}
	if v := raw.GetAttr("allow_qos_upgrade"); v.IsKnown() && !v.IsNull() {
		return nil // user set it explicitly — respect that
	}
	switch d.Get("optimization_preset").(string) {
	case rsPresetStaging, rsPresetProduction, rsPresetCustom:
		return d.SetNew("allow_qos_upgrade", true)
	}
	return nil
}

func validatePresetGuardRailsCombination(d *schema.ResourceDiff) error {
	preset := d.Get("optimization_preset").(string)
	hasGuardRails := userProvidedGuardRails(d)

	if preset != rsPresetCustom && hasGuardRails {
		return fmt.Errorf(`optimization_preset = %q cannot be combined with an explicit guardrails block. To override preset values, set optimization_preset = "custom" and provide the full guardrails block`, preset)
	}
	if preset == rsPresetCustom && !hasGuardRails {
		return fmt.Errorf(`guardrails is required when optimization_preset = "custom"`)
	}
	return nil
}

func validateApplyProtocolWithRestart(d *schema.ResourceDiff) error {
	if d.Get("apply_protocol").(string) == rsApplyImmediate && d.Get("allow_restart").(bool) {
		return fmt.Errorf(`allow_restart cannot be true when apply_protocol = "immediate". The "immediate" protocol applies right-sizing in-place and does not require workload restarts`)
	}
	return nil
}

func validateScopes(d *schema.ResourceDiff) error {
	scopes := d.Get("scope").([]interface{})
	for i, raw := range scopes {
		s := raw.(map[string]interface{})
		dims := []struct {
			items, patterns string
			required        bool
		}{
			{"clusters", "clusters_patterns", true},
			{"namespaces", "namespaces_patterns", true},
			{"workload_names", "workload_names_patterns", true},
			{"resource_types", "resource_types_patterns", false},
		}
		for _, dim := range dims {
			if err := validateScopeDimension(i, s, dim.items, dim.patterns, dim.required); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateScopeDimension(idx int, scope map[string]interface{}, itemsKey, patternsKey string, required bool) error {
	items, _ := scope[itemsKey].([]interface{})
	patterns, _ := scope[patternsKey].([]interface{})

	hasItems := len(items) > 0
	hasPatterns := len(patterns) > 0

	if hasItems && hasPatterns {
		return fmt.Errorf(`in scope[%d], %q and %q are mutually exclusive — provide exactly one`, idx, itemsKey, patternsKey)
	}
	if required && !hasItems && !hasPatterns {
		return fmt.Errorf(`in scope[%d], dimension must be set — provide one of %q or %q`, idx, itemsKey, patternsKey)
	}
	return nil
}

func validateGuardRailsBlock(d *schema.ResourceDiff) error {
	if !userProvidedGuardRails(d) {
		return nil
	}
	grBlocks := d.Get("guardrails").([]interface{})
	if len(grBlocks) == 0 {
		return nil
	}
	gr := grBlocks[0].(map[string]interface{})

	if err := validateManagedResources(gr); err != nil {
		return err
	}
	if err := validateAbsoluteConstraints(gr); err != nil {
		return err
	}
	return nil
}

func validateManagedResources(gr map[string]interface{}) error {
	mrBlocks, _ := gr["managed_resources"].([]interface{})
	if len(mrBlocks) == 0 {
		return nil
	}
	mr := mrBlocks[0].(map[string]interface{})
	if !mr["cpu_requests"].(bool) && !mr["cpu_limits"].(bool) && !mr["memory_requests"].(bool) && !mr["memory_limits"].(bool) {
		return fmt.Errorf(`at least one of cpu_requests, cpu_limits, memory_requests, memory_limits must be true in managed_resources`)
	}
	return nil
}

func validateAbsoluteConstraints(gr map[string]interface{}) error {
	acBlocks, _ := gr["absolute_constraints"].([]interface{})
	if len(acBlocks) == 0 {
		return nil
	}
	ac := acBlocks[0].(map[string]interface{})

	pairs := []struct {
		groupName string
		minKey    string
		maxKey    string
	}{
		{"absolute_constraints.cpu_request_millicores", "cpu_request_millicores_min", "cpu_request_millicores_max"},
		{"absolute_constraints.cpu_limits_millicores", "cpu_limits_millicores_min", "cpu_limits_millicores_max"},
		{"absolute_constraints.memory_request_bytes", "memory_request_bytes_min", "memory_request_bytes_max"},
		{"absolute_constraints.memory_limits_bytes", "memory_limits_bytes_min", "memory_limits_bytes_max"},
	}
	for _, p := range pairs {
		minVal, minEnabled := readToggleable(ac[p.minKey])
		maxVal, maxEnabled := readToggleable(ac[p.maxKey])
		if minEnabled && maxEnabled && minVal > maxVal {
			return fmt.Errorf(`in %s, min (%d) must be <= max (%d)`, p.groupName, minVal, maxVal)
		}
	}
	return nil
}

func readToggleable(v interface{}) (int, bool) {
	raw, ok := v.([]interface{})
	if !ok || len(raw) == 0 {
		return 0, false
	}
	m := raw[0].(map[string]interface{})
	return m["value"].(int), m["enabled"].(bool)
}

func userProvidedGuardRails(d *schema.ResourceDiff) bool {
	raw := d.GetRawConfig()
	if !raw.IsKnown() || raw.IsNull() {
		return false
	}
	gr := raw.GetAttr("guardrails")
	if !gr.IsKnown() || gr.IsNull() {
		return false
	}
	return gr.LengthInt() > 0
}

func validateUnsupportedString(field string, allowed []string) schema.SchemaValidateDiagFunc {
	return func(v interface{}, _ cty.Path) diag.Diagnostics {
		val, ok := v.(string)
		if !ok {
			return diag.Errorf("%s: expected a string, got %T", field, v)
		}
		for _, a := range allowed {
			if val == a {
				return nil
			}
		}
		return diag.Errorf("unsupported %s %q — must be one of %s", field, val, formatQuotedStringList(allowed))
	}
}

func validateUnsupportedInt(field string, allowed []int) schema.SchemaValidateDiagFunc {
	return func(v interface{}, _ cty.Path) diag.Diagnostics {
		val, ok := v.(int)
		if !ok {
			return diag.Errorf("%s: expected an int, got %T", field, v)
		}
		for _, a := range allowed {
			if val == a {
				return nil
			}
		}
		return diag.Errorf("unsupported %s %d — must be one of %s", field, val, formatIntList(allowed))
	}
}

func formatQuotedStringList(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func formatIntList(items []int) string {
	parts := make([]string, len(items))
	for i, n := range items {
		parts[i] = strconv.Itoa(n)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
