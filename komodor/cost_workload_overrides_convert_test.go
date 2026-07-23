package komodor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWorkloadOverridePolicyID(t *testing.T) {
	tests := []struct {
		name      string
		action    string
		policyID  string
		wantError bool
	}{
		{name: "include with policy_id is valid", action: workloadOverrideActionInclude, policyID: "11111111-1111-1111-1111-111111111111", wantError: false},
		{name: "include without policy_id is rejected", action: workloadOverrideActionInclude, policyID: "", wantError: true},
		{name: "exclude without policy_id is valid", action: workloadOverrideActionExclude, policyID: "", wantError: false},
		{name: "exclude with policy_id is valid", action: workloadOverrideActionExclude, policyID: "22222222-2222-2222-2222-222222222222", wantError: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWorkloadOverridePolicyID(tc.action, tc.policyID)
			if tc.wantError {
				assert.ErrorIs(t, err, errPolicyIDRequiredForInclude)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTfToAPIWorkloadOverride_Include(t *testing.T) {
	tf := workloadOverrideTFData{
		Action:      workloadOverrideActionInclude,
		ClusterName: "prod-us-east-1",
		Namespace:   "payments",
		Kind:        "Deployment",
		Name:        "checkout-api",
		PolicyId:    "11111111-1111-1111-1111-111111111111",
	}
	api := tfToAPIWorkloadOverride(tf)
	assert.Equal(t, workloadOverrideActionInclude, api.Action)
	assert.Equal(t, "prod-us-east-1", api.ClusterName)
	assert.Equal(t, "payments", api.Namespace)
	assert.Equal(t, "Deployment", api.Kind)
	assert.Equal(t, "checkout-api", api.Name)
	require.NotNil(t, api.PolicyId)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", *api.PolicyId)
}

func TestTfToAPIWorkloadOverride_ExcludeNilPolicy(t *testing.T) {
	tf := workloadOverrideTFData{
		Action:      workloadOverrideActionExclude,
		ClusterName: "prod-us-east-1",
		Namespace:   "legacy",
		Kind:        "StatefulSet",
		Name:        "old-oracle",
	}
	api := tfToAPIWorkloadOverride(tf)
	assert.Equal(t, workloadOverrideActionExclude, api.Action)
	assert.Nil(t, api.PolicyId)
}

func TestApiToTFWorkloadOverride_RoundTrip(t *testing.T) {
	policyID := "22222222-2222-2222-2222-222222222222"
	createdAt := "2026-01-01T00:00:00Z"
	updatedAt := "2026-01-02T00:00:00Z"
	createdBy := "creator@komodor.io"
	updatedBy := "updater@komodor.io"
	api := WorkloadOverride{
		Id:             "33333333-3333-3333-3333-333333333333",
		Action:         workloadOverrideActionInclude,
		ClusterName:    "prod-us-east-1",
		Namespace:      "payments",
		Kind:           "Deployment",
		Name:           "checkout-api",
		PolicyId:       &policyID,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
		CreatedByEmail: &createdBy,
		LastUpdatedBy:  &updatedBy,
	}

	tf := apiToTFWorkloadOverride(api)
	assert.Equal(t, "33333333-3333-3333-3333-333333333333", tf.Id)
	assert.Equal(t, policyID, tf.PolicyId)
	assert.Equal(t, createdBy, tf.CreatedBy)
	assert.Equal(t, updatedBy, tf.UpdatedBy)
	assert.Equal(t, createdAt, tf.CreatedAt)
	assert.Equal(t, updatedAt, tf.UpdatedAt)

	back := tfToAPIWorkloadOverride(tf)
	assert.Equal(t, api.Action, back.Action)
	assert.Equal(t, api.ClusterName, back.ClusterName)
	assert.Equal(t, api.Namespace, back.Namespace)
	assert.Equal(t, api.Kind, back.Kind)
	assert.Equal(t, api.Name, back.Name)
	require.NotNil(t, back.PolicyId)
	assert.Equal(t, policyID, *back.PolicyId)
}

func TestApiToTFWorkloadOverride_ExcludeNilPolicy(t *testing.T) {
	api := WorkloadOverride{
		Id:          "44444444-4444-4444-4444-444444444444",
		Action:      workloadOverrideActionExclude,
		ClusterName: "prod-us-east-1",
		Namespace:   "legacy",
		Kind:        "StatefulSet",
		Name:        "old-oracle",
	}
	tf := apiToTFWorkloadOverride(api)
	assert.Equal(t, "", tf.PolicyId)

	back := tfToAPIWorkloadOverride(tf)
	assert.Nil(t, back.PolicyId)
}
