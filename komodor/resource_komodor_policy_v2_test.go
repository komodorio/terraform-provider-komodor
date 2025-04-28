package komodor

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// TestResourceKomodorPolicyV2 tests the schema expansion logic for different policy configurations.
// It verifies that the Terraform configuration is correctly converted to the internal policy structure
// for various policy types including basic policies, pattern-based policies, and selector-based policies.
func TestResourceKomodorPolicyV2(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected *NewPolicy
	}{
		{
			name: "basic policy",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default", "kube-system"},
							},
						},
					},
				},
			},
			expected: &NewPolicy{
				Name: "test-policy",
				Statements: []Statement{
					{
						Actions: []string{"view:all"},
						ResourcesScope: &ResourcesScope{
							Clusters:           []string{"prod-cluster"},
							Namespaces:         []string{"default", "kube-system"},
							ClustersPatterns:   []Pattern{},
							NamespacesPatterns: []Pattern{},
							Selectors:          []Selector{},
							SelectorsPatterns:  []SelectorPattern{},
						},
					},
				},
			},
		},
		{
			name: "policy with patterns",
			config: map[string]interface{}{
				"name": "pattern-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters_patterns": []interface{}{
									map[string]interface{}{
										"include": "prod-*",
										"exclude": "prod-legacy",
									},
								},
								"namespaces_patterns": []interface{}{
									map[string]interface{}{
										"include": "team-*",
										"exclude": "team-internal",
									},
								},
							},
						},
					},
				},
			},
			expected: &NewPolicy{
				Name: "pattern-policy",
				Statements: []Statement{
					{
						Actions: []string{"view:all"},
						ResourcesScope: &ResourcesScope{
							Clusters:           []string{},
							Namespaces:         []string{},
							ClustersPatterns:   []Pattern{{Include: "prod-*", Exclude: "prod-legacy"}},
							NamespacesPatterns: []Pattern{{Include: "team-*", Exclude: "team-internal"}},
							Selectors:          []Selector{},
							SelectorsPatterns:  []SelectorPattern{},
						},
					},
				},
			},
		},
		{
			name: "policy with selectors",
			config: map[string]interface{}{
				"name": "selector-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"get", "list", "watch"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
								"selectors": []interface{}{
									map[string]interface{}{
										"key":   "team",
										"type":  "annotation",
										"value": "platform",
									},
									map[string]interface{}{
										"key":   "env",
										"type":  "label",
										"value": "production",
									},
								},
							},
						},
					},
				},
			},
			expected: &NewPolicy{
				Name: "selector-policy",
				Statements: []Statement{
					{
						Actions: []string{"get", "list", "watch"},
						ResourcesScope: &ResourcesScope{
							Clusters:           []string{"prod-cluster"},
							Namespaces:         []string{"default"},
							ClustersPatterns:   []Pattern{},
							NamespacesPatterns: []Pattern{},
							Selectors: []Selector{
								{Key: "team", Type: "annotation", Value: "platform"},
								{Key: "env", Type: "label", Value: "production"},
							},
							SelectorsPatterns: []SelectorPattern{},
						},
					},
				},
			},
		},
		{
			name: "policy with selector patterns",
			config: map[string]interface{}{
				"name": "selector-pattern-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"get", "list"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
								"selectors_patterns": []interface{}{
									map[string]interface{}{
										"key":  "team",
										"type": "annotation",
										"value": []interface{}{
											map[string]interface{}{
												"include": "team-*",
												"exclude": "team-internal",
											},
										},
									},
									map[string]interface{}{
										"key":  "env",
										"type": "label",
										"value": []interface{}{
											map[string]interface{}{
												"include": "production",
												"exclude": "",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &NewPolicy{
				Name: "selector-pattern-policy",
				Statements: []Statement{
					{
						Actions: []string{"get", "list"},
						ResourcesScope: &ResourcesScope{
							Clusters:           []string{"prod-cluster"},
							Namespaces:         []string{"default"},
							ClustersPatterns:   []Pattern{},
							NamespacesPatterns: []Pattern{},
							Selectors:          []Selector{},
							SelectorsPatterns: []SelectorPattern{
								{
									Key:  "team",
									Type: "annotation",
									Value: Pattern{
										Include: "team-*",
										Exclude: "team-internal",
									},
								},
								{
									Key:  "env",
									Type: "label",
									Value: Pattern{
										Include: "production",
										Exclude: "",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := resourceKomodorPolicyV2()
			d := schema.TestResourceDataRaw(t, resource.Schema, tt.config)

			actual := expandPolicy(d)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestResourceKomodorPolicyV2Validation tests the schema validation rules for the policy resource.
// It verifies that the schema correctly enforces required fields and validates field types,
// ensuring that invalid configurations are rejected with appropriate errors.
func TestResourceKomodorPolicyV2Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "missing required name",
			config: map[string]interface{}{
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing required statements",
			config: map[string]interface{}{
				"name": "test-policy",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			config: map[string]interface{}{
				"name": "test-policy",
				"type": "invalid",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing required actions in statement",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing required resources_scope in statement",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid selector type",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
								"selectors": []interface{}{
									map[string]interface{}{
										"key":   "team",
										"type":  "invalid-type",
										"value": "platform",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid selector pattern type",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
								"selectors_patterns": []interface{}{
									map[string]interface{}{
										"key":  "team",
										"type": "invalid-type",
										"value": []interface{}{
											map[string]interface{}{
												"include": "team-*",
												"exclude": "team-internal",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing required include in pattern",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters_patterns": []interface{}{
									map[string]interface{}{
										"exclude": "prod-legacy",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing required exclude in pattern",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters_patterns": []interface{}{
									map[string]interface{}{
										"include": "prod-*",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid configuration",
			config: map[string]interface{}{
				"name": "test-policy",
				"statements": []interface{}{
					map[string]interface{}{
						"actions": []interface{}{"view:all"},
						"resources_scope": []interface{}{
							map[string]interface{}{
								"clusters":   []interface{}{"prod-cluster"},
								"namespaces": []interface{}{"default"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := resourceKomodorPolicyV2()
			rc := &terraform.ResourceConfig{
				Raw:    tt.config,
				Config: tt.config,
			}

			// Test schema validation
			errors := schema.InternalMap(resource.Schema).Validate(rc)
			if tt.wantErr {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}
