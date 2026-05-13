package komodor

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// tfToAPIRightSizingPolicy converts the TF-layer model into the API request
// body. Optional fields with zero values are sent as nil pointers; bools are
// always sent (the user's value, even when false).
func tfToAPIRightSizingPolicy(tf rightSizingPolicyTFData) RightSizingMultiScopePolicy {
	api := RightSizingMultiScopePolicy{
		Name:                tf.Name,
		Priority:            int32(tf.Priority),
		OptimizationPreset:  tf.OptimizationPreset,
		Percentile:          RightSizingPolicyPercentile(tf.Percentile),
		ApplyProtocol:       tf.ApplyProtocol,
		AllowQoSUpgradeV2:   boolPtr(tf.AllowQoSUpgrade),
		AllowQoSDowngrade:   boolPtr(tf.AllowQoSDowngrade),
		AllowHpaRightSizing: boolPtr(tf.AllowHpaRightSizing),
		Scopes:              tfToAPIScopes(tf.Scopes),
	}
	if tf.ApplyProtocol != rsApplyImmediate {
		api.AllowRestart = boolPtr(tf.AllowRestart)
	}
	if tf.Description != "" {
		api.Description = stringPtr(tf.Description)
	}
	if tf.GuardRails != nil {
		gr := tfToAPIGuardRails(*tf.GuardRails)
		api.GuardRails = &gr
	}
	if len(tf.Tags) > 0 {
		tags := tf.Tags
		api.Tags = &tags
	}
	return api
}

// apiToTFRightSizingPolicy converts an API response into the TF-layer model.
// Pointer fields default to the zero value when nil.
func apiToTFRightSizingPolicy(api RightSizingMultiScopePolicy) rightSizingPolicyTFData {
	tf := rightSizingPolicyTFData{
		Name:                api.Name,
		Description:         stringValue(api.Description),
		Priority:            int(api.Priority),
		Scopes:              apiToTFScopes(api.Scopes),
		ApplyProtocol:       api.ApplyProtocol,
		AllowRestart:        boolValue(api.AllowRestart),
		AllowHpaRightSizing: boolValue(api.AllowHpaRightSizing),
		Percentile:          int(api.Percentile),
		OptimizationPreset:  api.OptimizationPreset,
		AllowQoSUpgrade:     boolValue(api.AllowQoSUpgradeV2),
		AllowQoSDowngrade:   boolValue(api.AllowQoSDowngrade),
		PolicySource:        stringValue(api.PolicySource),
		CreatedBy:           stringValue(api.CreatedBy),
		LastModifiedBy:      stringValue(api.LastModifiedBy),
		CreatedAt:           stringValue(api.CreatedAt),
		UpdatedAt:           stringValue(api.UpdatedAt),
	}
	if api.GuardRails != nil {
		gr := apiToTFGuardRails(*api.GuardRails)
		tf.GuardRails = &gr
	}
	if api.Tags != nil {
		tf.Tags = *api.Tags
	}
	return tf
}

// ---- scopes ----

func tfToAPIScopes(tfScopes []scopeTFData) []PolicyResourceScope {
	if len(tfScopes) == 0 {
		return nil
	}
	out := make([]PolicyResourceScope, 0, len(tfScopes))
	for _, s := range tfScopes {
		out = append(out, tfToAPIScope(s))
	}
	return out
}

func tfToAPIScope(s scopeTFData) PolicyResourceScope {
	scope := PolicyResourceScope{}
	if len(s.Clusters) > 0 {
		v := s.Clusters
		scope.Clusters = &v
	}
	if s.ClustersPatterns != nil {
		p := tfToAPIPattern(*s.ClustersPatterns)
		scope.ClustersPatterns = &p
	}
	if len(s.Namespaces) > 0 {
		v := s.Namespaces
		scope.Namespaces = &v
	}
	if s.NamespacesPatterns != nil {
		p := tfToAPIPattern(*s.NamespacesPatterns)
		scope.NamespacesPatterns = &p
	}
	if len(s.ResourceTypes) > 0 {
		v := s.ResourceTypes
		scope.ResourceTypes = &v
	}
	if s.ResourceTypesPatterns != nil {
		p := tfToAPIPattern(*s.ResourceTypesPatterns)
		scope.ResourceTypesPatterns = &p
	}
	if len(s.WorkloadNames) > 0 {
		v := s.WorkloadNames
		scope.Workloads = &v
	}
	if s.WorkloadNamesPatterns != nil {
		p := tfToAPIPattern(*s.WorkloadNamesPatterns)
		scope.WorkloadsPatterns = &p
	}
	return scope
}

func apiToTFScopes(apiScopes []PolicyResourceScope) []scopeTFData {
	if len(apiScopes) == 0 {
		return nil
	}
	out := make([]scopeTFData, 0, len(apiScopes))
	for _, s := range apiScopes {
		out = append(out, apiToTFScope(s))
	}
	return out
}

func apiToTFScope(s PolicyResourceScope) scopeTFData {
	tf := scopeTFData{}
	if s.Clusters != nil {
		tf.Clusters = *s.Clusters
	}
	if s.ClustersPatterns != nil {
		p := apiToTFPattern(*s.ClustersPatterns)
		tf.ClustersPatterns = &p
	}
	if s.Namespaces != nil {
		tf.Namespaces = *s.Namespaces
	}
	if s.NamespacesPatterns != nil {
		p := apiToTFPattern(*s.NamespacesPatterns)
		tf.NamespacesPatterns = &p
	}
	if s.ResourceTypes != nil {
		tf.ResourceTypes = *s.ResourceTypes
	}
	if s.ResourceTypesPatterns != nil {
		p := apiToTFPattern(*s.ResourceTypesPatterns)
		tf.ResourceTypesPatterns = &p
	}
	if s.Workloads != nil {
		tf.WorkloadNames = *s.Workloads
	}
	if s.WorkloadsPatterns != nil {
		p := apiToTFPattern(*s.WorkloadsPatterns)
		tf.WorkloadNamesPatterns = &p
	}
	return tf
}

func tfToAPIPattern(p patternTFData) PolicyPattern {
	out := PolicyPattern{}
	if p.Include != "" {
		out.Include = stringPtr(p.Include)
	}
	if p.Exclude != "" {
		out.Exclude = stringPtr(p.Exclude)
	}
	return out
}

func apiToTFPattern(p PolicyPattern) patternTFData {
	return patternTFData{
		Include: stringValue(p.Include),
		Exclude: stringValue(p.Exclude),
	}
}

// ---- guardrails ----

func tfToAPIGuardRails(tf guardRailsTFData) PolicyGuardRails {
	gr := PolicyGuardRails{
		ManagedResources: PolicyGuardRailsManagedResources{
			CpuRequests:    tf.ManagedResources.CpuRequests,
			CpuLimits:      tf.ManagedResources.CpuLimits,
			MemoryRequests: tf.ManagedResources.MemoryRequests,
			MemoryLimits:   tf.ManagedResources.MemoryLimits,
		},
		AllowRightSizingUp: tf.AllowRightSizingUp,
		RightSizingConstraints: PolicyGuardRailsConstraints{
			IncreaseCpuBy:    tfToAPIToggleable(tf.Constraints.IncreaseCpuBy),
			DecreaseCpuBy:    tfToAPIToggleable(tf.Constraints.DecreaseCpuBy),
			IncreaseMemoryBy: tfToAPIToggleable(tf.Constraints.IncreaseMemoryBy),
			DecreaseMemoryBy: tfToAPIToggleable(tf.Constraints.DecreaseMemoryBy),
		},
		Buffer: PolicyGuardRailsBuffer{
			Cpu:    tfToAPIToggleable(tf.Buffer.Cpu),
			Memory: tfToAPIToggleable(tf.Buffer.Memory),
		},
	}
	if tf.AbsoluteConstraints != nil {
		ac := tfToAPIAbsoluteConstraints(*tf.AbsoluteConstraints)
		gr.RightSizingAbsoluteConstraints = &ac
	}
	return gr
}

func apiToTFGuardRails(api PolicyGuardRails) guardRailsTFData {
	tf := guardRailsTFData{
		ManagedResources: managedResourcesTFData{
			CpuRequests:    api.ManagedResources.CpuRequests,
			CpuLimits:      api.ManagedResources.CpuLimits,
			MemoryRequests: api.ManagedResources.MemoryRequests,
			MemoryLimits:   api.ManagedResources.MemoryLimits,
		},
		AllowRightSizingUp: api.AllowRightSizingUp,
		Constraints: constraintsTFData{
			IncreaseCpuBy:    apiToTFToggleable(api.RightSizingConstraints.IncreaseCpuBy),
			DecreaseCpuBy:    apiToTFToggleable(api.RightSizingConstraints.DecreaseCpuBy),
			IncreaseMemoryBy: apiToTFToggleable(api.RightSizingConstraints.IncreaseMemoryBy),
			DecreaseMemoryBy: apiToTFToggleable(api.RightSizingConstraints.DecreaseMemoryBy),
		},
		Buffer: bufferTFData{
			Cpu:    apiToTFToggleable(api.Buffer.Cpu),
			Memory: apiToTFToggleable(api.Buffer.Memory),
		},
	}
	if api.RightSizingAbsoluteConstraints != nil {
		ac := apiToTFAbsoluteConstraints(*api.RightSizingAbsoluteConstraints)
		tf.AbsoluteConstraints = &ac
	}
	return tf
}

func tfToAPIAbsoluteConstraints(tf absoluteConstraintsTFData) PolicyGuardRailsAbsoluteConstraints {
	return PolicyGuardRailsAbsoluteConstraints{
		CpuRequestMillicoresMin: tfToAPIToggleable(tf.CpuRequestMillicoresMin),
		CpuRequestMillicoresMax: tfToAPIToggleable(tf.CpuRequestMillicoresMax),
		CpuLimitsMillicoresMin:  tfToAPIToggleable(tf.CpuLimitsMillicoresMin),
		CpuLimitsMillicoresMax:  tfToAPIToggleable(tf.CpuLimitsMillicoresMax),
		MemoryRequestBytesMin:   tfToAPIToggleable(tf.MemoryRequestBytesMin),
		MemoryRequestBytesMax:   tfToAPIToggleable(tf.MemoryRequestBytesMax),
		MemoryLimitsBytesMin:    tfToAPIToggleable(tf.MemoryLimitsBytesMin),
		MemoryLimitsBytesMax:    tfToAPIToggleable(tf.MemoryLimitsBytesMax),
	}
}

func apiToTFAbsoluteConstraints(api PolicyGuardRailsAbsoluteConstraints) absoluteConstraintsTFData {
	return absoluteConstraintsTFData{
		CpuRequestMillicoresMin: apiToTFToggleable(api.CpuRequestMillicoresMin),
		CpuRequestMillicoresMax: apiToTFToggleable(api.CpuRequestMillicoresMax),
		CpuLimitsMillicoresMin:  apiToTFToggleable(api.CpuLimitsMillicoresMin),
		CpuLimitsMillicoresMax:  apiToTFToggleable(api.CpuLimitsMillicoresMax),
		MemoryRequestBytesMin:   apiToTFToggleable(api.MemoryRequestBytesMin),
		MemoryRequestBytesMax:   apiToTFToggleable(api.MemoryRequestBytesMax),
		MemoryLimitsBytesMin:    apiToTFToggleable(api.MemoryLimitsBytesMin),
		MemoryLimitsBytesMax:    apiToTFToggleable(api.MemoryLimitsBytesMax),
	}
}

func tfToAPIToggleable(tf toggleableValueTFData) ToggleableValue {
	return ToggleableValue{Enabled: tf.Enabled, Value: int64(tf.Value)}
}

func apiToTFToggleable(api ToggleableValue) toggleableValueTFData {
	return toggleableValueTFData{Enabled: api.Enabled, Value: int(api.Value)}
}

// ---- pointer helpers ----

func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }

func stringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func boolValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// ============================================================================
// expand: HCL → TF struct (read from *schema.ResourceData)
// ============================================================================

func expandRightSizingPolicy(d *schema.ResourceData) rightSizingPolicyTFData {
	tf := rightSizingPolicyTFData{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Priority:            d.Get("priority").(int),
		Scopes:              expandScopes(d.Get("scope").([]interface{})),
		ApplyProtocol:       d.Get("apply_protocol").(string),
		AllowRestart:        d.Get("allow_restart").(bool),
		AllowHpaRightSizing: d.Get("allow_hpa_right_sizing").(bool),
		Percentile:          d.Get("percentile").(int),
		OptimizationPreset:  d.Get("optimization_preset").(string),
		AllowQoSUpgrade:     d.Get("allow_qos_upgrade").(bool),
		AllowQoSDowngrade:   d.Get("allow_qos_downgrade").(bool),
		Tags:                toStringList(d.Get("tags").([]interface{})),
		ForceDelete:         d.Get("force_delete").(bool),
	}
	if gr := d.Get("guard_rails").([]interface{}); len(gr) > 0 {
		expanded := expandGuardRails(gr[0].(map[string]interface{}))
		tf.GuardRails = &expanded
	}
	return tf
}

func expandScopes(raw []interface{}) []scopeTFData {
	if len(raw) == 0 {
		return nil
	}
	out := make([]scopeTFData, 0, len(raw))
	for _, r := range raw {
		out = append(out, expandScope(r.(map[string]interface{})))
	}
	return out
}

func expandScope(m map[string]interface{}) scopeTFData {
	return scopeTFData{
		Clusters:              toStringList(m["clusters"].([]interface{})),
		ClustersPatterns:      expandPattern(m["clusters_patterns"]),
		Namespaces:            toStringList(m["namespaces"].([]interface{})),
		NamespacesPatterns:    expandPattern(m["namespaces_patterns"]),
		ResourceTypes:         toStringList(m["resource_types"].([]interface{})),
		ResourceTypesPatterns: expandPattern(m["resource_types_patterns"]),
		WorkloadNames:         toStringList(m["workload_names"].([]interface{})),
		WorkloadNamesPatterns: expandPattern(m["workload_names_patterns"]),
	}
}

func expandPattern(v interface{}) *patternTFData {
	raw, ok := v.([]interface{})
	if !ok || len(raw) == 0 {
		return nil
	}
	m := raw[0].(map[string]interface{})
	return &patternTFData{
		Include: m["include"].(string),
		Exclude: stringFromMap(m, "exclude"),
	}
}

func expandGuardRails(m map[string]interface{}) guardRailsTFData {
	gr := guardRailsTFData{
		AllowRightSizingUp: m["allow_right_sizing_up"].(bool),
	}
	if mr := m["managed_resources"].([]interface{}); len(mr) > 0 {
		gr.ManagedResources = expandManagedResources(mr[0].(map[string]interface{}))
	}
	if c := m["constraints"].([]interface{}); len(c) > 0 {
		gr.Constraints = expandConstraints(c[0].(map[string]interface{}))
	}
	if ac := m["absolute_constraints"].([]interface{}); len(ac) > 0 {
		expanded := expandAbsoluteConstraints(ac[0].(map[string]interface{}))
		gr.AbsoluteConstraints = &expanded
	}
	if b := m["buffer"].([]interface{}); len(b) > 0 {
		gr.Buffer = expandBuffer(b[0].(map[string]interface{}))
	}
	return gr
}

func expandManagedResources(m map[string]interface{}) managedResourcesTFData {
	return managedResourcesTFData{
		CpuRequests:    m["cpu_requests"].(bool),
		CpuLimits:      m["cpu_limits"].(bool),
		MemoryRequests: m["memory_requests"].(bool),
		MemoryLimits:   m["memory_limits"].(bool),
	}
}

func expandConstraints(m map[string]interface{}) constraintsTFData {
	return constraintsTFData{
		IncreaseCpuBy:    expandToggleable(m["increase_cpu_by"]),
		DecreaseCpuBy:    expandToggleable(m["decrease_cpu_by"]),
		IncreaseMemoryBy: expandToggleable(m["increase_memory_by"]),
		DecreaseMemoryBy: expandToggleable(m["decrease_memory_by"]),
	}
}

func expandAbsoluteConstraints(m map[string]interface{}) absoluteConstraintsTFData {
	return absoluteConstraintsTFData{
		CpuRequestMillicoresMin: expandToggleable(m["cpu_request_millicores_min"]),
		CpuRequestMillicoresMax: expandToggleable(m["cpu_request_millicores_max"]),
		CpuLimitsMillicoresMin:  expandToggleable(m["cpu_limits_millicores_min"]),
		CpuLimitsMillicoresMax:  expandToggleable(m["cpu_limits_millicores_max"]),
		MemoryRequestBytesMin:   expandToggleable(m["memory_request_bytes_min"]),
		MemoryRequestBytesMax:   expandToggleable(m["memory_request_bytes_max"]),
		MemoryLimitsBytesMin:    expandToggleable(m["memory_limits_bytes_min"]),
		MemoryLimitsBytesMax:    expandToggleable(m["memory_limits_bytes_max"]),
	}
}

func expandBuffer(m map[string]interface{}) bufferTFData {
	return bufferTFData{
		Cpu:    expandToggleable(m["cpu"]),
		Memory: expandToggleable(m["memory"]),
	}
}

func expandToggleable(v interface{}) toggleableValueTFData {
	raw, ok := v.([]interface{})
	if !ok || len(raw) == 0 {
		return toggleableValueTFData{}
	}
	m := raw[0].(map[string]interface{})
	return toggleableValueTFData{
		Enabled: m["enabled"].(bool),
		Value:   m["value"].(int),
	}
}

func stringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// ============================================================================
// flatten: TF struct → HCL (write to *schema.ResourceData)
// Skips force_delete (TFP-only) and id (set via d.SetId).
// ============================================================================

func flattenRightSizingPolicy(d *schema.ResourceData, tf rightSizingPolicyTFData) error {
	for k, v := range map[string]interface{}{
		"name":                   tf.Name,
		"description":            tf.Description,
		"priority":               tf.Priority,
		"scope":                  flattenScopes(tf.Scopes),
		"apply_protocol":         tf.ApplyProtocol,
		"allow_restart":          tf.AllowRestart,
		"allow_hpa_right_sizing": tf.AllowHpaRightSizing,
		"percentile":             tf.Percentile,
		"optimization_preset":    tf.OptimizationPreset,
		"allow_qos_upgrade":      tf.AllowQoSUpgrade,
		"allow_qos_downgrade":    tf.AllowQoSDowngrade,
		"guard_rails":            flattenGuardRails(tf.GuardRails),
		"tags":                   tf.Tags,
		"policy_source":          tf.PolicySource,
		"created_by":             tf.CreatedBy,
		"last_modified_by":       tf.LastModifiedBy,
		"created_at":             tf.CreatedAt,
		"updated_at":             tf.UpdatedAt,
	} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func flattenScopes(scopes []scopeTFData) []interface{} {
	if len(scopes) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(scopes))
	for _, s := range scopes {
		out = append(out, flattenScope(s))
	}
	return out
}

func flattenScope(s scopeTFData) map[string]interface{} {
	return map[string]interface{}{
		"clusters":                s.Clusters,
		"clusters_patterns":       flattenPattern(s.ClustersPatterns),
		"namespaces":              s.Namespaces,
		"namespaces_patterns":     flattenPattern(s.NamespacesPatterns),
		"resource_types":          s.ResourceTypes,
		"resource_types_patterns": flattenPattern(s.ResourceTypesPatterns),
		"workload_names":          s.WorkloadNames,
		"workload_names_patterns": flattenPattern(s.WorkloadNamesPatterns),
	}
}

func flattenPattern(p *patternTFData) []interface{} {
	if p == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"include": p.Include,
			"exclude": p.Exclude,
		},
	}
}

func flattenGuardRails(gr *guardRailsTFData) []interface{} {
	if gr == nil {
		return nil
	}
	m := map[string]interface{}{
		"managed_resources":     flattenManagedResources(gr.ManagedResources),
		"allow_right_sizing_up": gr.AllowRightSizingUp,
		"constraints":           flattenConstraints(gr.Constraints),
		"buffer":                flattenBuffer(gr.Buffer),
	}
	if gr.AbsoluteConstraints != nil {
		m["absolute_constraints"] = flattenAbsoluteConstraints(*gr.AbsoluteConstraints)
	}
	return []interface{}{m}
}

func flattenManagedResources(mr managedResourcesTFData) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"cpu_requests":    mr.CpuRequests,
			"cpu_limits":      mr.CpuLimits,
			"memory_requests": mr.MemoryRequests,
			"memory_limits":   mr.MemoryLimits,
		},
	}
}

func flattenConstraints(c constraintsTFData) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"increase_cpu_by":    flattenToggleable(c.IncreaseCpuBy),
			"decrease_cpu_by":    flattenToggleable(c.DecreaseCpuBy),
			"increase_memory_by": flattenToggleable(c.IncreaseMemoryBy),
			"decrease_memory_by": flattenToggleable(c.DecreaseMemoryBy),
		},
	}
}

func flattenAbsoluteConstraints(ac absoluteConstraintsTFData) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"cpu_request_millicores_min": flattenToggleable(ac.CpuRequestMillicoresMin),
			"cpu_request_millicores_max": flattenToggleable(ac.CpuRequestMillicoresMax),
			"cpu_limits_millicores_min":  flattenToggleable(ac.CpuLimitsMillicoresMin),
			"cpu_limits_millicores_max":  flattenToggleable(ac.CpuLimitsMillicoresMax),
			"memory_request_bytes_min":   flattenToggleable(ac.MemoryRequestBytesMin),
			"memory_request_bytes_max":   flattenToggleable(ac.MemoryRequestBytesMax),
			"memory_limits_bytes_min":    flattenToggleable(ac.MemoryLimitsBytesMin),
			"memory_limits_bytes_max":    flattenToggleable(ac.MemoryLimitsBytesMax),
		},
	}
}

func flattenBuffer(b bufferTFData) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"cpu":    flattenToggleable(b.Cpu),
			"memory": flattenToggleable(b.Memory),
		},
	}
}

func flattenToggleable(t toggleableValueTFData) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"enabled": t.Enabled,
			"value":   t.Value,
		},
	}
}
