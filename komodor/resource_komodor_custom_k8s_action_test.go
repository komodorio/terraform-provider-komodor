package komodor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRuleset(t *testing.T) {
	tests := []struct {
		name     string
		ruleset  string
		expected []CustomK8sActionStatement
		wantErr  bool
	}{
		{
			name:    "resource rule",
			ruleset: `[{"apiGroups":[""],"resources":["pods"],"verbs":["get","list"]}]`,
			expected: []CustomK8sActionStatement{
				{ApiGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get", "list"}},
			},
		},
		{
			name:    "nonResourceURLs is kept",
			ruleset: `[{"nonResourceURLs":["/metrics"],"verbs":["get"]}]`,
			expected: []CustomK8sActionStatement{
				{NonResourceURLs: []string{"/metrics"}, Verbs: []string{"get"}},
			},
		},
		{
			name:    "misspelled key is rejected",
			ruleset: `[{"apiGroup":[""],"resources":["pods"],"verbs":["get"]}]`,
			wantErr: true,
		},
		{
			name:    "resourceNames is rejected until the API supports it",
			ruleset: `[{"apiGroups":[""],"resources":["configmaps"],"resourceNames":["my-cm"],"verbs":["get"]}]`,
			wantErr: true,
		},
		{
			name:    "trailing data is rejected",
			ruleset: `[{"verbs":["get"],"nonResourceURLs":["/a"]}][{"verbs":["delete"],"nonResourceURLs":["/b"]}]`,
			wantErr: true,
		},
		{
			name:    "rule without apiGroups and resources is rejected",
			ruleset: `[{"verbs":["get"]}]`,
			wantErr: true,
		},
		{
			name:    "rule with resources but no apiGroups is rejected",
			ruleset: `[{"resources":["pods"],"verbs":["get"]}]`,
			wantErr: true,
		},
		{
			name:    "rule mixing resources and nonResourceURLs is rejected",
			ruleset: `[{"apiGroups":[""],"resources":["pods"],"nonResourceURLs":["/metrics"],"verbs":["get"]}]`,
			wantErr: true,
		},
		{
			name:    "rule without verbs is rejected",
			ruleset: `[{"apiGroups":[""],"resources":["pods"]}]`,
			wantErr: true,
		},
		{
			name:    "invalid JSON is rejected",
			ruleset: `[{"verbs":["get"]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRuleset(tt.ruleset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCustomK8sActionStatementJSONMatchesJsonencode(t *testing.T) {
	// jsonencode() sorts keys and the API omits empty fields, so this is what state must hold to avoid a diff.
	statements := []CustomK8sActionStatement{
		{ApiGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
		{NonResourceURLs: []string{"/metrics"}, Verbs: []string{"get"}},
	}
	out, err := json.Marshal(statements)
	require.NoError(t, err)
	assert.Equal(t, `[{"apiGroups":[""],"resources":["pods"],"verbs":["get"]},{"nonResourceURLs":["/metrics"],"verbs":["get"]}]`, string(out))
}

func TestResourceKomodorCustomK8sActionUpdateSendsFullRules(t *testing.T) {
	ruleset := `[{"nonResourceURLs":["/metrics"],"verbs":["get"]}]`
	var putBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			putBody, _ = io.ReadAll(r.Body)
		}
		_, _ = w.Write([]byte(`{"id":"x","action":"a","description":"d","k8sRuleset":` + ruleset + `}`))
	}))
	defer srv.Close()

	d := schema.TestResourceDataRaw(t, resourceKomodorCustomK8sAction().Schema, map[string]interface{}{
		"action":      "a",
		"description": "d",
		"ruleset":     ruleset,
	})
	d.SetId("x")

	diags := resourceKomodorCustomK8sActionUpdate(context.Background(), d, NewClient("key", srv.URL))
	require.False(t, diags.HasError(), "%v", diags)
	assert.JSONEq(t, `{"description":"d","k8sRuleset":[{"nonResourceURLs":["/metrics"],"verbs":["get"]}]}`, string(putBody))
	assert.Equal(t, ruleset, d.Get("ruleset"))
}
