package komodor

import (
	"sort"
	"strings"
	"testing"
)

// registeredAccTests is populated by init() calls in each *_acc_test.go file.
// TestAccCoverage checks that every non-deprecated resource and data source in
// the provider has an entry here, ensuring CI fails automatically when a new
// resource or data source is added to provider.go without a corresponding
// acceptance test file.
var registeredAccTests = map[string]bool{}

// registerAccTest must be called from an init() function in each *_acc_test.go
// file that covers the given resource or data source type.
func registerAccTest(resourceName string) {
	registeredAccTests[resourceName] = true
}

// deprecatedResources lists resource names that are intentionally excluded from
// the coverage requirement.
var deprecatedResources = map[string]bool{}

// TestAccCoverage verifies that every active resource and data source in the
// provider has at least one acceptance test registered via registerAccTest.
//
// This test does NOT require TF_ACC or API credentials — it runs as part of
// Stage 4 (pre-acceptance gate) to catch missing tests before expensive E2E
// runs are attempted.
func TestAccCoverage(t *testing.T) {
	provider := Provider()

	var uncovered []string

	for resourceName := range provider.ResourcesMap {
		if deprecatedResources[resourceName] {
			continue
		}
		if !registeredAccTests[resourceName] {
			uncovered = append(uncovered, "resource: "+resourceName)
		}
	}

	for dsName := range provider.DataSourcesMap {
		if !registeredAccTests["datasource_"+dsName] {
			uncovered = append(uncovered, "data source: "+dsName)
		}
	}

	if len(uncovered) > 0 {
		sort.Strings(uncovered)
		t.Errorf(
			"the following resources/data sources have no acceptance tests — create a"+
				" *_acc_test.go file for each and call registerAccTest() in its init():\n  %s",
			strings.Join(uncovered, "\n  "),
		)
	}
}
