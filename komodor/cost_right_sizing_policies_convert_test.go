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
