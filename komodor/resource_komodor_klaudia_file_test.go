package komodor

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/require"
)

func TestExtractKlaudiaFileConfigString(t *testing.T) {
	raw := cty.ObjectVal(map[string]cty.Value{
		"type":     cty.StringVal("knowledge-base"),
		"filename": cty.StringVal("runbook.md"),
		"content":  cty.StringVal("hello world"),
	})

	value, ok := extractKlaudiaFileConfigString(raw, "content")
	require.True(t, ok)
	require.Equal(t, "hello world", value)
}

func TestParseKlaudiaFileImportID(t *testing.T) {
	fileType, fileID, err := parseKlaudiaFileImportID("knowledge-base:abc-123")
	require.NoError(t, err)
	require.Equal(t, "knowledge-base", fileType)
	require.Equal(t, "abc-123", fileID)
}

func TestParseKlaudiaFileImportIDRejectsMalformedValue(t *testing.T) {
	_, _, err := parseKlaudiaFileImportID("knowledge-base")
	require.Error(t, err)
}
