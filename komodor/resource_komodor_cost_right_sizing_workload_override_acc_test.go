package komodor

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_cost_right_sizing_workload_override")
}

func TestAcc_workload_override_include(t *testing.T) {
	t.Run("creates an include override that pins a workload to a policy", func(t *testing.T) {
		workload := testResourceName("workload-override-include")
		resourceAddr := "komodor_cost_right_sizing_workload_override.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckWorkloadOverrideDestroyed(workload),
			Steps: []resource.TestStep{
				{
					Config: testAccWorkloadOverrideConfigInclude(workload),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionInclude),
						resource.TestCheckResourceAttr(resourceAddr, "name", workload),
						resource.TestCheckResourceAttr(resourceAddr, "kind", "Deployment"),
						resource.TestCheckResourceAttrSet(resourceAddr, "id"),
						resource.TestCheckResourceAttrPair(resourceAddr, "policy_id", "komodor_cost_right_sizing_policy.test", "id"),
					),
				},
				{
					ResourceName:      resourceAddr,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAcc_workload_override_exclude(t *testing.T) {
	t.Run("creates an exclude override that removes a workload from right-sizing", func(t *testing.T) {
		workload := testResourceName("workload-override-exclude")
		resourceAddr := "komodor_cost_right_sizing_workload_override.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckWorkloadOverrideDestroyed(workload),
			Steps: []resource.TestStep{
				{
					Config: testAccWorkloadOverrideConfigExclude(workload),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionExclude),
						resource.TestCheckResourceAttr(resourceAddr, "name", workload),
						resource.TestCheckResourceAttr(resourceAddr, "policy_id", ""),
						resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					),
				},
				{
					ResourceName:      resourceAddr,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAcc_workload_override_exclude_with_policy(t *testing.T) {
	t.Run("creates an exclude override that keeps an explicitly assigned policy", func(t *testing.T) {
		workload := testResourceName("workload-override-exclude-policy")
		resourceAddr := "komodor_cost_right_sizing_workload_override.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckWorkloadOverrideDestroyed(workload),
			Steps: []resource.TestStep{
				{
					Config: testAccWorkloadOverrideConfigExcludeWithPolicy(workload),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionExclude),
						resource.TestCheckResourceAttrPair(resourceAddr, "policy_id", "komodor_cost_right_sizing_policy.test", "id"),
						resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					),
				},
				{
					ResourceName:      resourceAddr,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAcc_workload_override_update(t *testing.T) {
	t.Run("updates the policy and action while keeping the same id", func(t *testing.T) {
		workload := testResourceName("workload-override-update")
		resourceAddr := "komodor_cost_right_sizing_workload_override.test"
		var overrideID string

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckWorkloadOverrideDestroyed(workload),
			Steps: []resource.TestStep{
				{
					Config: testAccWorkloadOverrideUpdateInclude(workload, "komodor_cost_right_sizing_policy.test"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideIDStable(resourceAddr, &overrideID),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionInclude),
						resource.TestCheckResourceAttrPair(resourceAddr, "policy_id", "komodor_cost_right_sizing_policy.test", "id"),
					),
				},
				{
					Config: testAccWorkloadOverrideUpdateInclude(workload, "komodor_cost_right_sizing_policy.test2"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideIDStable(resourceAddr, &overrideID),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionInclude),
						resource.TestCheckResourceAttrPair(resourceAddr, "policy_id", "komodor_cost_right_sizing_policy.test2", "id"),
					),
				},
				{
					Config: testAccWorkloadOverrideUpdateExclude(workload),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideIDStable(resourceAddr, &overrideID),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionExclude),
						resource.TestCheckResourceAttrPair(resourceAddr, "policy_id", "komodor_cost_right_sizing_policy.test2", "id"),
					),
				},
				{
					Config: testAccWorkloadOverrideUpdateInclude(workload, "komodor_cost_right_sizing_policy.test"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideIDStable(resourceAddr, &overrideID),
						resource.TestCheckResourceAttr(resourceAddr, "action", workloadOverrideActionInclude),
						resource.TestCheckResourceAttrPair(resourceAddr, "policy_id", "komodor_cost_right_sizing_policy.test", "id"),
					),
				},
			},
		})
	})
}

func TestAcc_workload_override_recreate(t *testing.T) {
	t.Run("recreates the override after it is deleted outside terraform", func(t *testing.T) {
		workload := testResourceName("workload-override-recreate")
		resourceAddr := "komodor_cost_right_sizing_workload_override.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckWorkloadOverrideDestroyed(workload),
			Steps: []resource.TestStep{
				{
					Config: testAccWorkloadOverrideConfigInclude(workload),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideManuallyDeleted(resourceAddr),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		})
	})
}

func TestAcc_workload_override_update_identity(t *testing.T) {
	t.Run("changes a workload identity field without recreating the override", func(t *testing.T) {
		workload := testResourceName("workload-override-identity")
		resourceAddr := "komodor_cost_right_sizing_workload_override.test"
		var overrideID string

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckWorkloadOverrideDestroyed(workload),
			Steps: []resource.TestStep{
				{
					Config: testAccWorkloadOverrideConfigIncludeKind(workload, "Deployment"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideIDStable(resourceAddr, &overrideID),
						resource.TestCheckResourceAttr(resourceAddr, "kind", "Deployment"),
					),
				},
				{
					Config: testAccWorkloadOverrideConfigIncludeKind(workload, "StatefulSet"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckWorkloadOverrideExists(resourceAddr),
						testAccCheckWorkloadOverrideIDStable(resourceAddr, &overrideID),
						resource.TestCheckResourceAttr(resourceAddr, "kind", "StatefulSet"),
					),
				},
			},
		})
	})
}

func TestAcc_workload_override_rejects_invalid_config(t *testing.T) {
	cases := []struct {
		name   string
		config string
		errRe  string
	}{
		{
			name: "rejects an include override without a policy_id",
			config: `
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = "include"
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = "tf-acc-validation"
}
`,
			errRe: `policy_id is required`,
		},
		{
			name: "rejects a policy_id that is not a valid uuid",
			config: `
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = "include"
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = "tf-acc-validation"
  policy_id    = "not-a-uuid"
}
`,
			errRe: `to be a valid UUID`,
		},
		{
			name: "rejects an unsupported action value",
			config: `
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = "destroy"
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = "tf-acc-validation"
}
`,
			errRe: `unsupported action`,
		},
		{
			name: "rejects an empty required field",
			config: `
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = "exclude"
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = ""
}
`,
			errRe: `must not be empty`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(tc.errRe),
					},
				},
			})
		})
	}
}

func testAccCheckWorkloadOverrideExists(resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}
		client := testAccProvider.Meta().(*Client)
		_, found, err := client.GetOverrideByID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("checking workload override %q exists: %s", rs.Primary.ID, err)
		}
		if !found {
			return fmt.Errorf("workload override %q not found via API", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckWorkloadOverrideManuallyDeleted(resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}
		client := testAccProvider.Meta().(*Client)
		if err := client.DeleteOverride(rs.Primary.ID); err != nil {
			return fmt.Errorf("deleting workload override %q out-of-band: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

// testAccCheckWorkloadOverrideIDStable captures the resource ID on first call
// and asserts it stays the same on subsequent calls, proving the override keeps
// its ID rather than being destroyed and recreated.
func testAccCheckWorkloadOverrideIDStable(resourceAddr string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}
		if *id == "" {
			*id = rs.Primary.ID
			return nil
		}
		if rs.Primary.ID != *id {
			return fmt.Errorf("expected the override to keep its ID but it changed: %s -> %s", *id, rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckWorkloadOverrideDestroyed(workloadName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		overrides, err := client.GetOverrides()
		if err != nil {
			return fmt.Errorf("checking workload override %q destroy: %s", workloadName, err)
		}
		for _, o := range overrides {
			if o.Name == workloadName {
				return fmt.Errorf("workload override for %q still exists after destroy", workloadName)
			}
		}
		return nil
	}
}

func testAccWorkloadOverridePolicy(workloadName string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = "%s-policy"
  priority            = 99
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true

  scope {
    clusters   = ["cost-tests"]
    namespaces = ["default"]
    workload_names_patterns {
      include = "tf-acc-*"
    }
  }
}
`, workloadName, presetSandbox, applyOnCreation)
}

func testAccWorkloadOverrideConfigInclude(workloadName string) string {
	return testAccWorkloadOverrideConfigIncludeKind(workloadName, "Deployment")
}

func testAccWorkloadOverrideConfigIncludeKind(workloadName, kind string) string {
	return testAccWorkloadOverridePolicy(workloadName) + fmt.Sprintf(`
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = %q
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = %q
  name         = %q
  policy_id    = komodor_cost_right_sizing_policy.test.id
}
`, workloadOverrideActionInclude, kind, workloadName)
}

func testAccWorkloadOverrideConfigExclude(workloadName string) string {
	return testAccWorkloadOverridePolicy(workloadName) + fmt.Sprintf(`
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = %q
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = %q
}
`, workloadOverrideActionExclude, workloadName)
}

func testAccWorkloadOverrideConfigExcludeWithPolicy(workloadName string) string {
	return testAccWorkloadOverridePolicy(workloadName) + fmt.Sprintf(`
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = %q
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = %q
  policy_id    = komodor_cost_right_sizing_policy.test.id
}
`, workloadOverrideActionExclude, workloadName)
}

func testAccWorkloadOverrideUpdatePolicies(workloadName string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = "%[1]s-policy-a"
  priority            = 99
  optimization_preset = %[2]q
  apply_protocol      = %[3]q
  force_delete        = true

  scope {
    clusters   = ["cost-tests"]
    namespaces = ["default"]
    workload_names_patterns {
      include = "tf-acc-*"
    }
  }
}

resource "komodor_cost_right_sizing_policy" "test2" {
  name                = "%[1]s-policy-b"
  priority            = 200
  optimization_preset = %[2]q
  apply_protocol      = %[3]q
  force_delete        = true

  scope {
    clusters   = ["cost-tests"]
    namespaces = ["default"]
    workload_names_patterns {
      include = "tf-acc-*"
    }
  }
}
`, workloadName, presetSandbox, applyOnCreation)
}

func testAccWorkloadOverrideUpdateInclude(workloadName, policyRef string) string {
	return testAccWorkloadOverrideUpdatePolicies(workloadName) + fmt.Sprintf(`
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = %q
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = %q
  policy_id    = %s.id
}
`, workloadOverrideActionInclude, workloadName, policyRef)
}

func testAccWorkloadOverrideUpdateExclude(workloadName string) string {
	return testAccWorkloadOverrideUpdatePolicies(workloadName) + fmt.Sprintf(`
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = %q
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = %q
}
`, workloadOverrideActionExclude, workloadName)
}
