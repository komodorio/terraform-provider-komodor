package komodor

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQosUpgradeV1FromV2(t *testing.T) {
	assert.Equal(t, qosUpgradeBestEffortToBurstable, qosUpgradeV1FromV2(true))
	assert.Equal(t, qosUpgradeNotAllowed, qosUpgradeV1FromV2(false))
}

func TestTfToAPIPattern(t *testing.T) {
	t.Run("legacy singular include/exclude", func(t *testing.T) {
		api := tfToAPIPattern(patternTFData{Include: "prod-*", Exclude: "prod-canary"})
		require.NotNil(t, api.Include)
		assert.Equal(t, "prod-*", *api.Include)
		require.NotNil(t, api.Exclude)
		assert.Equal(t, "prod-canary", *api.Exclude)
		assert.Nil(t, api.Includes)
		assert.Nil(t, api.Excludes)
	})

	t.Run("multi-value includes/excludes", func(t *testing.T) {
		api := tfToAPIPattern(patternTFData{
			Includes: []string{"prod-*", "staging-*"},
			Excludes: []string{"prod-canary"},
		})
		require.NotNil(t, api.Includes)
		assert.Equal(t, []string{"prod-*", "staging-*"}, *api.Includes)
		require.NotNil(t, api.Excludes)
		assert.Equal(t, []string{"prod-canary"}, *api.Excludes)
		assert.Nil(t, api.Include)
		assert.Nil(t, api.Exclude)
	})

	t.Run("empty excludes list is never sent, not sent as an empty slice", func(t *testing.T) {
		api := tfToAPIPattern(patternTFData{Include: "prod-*"})
		assert.Nil(t, api.Excludes, "an unset excludes list must not become a wire value")
	})
}

func TestApiToTFPattern_RoundTrip(t *testing.T) {
	t.Run("legacy singular round-trips", func(t *testing.T) {
		tf := apiToTFPattern(tfToAPIPattern(patternTFData{Include: "prod-*", Exclude: "prod-canary"}))
		assert.Equal(t, "prod-*", tf.Include)
		assert.Equal(t, "prod-canary", tf.Exclude)
		assert.Empty(t, tf.Includes)
		assert.Empty(t, tf.Excludes)
	})

	t.Run("multi-value round-trips", func(t *testing.T) {
		tf := apiToTFPattern(tfToAPIPattern(patternTFData{
			Includes: []string{"prod-*", "staging-*"},
			Excludes: []string{"prod-canary"},
		}))
		assert.Equal(t, []string{"prod-*", "staging-*"}, tf.Includes)
		assert.Equal(t, []string{"prod-canary"}, tf.Excludes)
		assert.Empty(t, tf.Include)
		assert.Empty(t, tf.Exclude)
	})

	t.Run("plural wins if the API ever returns both (validation normally prevents this on write)", func(t *testing.T) {
		tf := apiToTFPattern(PolicyPattern{
			Include:  stringPtr("prod-*"),
			Includes: &[]string{"prod-*", "staging-*"},
			Exclude:  stringPtr("prod-canary"),
			Excludes: &[]string{"prod-canary"},
		})
		assert.Equal(t, []string{"prod-*", "staging-*"}, tf.Includes)
		assert.Equal(t, []string{"prod-canary"}, tf.Excludes)
		assert.Empty(t, tf.Include, "singular must not also flatten into state when plural is present")
		assert.Empty(t, tf.Exclude, "singular must not also flatten into state when plural is present")
	})
}

func TestTfToAPIRightSizingPolicy_Percentiles(t *testing.T) {
	base := func(gr guardRailsTFData) rightSizingPolicyTFData {
		gr.ManagedResources = managedResourcesTFData{CpuRequests: true}
		return rightSizingPolicyTFData{
			Name:               "percentile-test",
			Priority:           1,
			OptimizationPreset: presetCustom,
			ApplyProtocol:      applyOnCreation,
			Scopes:             []scopeTFData{{Clusters: []string{"c"}, Namespaces: []string{"n"}, WorkloadNames: []string{"w"}}},
			GuardRails:         &gr,
		}
	}

	t.Run("split cpu/memory set, shared unset", func(t *testing.T) {
		api := tfToAPIRightSizingPolicy(base(guardRailsTFData{CpuPercentile: 90, MemoryPercentile: 95}))
		assert.Nil(t, api.Percentile)
		require.NotNil(t, api.CpuPercentile)
		assert.Equal(t, RightSizingPolicyPercentile(90), *api.CpuPercentile)
		require.NotNil(t, api.MemoryPercentile)
		assert.Equal(t, RightSizingPolicyPercentile(95), *api.MemoryPercentile)
	})

	t.Run("shared set, split unset", func(t *testing.T) {
		api := tfToAPIRightSizingPolicy(base(guardRailsTFData{Percentile: 80}))
		require.NotNil(t, api.Percentile)
		assert.Equal(t, RightSizingPolicyPercentile(80), *api.Percentile)
		assert.Nil(t, api.CpuPercentile)
		assert.Nil(t, api.MemoryPercentile)
	})

	t.Run("round-trip through apiToTF", func(t *testing.T) {
		api := tfToAPIRightSizingPolicy(base(guardRailsTFData{CpuPercentile: 90, MemoryPercentile: 99}))
		tf := apiToTFRightSizingPolicy(api)
		require.NotNil(t, tf.GuardRails)
		assert.Equal(t, 0, tf.GuardRails.Percentile)
		assert.Equal(t, 90, tf.GuardRails.CpuPercentile)
		assert.Equal(t, 99, tf.GuardRails.MemoryPercentile)
	})
}

func TestTfToAPIRightSizingPolicy_QosFields(t *testing.T) {
	tests := []struct {
		name             string
		upgrade          bool
		downgrade        bool
		wantV1Upgrade    string
		wantV2Upgrade    bool
		wantQosDowngrade bool
	}{
		{name: "upgrade=true downgrade=false", upgrade: true, downgrade: false, wantV1Upgrade: qosUpgradeBestEffortToBurstable, wantV2Upgrade: true, wantQosDowngrade: false},
		{name: "upgrade=false downgrade=true", upgrade: false, downgrade: true, wantV1Upgrade: qosUpgradeNotAllowed, wantV2Upgrade: false, wantQosDowngrade: true},
		{name: "upgrade=false downgrade=false", upgrade: false, downgrade: false, wantV1Upgrade: qosUpgradeNotAllowed, wantV2Upgrade: false, wantQosDowngrade: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tf := rightSizingPolicyTFData{
				Name:               "qos-test",
				Priority:           1,
				OptimizationPreset: presetCustom,
				ApplyProtocol:      applyOnCreation,
				Scopes:             []scopeTFData{{Clusters: []string{"c"}, Namespaces: []string{"n"}, WorkloadNames: []string{"w"}}},
				GuardRails: &guardRailsTFData{
					Percentile:        90,
					AllowQoSUpgrade:   tc.upgrade,
					AllowQoSDowngrade: tc.downgrade,
					ManagedResources:  managedResourcesTFData{CpuRequests: true},
				},
			}
			api := tfToAPIRightSizingPolicy(tf)
			require.NotNil(t, api.AllowQoSUpgrade)
			assert.Equal(t, tc.wantV1Upgrade, *api.AllowQoSUpgrade)
			require.NotNil(t, api.AllowQoSUpgradeV2)
			assert.Equal(t, tc.wantV2Upgrade, *api.AllowQoSUpgradeV2)
			require.NotNil(t, api.AllowQoSDowngrade)
			assert.Equal(t, tc.wantQosDowngrade, *api.AllowQoSDowngrade)
		})
	}
}

func TestTfToAPIExcludePatterns(t *testing.T) {
	tests := []struct {
		name string
		in   *excludePatternsTFData
		want *PolicyExcludePatterns
	}{
		{name: "nil block means feature off", in: nil, want: nil},
		{name: "empty list is never sent", in: &excludePatternsTFData{Excludes: []string{}}, want: nil},
		{name: "nil list is never sent", in: &excludePatternsTFData{}, want: nil},
		{
			name: "values are sent in submitted order",
			in:   &excludePatternsTFData{Excludes: []string{"istio-proxy", "sidecar-*"}},
			want: &PolicyExcludePatterns{Excludes: []string{"istio-proxy", "sidecar-*"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tfToAPIExcludePatterns(tc.in))
		})
	}
}

func TestApiToTFExcludePatterns(t *testing.T) {
	tests := []struct {
		name string
		in   *PolicyExcludePatterns
		want *excludePatternsTFData
	}{
		{name: "absent field", in: nil, want: nil},
		{name: "present but empty is read as off", in: &PolicyExcludePatterns{Excludes: []string{}}, want: nil},
		{
			name: "values keep their order",
			in:   &PolicyExcludePatterns{Excludes: []string{"b", "a"}},
			want: &excludePatternsTFData{Excludes: []string{"b", "a"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, apiToTFExcludePatterns(tc.in))
		})
	}
}

func TestRightSizingPolicy_ContainersPatternsWireFormat(t *testing.T) {
	base := rightSizingPolicyTFData{
		Name:               "wire-test",
		Priority:           1,
		OptimizationPreset: presetSandbox,
		ApplyProtocol:      applyOnCreation,
		Scopes:             []scopeTFData{{Clusters: []string{"c"}, Namespaces: []string{"n"}, WorkloadNames: []string{"w"}}},
	}

	t.Run("without the block the field is absent from the request body", func(t *testing.T) {
		body, err := json.Marshal(tfToAPIRightSizingPolicy(base))
		require.NoError(t, err)
		assert.NotContains(t, string(body), "containersPatterns")
	})

	t.Run("an empty block is also omitted, never sent as {} or an empty list", func(t *testing.T) {
		tf := base
		tf.ContainerNamesPatterns = &excludePatternsTFData{}
		body, err := json.Marshal(tfToAPIRightSizingPolicy(tf))
		require.NoError(t, err)
		assert.NotContains(t, string(body), "containersPatterns")
	})

	t.Run("with the block it is a policy-level excludes-only object", func(t *testing.T) {
		tf := base
		tf.ContainerNamesPatterns = &excludePatternsTFData{Excludes: []string{"istio-proxy", "sidecar-*"}}
		body, err := json.Marshal(tfToAPIRightSizingPolicy(tf))
		require.NoError(t, err)

		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, map[string]interface{}{"excludes": []interface{}{"istio-proxy", "sidecar-*"}}, got["containersPatterns"])
		for _, s := range got["scopes"].([]interface{}) {
			assert.NotContains(t, s, "containersPatterns", "the field must not leak into the scope block")
		}
	})
}

// The critical behaviour: an API policy carrying containersPatterns, read and re-applied
// unchanged, must keep the field (policy update is a full replace).
func TestRightSizingPolicy_ContainersPatternsRoundTrip(t *testing.T) {
	apiJSON := `{
		"name": "rt", "priority": 1, "optimizationPreset": "sandbox", "applyProtocol": "onCreation",
		"containersPatterns": {"excludes": ["istio-proxy", "sidecar-*"]},
		"scopes": [{"clusters": ["c"], "namespaces": ["n"], "workloads": ["w"]}]
	}`
	var fromAPI RightSizingMultiScopePolicy
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &fromAPI))

	// API -> TF data -> flatten into state -> expand from state -> TF data -> API
	res := resourceKomodorCostRightSizingPolicy()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	require.NoError(t, flattenRightSizingPolicy(d, apiToTFRightSizingPolicy(fromAPI)))
	assert.Equal(t, 1, d.Get(containerNamesPatternsAttr+".#"))
	assert.Equal(t, "istio-proxy", d.Get(containerNamesPatternsAttr+".0.excludes.0"))

	resent := tfToAPIRightSizingPolicy(expandRightSizingPolicy(d))
	assert.Equal(t, fromAPI.ContainersPatterns, resent.ContainersPatterns)

	t.Run("an API policy without the field stays without it", func(t *testing.T) {
		var noField RightSizingMultiScopePolicy
		require.NoError(t, json.Unmarshal([]byte(`{"name":"rt","priority":1,"optimizationPreset":"sandbox","applyProtocol":"onCreation","scopes":[]}`), &noField))
		d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
		require.NoError(t, flattenRightSizingPolicy(d, apiToTFRightSizingPolicy(noField)))
		assert.Equal(t, 0, d.Get(containerNamesPatternsAttr+".#"))
		assert.Nil(t, tfToAPIRightSizingPolicy(expandRightSizingPolicy(d)).ContainersPatterns)
	})
}
