package komodor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandKnowledgeBaseClusters verifies that expandKnowledgeBaseClusters correctly
// converts the Terraform schema clusters block into a KnowledgeBaseScopedClusters struct
// for various input configurations.
func TestExpandKnowledgeBaseClusters(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected *KnowledgeBaseScopedClusters
	}{
		{
			name: "no clusters block returns nil",
			config: map[string]interface{}{
				"filename": "test.md",
				"content":  "content",
			},
			expected: nil,
		},
		{
			name: "empty clusters list returns nil",
			config: map[string]interface{}{
				"filename": "test.md",
				"content":  "content",
				"clusters": []interface{}{},
			},
			expected: nil,
		},
		{
			name: "clusters with include only",
			config: map[string]interface{}{
				"filename": "test.md",
				"content":  "content",
				"clusters": []interface{}{
					map[string]interface{}{
						"include": []interface{}{"prod-us-east-1", "prod-eu-west-1"},
						"exclude": []interface{}{},
					},
				},
			},
			expected: &KnowledgeBaseScopedClusters{
				Include: []string{"prod-us-east-1", "prod-eu-west-1"},
			},
		},
		{
			name: "clusters with exclude only",
			config: map[string]interface{}{
				"filename": "test.md",
				"content":  "content",
				"clusters": []interface{}{
					map[string]interface{}{
						"include": []interface{}{},
						"exclude": []interface{}{"staging"},
					},
				},
			},
			expected: &KnowledgeBaseScopedClusters{
				Exclude: []string{"staging"},
			},
		},
		{
			name: "clusters with both include and exclude",
			config: map[string]interface{}{
				"filename": "test.md",
				"content":  "content",
				"clusters": []interface{}{
					map[string]interface{}{
						"include": []interface{}{"prod-us-east-1"},
						"exclude": []interface{}{"staging"},
					},
				},
			},
			expected: &KnowledgeBaseScopedClusters{
				Include: []string{"prod-us-east-1"},
				Exclude: []string{"staging"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := resourceKomodorKnowledgeBaseFile()
			d := schema.TestResourceDataRaw(t, resource.Schema, tt.config)

			actual := expandKnowledgeBaseClusters(d)

			if tt.expected == nil {
				assert.Nil(t, actual)
				return
			}

			assert.NotNil(t, actual)
			assert.ElementsMatch(t, tt.expected.Include, actual.Include)
			assert.ElementsMatch(t, tt.expected.Exclude, actual.Exclude)
		})
	}
}

// TestDeleteKnowledgeBaseFilesEmptyResponse verifies that DeleteKnowledgeBaseFiles
// handles an empty (204 No Content) response body without error.
func TestDeleteKnowledgeBaseFilesEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL)
	client.HttpClient = server.Client()

	resp, _, err := client.DeleteKnowledgeBaseFiles([]string{"file-id-1"}, "knowledge-base")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.DeletedFiles)
	assert.Empty(t, resp.FailedFiles)
}
