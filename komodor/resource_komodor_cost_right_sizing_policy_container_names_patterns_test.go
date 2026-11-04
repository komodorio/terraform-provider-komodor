package komodor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// fakePolicyAPI is an in-memory stand-in for the right-sizing policy endpoints. It stores the
// request body verbatim, like the real API: a PUT is a full replace, so any field the provider
// does not send is gone afterwards.
type fakePolicyAPI struct {
	mu     sync.Mutex
	policy map[string]interface{}
	puts   []map[string]interface{}
}

func (f *fakePolicyAPI) handler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	respond := func(status int) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "fake-policy-id", "policy": f.policy})
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == rsPoliciesPath, r.Method == http.MethodPut:
		raw, _ := io.ReadAll(r.Body)
		body := map[string]interface{}{}
		_ = json.Unmarshal(raw, &body)
		f.policy = body
		if r.Method == http.MethodPut {
			f.puts = append(f.puts, body)
			respond(http.StatusOK)
			return
		}
		respond(http.StatusCreated)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/fake-policy-id") && f.policy != nil:
		respond(http.StatusOK)
	case r.Method == http.MethodDelete:
		f.policy = nil
		w.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(w, r)
	}
}

func (f *fakePolicyAPI) containersPatterns() (interface{}, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.policy["containersPatterns"]
	return v, ok
}

func (f *fakePolicyAPI) setContainersPatterns(excludes ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	asAny := make([]interface{}, len(excludes))
	for i, e := range excludes {
		asAny[i] = e
	}
	f.policy["containersPatterns"] = map[string]interface{}{"excludes": asAny}
}

func testContainerNamesPatternsConfig(priority int, block string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = "tf-acc-container-names-patterns"
  priority            = %d
  optimization_preset = "sandbox"
  apply_protocol      = "onCreation"
  force_delete        = true

  scope {
    clusters       = ["tf-acc-cluster"]
    namespaces     = ["default"]
    workload_names = ["tf-acc-workload"]
  }
%s
}
`, priority, block)
}

func requireTerraformCLI(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform CLI not found in PATH; skipping plan-level test against the fake API")
	}
}

func runWithFakePolicyAPI(t *testing.T, steps func(api *fakePolicyAPI) []resource.TestStep) {
	t.Helper()
	requireTerraformCLI(t)
	api := &fakePolicyAPI{}
	srv := httptest.NewServer(http.HandlerFunc(api.handler))
	t.Cleanup(srv.Close)
	t.Setenv(KomodorAPIKeyEnvName, "00000000-0000-0000-0000-000000000000")
	t.Setenv(KomodorAPIURLEnvName, srv.URL)

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps:     steps(api),
	})
}

var testContainerNamesPatternsBlock = fmt.Sprintf(`
  %s {
    excludes = ["istio-proxy", "sidecar-*"]
  }
`, containerNamesPatternsAttr)

// Policy-level exclusions survive create, a no-op plan, an unrelated update and import, and
// removing the block clears them.
func TestFakeAPI_komodor_cost_right_sizing_policy_container_names_patterns_round_trip(t *testing.T) {
	const addr = "komodor_cost_right_sizing_policy.test"
	runWithFakePolicyAPI(t, func(api *fakePolicyAPI) []resource.TestStep {
		return []resource.TestStep{
			{
				Config: testContainerNamesPatternsConfig(99, testContainerNamesPatternsBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".#", "1"),
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".0.excludes.#", "2"),
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".0.excludes.0", "istio-proxy"),
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".0.excludes.1", "sidecar-*"),
					func(*terraform.State) error {
						got, ok := api.containersPatterns()
						if !ok {
							return fmt.Errorf("containersPatterns was not sent on create")
						}
						want := map[string]interface{}{"excludes": []interface{}{"istio-proxy", "sidecar-*"}}
						if fmt.Sprint(got) != fmt.Sprint(want) {
							return fmt.Errorf("containersPatterns = %v, want %v", got, want)
						}
						return nil
					},
				),
			},
			{
				// Read back from the API and re-applied unchanged: no diff.
				Config:   testContainerNamesPatternsConfig(99, testContainerNamesPatternsBlock),
				PlanOnly: true,
			},
			{
				// An update that does not touch the block must still send it (full-replace PUT).
				Config: testContainerNamesPatternsConfig(100, testContainerNamesPatternsBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "priority", "100"),
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".0.excludes.#", "2"),
					func(*terraform.State) error {
						if _, ok := api.containersPatterns(); !ok {
							return fmt.Errorf("containersPatterns was dropped by an unrelated update")
						}
						return nil
					},
				),
			},
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_delete"},
			},
			{
				Config: testContainerNamesPatternsConfig(100, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".#", "0"),
					func(*terraform.State) error {
						if v, ok := api.containersPatterns(); ok {
							return fmt.Errorf("containersPatterns = %v after removing the block, want the field absent", v)
						}
						return nil
					},
				),
			},
		}
	})
}

// A config without the block never sends the field and shows no diff against an API policy that
// has none.
func TestFakeAPI_komodor_cost_right_sizing_policy_container_names_patterns_omitted(t *testing.T) {
	const addr = "komodor_cost_right_sizing_policy.test"
	runWithFakePolicyAPI(t, func(api *fakePolicyAPI) []resource.TestStep {
		return []resource.TestStep{
			{
				Config: testContainerNamesPatternsConfig(99, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, containerNamesPatternsAttr+".#", "0"),
					func(*terraform.State) error {
						if v, ok := api.containersPatterns(); ok {
							return fmt.Errorf("containersPatterns = %v was sent although the block is not configured", v)
						}
						return nil
					},
				),
			},
			{Config: testContainerNamesPatternsConfig(99, ""), PlanOnly: true},
			{
				Config: testContainerNamesPatternsConfig(100, ""),
				Check: func(*terraform.State) error {
					if v, ok := api.containersPatterns(); ok {
						return fmt.Errorf("containersPatterns = %v was sent by an update without the block", v)
					}
					return nil
				},
			},
		}
	})
}

// Exclusions set outside Terraform (e.g. in the UI) show up as drift on the next plan, never
// silently: Terraform is the source of truth, so applying the unchanged config removes them.
func TestFakeAPI_komodor_cost_right_sizing_policy_container_names_patterns_drift_is_planned(t *testing.T) {
	runWithFakePolicyAPI(t, func(api *fakePolicyAPI) []resource.TestStep {
		return []resource.TestStep{
			{Config: testContainerNamesPatternsConfig(99, "")},
			{
				PreConfig:          func() { api.setContainersPatterns("added-in-ui") },
				Config:             testContainerNamesPatternsConfig(99, ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		}
	})
}
