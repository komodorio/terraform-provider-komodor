package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_cost_right_sizing_workload_override")
}

func TestAcc_komodor_cost_right_sizing_workload_override_include(t *testing.T) {
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
}

func TestAcc_komodor_cost_right_sizing_workload_override_exclude(t *testing.T) {
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
}

func TestAcc_komodor_cost_right_sizing_workload_override_update(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceAddr, "policy_id", ""),
				),
			},
		},
	})
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

// testAccCheckWorkloadOverrideIDStable captures the resource ID on first call
// and asserts it stays the same on subsequent calls, proving updates happen
// in-place rather than as a destroy+recreate.
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
			return fmt.Errorf("expected in-place update but ID changed: %s -> %s", *id, rs.Primary.ID)
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
  priority            = 100
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
	return testAccWorkloadOverridePolicy(workloadName) + fmt.Sprintf(`
resource "komodor_cost_right_sizing_workload_override" "test" {
  action       = %q
  cluster_name = "cost-tests"
  namespace    = "default"
  kind         = "Deployment"
  name         = %q
  policy_id    = komodor_cost_right_sizing_policy.test.id
}
`, workloadOverrideActionInclude, workloadName)
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

func testAccWorkloadOverrideUpdatePolicies(workloadName string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = "%[1]s-policy-a"
  priority            = 100
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
