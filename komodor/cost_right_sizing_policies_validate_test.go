package komodor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
