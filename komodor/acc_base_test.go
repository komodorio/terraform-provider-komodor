package komodor

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

// accTestPrefix is prepended to all resource names created by acceptance tests,
// making it easy to identify and clean up leftover resources.
const accTestPrefix = "tf-acc-"

// TestMain wraps the test suite to perform pre-run cleanup of any resources
// left behind by a previously crashed acceptance test run.
// Cleanup only runs when TF_ACC=1 so that regular unit test runs are unaffected.
func TestMain(m *testing.M) {
	if os.Getenv("TF_ACC") == "1" {
		cleanupOrphanedAccResources()
	}
	os.Exit(m.Run())
}

// cleanupOrphanedAccResources deletes any test resources that start with
// accTestPrefix and were left behind by a previously killed test run.
// It only covers resource types that expose a list API — for others the
// terraform-plugin-sdk handles cleanup automatically during normal test runs.
func cleanupOrphanedAccResources() {
	apiKey := os.Getenv("KOMODOR_API_KEY")
	if apiKey == "" {
		return
	}
	apiURL := os.Getenv("KOMODOR_API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIBaseURL
	}
	client := NewClient(apiKey, apiURL)

	// Roles
	if roles, err := client.GetRoles(); err == nil {
		for _, role := range roles {
			if strings.HasPrefix(role.Name, accTestPrefix) {
				log.Printf("[CLEANUP] deleting orphaned role: %s (%s)", role.Name, role.Id)
				if err := client.DeleteRole(role.Id); err != nil {
					log.Printf("[CLEANUP] failed to delete role %s: %s", role.Id, err)
				}
			}
		}
	} else {
		log.Printf("[CLEANUP] could not list roles: %s", err)
	}

	// Monitors
	if monitors, err := client.GetMonitors(); err == nil {
		for _, monitor := range monitors {
			name := ""
			if monitor.Name != nil {
				name = *monitor.Name
			}
			if strings.HasPrefix(name, accTestPrefix) {
				log.Printf("[CLEANUP] deleting orphaned monitor: %s (%s)", name, monitor.Id)
				if err := client.DeleteMonitor(monitor.Id); err != nil {
					log.Printf("[CLEANUP] failed to delete monitor %s: %s", monitor.Id, err)
				}
			}
		}
	} else {
		log.Printf("[CLEANUP] could not list monitors: %s", err)
	}

	// Custom K8s Actions
	if actions, err := client.GetCustomK8sActions(); err == nil {
		for _, action := range actions {
			if strings.HasPrefix(action.Action, accTestPrefix) {
				log.Printf("[CLEANUP] deleting orphaned action: %s (%s)", action.Action, action.Id)
				if err := client.DeleteCustomK8sAction(action.Id); err != nil {
					log.Printf("[CLEANUP] failed to delete action %s: %s", action.Id, err)
				}
			}
		}
	} else {
		log.Printf("[CLEANUP] could not list custom k8s actions: %s", err)
	}

	// Knowledge Base Files
	for _, fileType := range []string{"knowledge-base", "blueprint"} {
		if files, _, err := client.ListKnowledgeBaseFiles(fileType); err == nil {
			var ids []string
			for _, f := range files {
				if strings.HasPrefix(f.Name, accTestPrefix) {
					log.Printf("[CLEANUP] deleting orphaned knowledge base file: %s (%s, type=%s)", f.Name, f.Id, fileType)
					ids = append(ids, f.Id)
				}
			}
			if len(ids) > 0 {
				if _, err := client.DeleteKnowledgeBaseFiles(ids, fileType); err != nil {
					log.Printf("[CLEANUP] failed to delete knowledge base files (type=%s): %s", fileType, err)
				}
			}
		} else {
			log.Printf("[CLEANUP] could not list knowledge base files (type=%s): %s", fileType, err)
		}
	}
}

// testAccPreCheck validates that the required environment variables are set
// before running any acceptance test.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("KOMODOR_API_KEY"); v == "" {
		t.Fatal("KOMODOR_API_KEY must be set for acceptance tests")
	}
}

// testResourceName returns a stable, prefixed name for a test resource of the
// given type. Using a fixed (non-random) name combined with concurrency: 1 in
// the Buildkite pipeline ensures that any resource left from a previous crashed
// run is overwritten or cleaned up predictably.
func testResourceName(resourceType string) string {
	return fmt.Sprintf("%s%s-test", accTestPrefix, resourceType)
}
