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
