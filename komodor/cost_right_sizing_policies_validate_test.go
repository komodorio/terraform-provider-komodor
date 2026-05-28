package komodor

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestWarnLiteralStarInExactList(t *testing.T) {
	path := cty.GetAttrPath("clusters").IndexInt(0)

	tests := []struct {
		name      string
		input     interface{}
		wantDiags int
	}{
		{name: "literal star emits warning", input: "*", wantDiags: 1},
		{name: "normal value emits nothing", input: "prod-cluster", wantDiags: 0},
		{name: "empty string emits nothing", input: "", wantDiags: 0},
		{name: "non-string input emits nothing", input: 42, wantDiags: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := warnLiteralStarInExactList(tc.input, path)
			assert.Len(t, diags, tc.wantDiags)
			if tc.wantDiags > 0 {
				assert.Equal(t, diag.Warning, diags[0].Severity, "must be a Warning, never an Error")
				assert.Equal(t, path, diags[0].AttributePath, "diagnostic must be anchored to the attribute path")
				assert.Contains(t, diags[0].Detail, `*_patterns`, "detail should suggest the *_patterns alternative")
			}
		})
	}
}

func TestValidateScopeDimension_MutualExclusion(t *testing.T) {
	itemsList := []interface{}{"foo", "bar"}
	patternsList := []interface{}{map[string]interface{}{"include": "foo-*"}}

	tests := []struct {
		name     string
		required bool
	}{
		{name: "required dimension rejects both", required: true},
		{name: "optional dimension rejects both", required: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope := map[string]interface{}{
				"clusters":          itemsList,
				"clusters_patterns": patternsList,
			}
			err := validateScopeDimension(0, scope, "clusters", "clusters_patterns", tc.required)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "mutually exclusive")
			}
		})
	}
}
