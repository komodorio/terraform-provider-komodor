package komodor

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestScopeStringFields_WireLiteralStarWarning(t *testing.T) {
	scopeRes := costRSPScopeResource()
	fields := []string{"clusters", "namespaces", "resource_types", "workload_names"}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			fieldSchema, ok := scopeRes.Schema[field]
			if !assert.True(t, ok, "scope schema must define %q", field) {
				return
			}
			elem, ok := fieldSchema.Elem.(*schema.Schema)
			if !assert.True(t, ok, "field %q must use *schema.Schema as Elem", field) {
				return
			}
			if !assert.NotNil(t, elem.ValidateDiagFunc, "field %q must wire ValidateDiagFunc on its element", field) {
				return
			}
			path := cty.GetAttrPath(field).IndexInt(0)

			warn := elem.ValidateDiagFunc("*", path)
			if assert.Len(t, warn, 1, "field %q must warn on literal *", field) {
				assert.Equal(t, diag.Warning, warn[0].Severity)
			}

			none := elem.ValidateDiagFunc("real-name", path)
			assert.Empty(t, none, "field %q must not warn on a normal value", field)
		})
	}
}

func TestCheckPercentileConfig(t *testing.T) {
	tests := []struct {
		name                string
		shared, cpu, memory int
		wantErr             string
	}{
		{name: "shared only", shared: 95},
		{name: "split only", cpu: 95, memory: 99},
		{name: "split equal values", cpu: 90, memory: 90},
		{name: "shared and cpu", shared: 95, cpu: 90, wantErr: "mutually exclusive"},
		{name: "shared and memory", shared: 95, memory: 90, wantErr: "mutually exclusive"},
		{name: "shared and both split", shared: 95, cpu: 90, memory: 99, wantErr: "mutually exclusive"},
		{name: "nothing set", wantErr: "requires a percentile for both"},
		{name: "cpu only", cpu: 95, wantErr: "requires a percentile for both"},
		{name: "memory only", memory: 95, wantErr: "requires a percentile for both"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPercentileConfig(tc.shared, tc.cpu, tc.memory)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestValidatePatternValue(t *testing.T) {
	pattern := func(m map[string]interface{}) interface{} {
		base := map[string]interface{}{
			"include":  "",
			"includes": []interface{}{},
			"exclude":  "",
			"excludes": []interface{}{},
		}
		for k, v := range m {
			base[k] = v
		}
		return []interface{}{base}
	}

	tests := []struct {
		name    string
		pattern interface{}
		wantErr string
	}{
		{
			name:    "no block set is fine (dimension-level check handles required-ness)",
			pattern: []interface{}{},
		},
		{
			name:    "include only",
			pattern: pattern(map[string]interface{}{"include": "prod-*"}),
		},
		{
			name:    "includes only",
			pattern: pattern(map[string]interface{}{"includes": []interface{}{"prod-*", "staging-*"}}),
		},
		{
			name:    "include and includes both set",
			pattern: pattern(map[string]interface{}{"include": "prod-*", "includes": []interface{}{"staging-*"}}),
			wantErr: `"include" and "includes" are mutually exclusive`,
		},
		{
			name:    "exclude and excludes both set",
			pattern: pattern(map[string]interface{}{"include": "prod-*", "exclude": "prod-canary", "excludes": []interface{}{"prod-canary"}}),
			wantErr: `"exclude" and "excludes" are mutually exclusive`,
		},
		{
			name:    "neither include nor includes set",
			pattern: pattern(map[string]interface{}{"exclude": "prod-canary"}),
			wantErr: `one of "include" or "includes" is required`,
		},
		{
			name:    "empty includes list is equivalent to no positive selector",
			pattern: pattern(map[string]interface{}{"includes": []interface{}{}}),
			wantErr: `one of "include" or "includes" is required`,
		},
		{
			name:    "empty excludes list is valid",
			pattern: pattern(map[string]interface{}{"include": "prod-*", "excludes": []interface{}{}}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePatternValue(0, "workload_names_patterns", tc.pattern)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.wantErr)
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

func TestValidateContainerNamePattern(t *testing.T) {
	tests := []struct {
		name    string
		in      interface{}
		wantErr string
	}{
		{name: "exact name", in: "istio-proxy"},
		{name: "wildcard suffix", in: "sidecar-*"},
		{name: "wildcard prefix", in: "*-logger"},
		{name: "wildcard in the middle", in: "a*b"},
		{name: "star mixed with text is fine", in: "*a*"},
		{name: "empty entry", in: "", wantErr: "must not be empty"},
		{name: "single star matches every container", in: "*", wantErr: "matches every container"},
		{name: "repeated stars match every container", in: "***", wantErr: "matches every container"},
		{name: "not a string", in: 1, wantErr: "expected a string"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := validateContainerNamePattern(tc.in, containerNamesPatternsAttr+".0.excludes.0")
			if tc.wantErr == "" {
				assert.Empty(t, errs)
				return
			}
			if assert.Len(t, errs, 1) {
				assert.Contains(t, errs[0].Error(), tc.wantErr)
			}
		})
	}
}

// Exercises the real resource schema (MinItems/MaxItems/Required plus the element validator)
// without any API access.
func TestContainerNamesPatternsBlock_SchemaValidation(t *testing.T) {
	cfg := func(block interface{}) map[string]interface{} {
		c := map[string]interface{}{
			"name":                "schema-test",
			"priority":            1,
			"optimization_preset": presetSandbox,
			"apply_protocol":      applyOnCreation,
			"scope": []interface{}{map[string]interface{}{
				"clusters":       []interface{}{"c"},
				"namespaces":     []interface{}{"n"},
				"workload_names": []interface{}{"w"},
			}},
		}
		if block != nil {
			c[containerNamesPatternsAttr] = block
		}
		return c
	}
	excludes := func(items ...interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"excludes": items}}
	}

	tests := []struct {
		name    string
		block   interface{}
		wantErr string
	}{
		{name: "block omitted", block: nil},
		{name: "valid", block: excludes("istio-proxy", "sidecar-*")},
		{name: "empty excludes list", block: excludes(), wantErr: "excludes"},
		{name: "empty entry", block: excludes("istio-proxy", ""), wantErr: "must not be empty"},
		{name: "match-all pattern", block: excludes("*"), wantErr: "matches every container"},
		{name: "block without excludes", block: []interface{}{map[string]interface{}{}}, wantErr: `"container_names_patterns.0.excludes" is required`},
		{
			name:    "two blocks",
			block:   append(excludes("a"), excludes("b")...),
			wantErr: "supports 1 item maximum",
		},
	}
	res := resourceKomodorCostRightSizingPolicy()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := res.Validate(terraform.NewResourceConfigRaw(cfg(tc.block)))
			if tc.wantErr == "" {
				assert.False(t, diags.HasError(), "%v", diags)
				return
			}
			require.True(t, diags.HasError(), "expected an error containing %q", tc.wantErr)
			var all string
			for _, d := range diags {
				all += d.Summary + " " + d.Detail + "\n"
			}
			assert.Contains(t, all, tc.wantErr)
		})
	}
}
