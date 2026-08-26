package komodor

import (
	"testing"

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
