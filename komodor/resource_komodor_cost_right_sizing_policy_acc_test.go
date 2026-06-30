package komodor

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_cost_right_sizing_policy")
}

func TestAcc_komodor_cost_right_sizing_policy_named_preset(t *testing.T) {
	name := testResourceName(t, "cost-rsp-named")
	resourceAddr := "komodor_cost_right_sizing_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRightSizingPolicyDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccCostRSPConfigNamedPreset(name, "initial description", 100, []string{"team:cost"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "optimization_preset", presetSandbox),
					resource.TestCheckResourceAttr(resourceAddr, "priority", "100"),
					resource.TestCheckResourceAttr(resourceAddr, "description", "initial description"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_at"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_by"),
					resource.TestCheckResourceAttrSet(resourceAddr, "updated_by"),
					resource.TestCheckResourceAttrSet(resourceAddr, "guardrails.0.percentile"),
					resource.TestCheckResourceAttrSet(resourceAddr, "guardrails.0.managed_resources.0.cpu_requests"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.0", "team:cost"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.1", managedByTag),
				),
			},
			{
				Config: testAccCostRSPConfigNamedPreset(name, "updated description", 200, []string{"team:cost", "owner:platform"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceAddr, "priority", "200"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.0", "team:cost"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.1", "owner:platform"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.2", managedByTag),
				),
			},
			{
				ResourceName:            resourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_delete"},
			},
		},
	})
}

func TestAcc_komodor_cost_right_sizing_policy_custom_preset(t *testing.T) {
	name := testResourceName(t, "cost-rsp-custom")
	resourceAddr := "komodor_cost_right_sizing_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRightSizingPolicyDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccCostRSPConfigCustomPreset(name, 95, true, []string{"team:cost"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "optimization_preset", presetCustom),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.percentile", "95"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.managed_resources.0.cpu_requests", "true"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.managed_resources.0.memory_requests", "true"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.0", "team:cost"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.1", managedByTag),
				),
			},
			{
				Config: testAccCostRSPConfigCustomPreset(name, 90, false, []string{managedByTag, "team:cost"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.percentile", "90"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.managed_resources.0.cpu_requests", "false"),
					resource.TestCheckResourceAttr(resourceAddr, "tags.#", "2"),
				),
			},
			{
				ResourceName:            resourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_delete"},
			},
			{
				Config: testAccCostRSPConfigCustomPreset(name, 90, false, []string{managedByTag, "team:cost"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRightSizingPolicyDisappears(resourceAddr),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc_komodor_cost_right_sizing_policy_multi_scope(t *testing.T) {
	name := testResourceName("cost-rsp-multi-scope")
	resourceAddr := "komodor_cost_right_sizing_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRightSizingPolicyDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccCostRSPConfigMultiScope(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.clusters.0", "cost-tests"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.namespaces.0", "noam"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.1.clusters.0", "cost-tests"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.1.namespaces.0", "default"),
				),
			},
			{
				ResourceName:            resourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_delete"},
			},
		},
	})
}

func testAccCheckRightSizingPolicyDestroyed(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := newRightSizingClientFromMeta(testAccProvider.Meta())
		_, status, err := client.GetByName(context.Background(), name)
		if status == http.StatusNotFound {
			return nil
		}
		if err != nil {
			return fmt.Errorf("checking right-sizing policy %q destroy: %s", name, err)
		}
		return fmt.Errorf("right-sizing policy %q still exists after destroy", name)
	}
}

func testAccCheckRightSizingPolicyDisappears(resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}
		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("resource %q has empty ID", resourceAddr)
		}
		client := newRightSizingClientFromMeta(testAccProvider.Meta())
		if err := client.Delete(context.Background(), id, true); err != nil {
			return fmt.Errorf("deleting right-sizing policy %q out-of-band: %s", id, err)
		}
		return nil
	}
}

const testAccCostRSPScope = `
  scope {
    clusters   = ["tf-acc-cluster"]
    namespaces = ["default"]
    workload_names_patterns {
      include = "tf-acc-*"
    }
  }
`

func testAccCostRSPConfigNamedPreset(name, description string, priority int, tags []string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  description         = %q
  priority            = %d
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true
  tags                = %s
%s
}
`, name, description, priority, presetSandbox, applyOnCreation, hclStringList(tags), testAccCostRSPScope)
}

func testAccCostRSPConfigCustomPreset(name string, percentile int, cpuRequestsEnabled bool, tags []string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 100
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true
  tags                = %s
%s
  guardrails {
    percentile          = %d
    allow_qos_upgrade   = false
    allow_qos_downgrade = false

    managed_resources {
      cpu_requests    = %t
      memory_requests = true
    }

    constraints {
      decrease_cpu_by {
        enabled = true
        value   = 25
      }
      decrease_memory_by {
        enabled = true
        value   = 25
      }
      increase_cpu_by {
        enabled = false
        value   = 0
      }
      increase_memory_by {
        enabled = false
        value   = 0
      }
    }

    buffer {
      cpu {
        enabled = true
        value   = 10
      }
      memory {
        enabled = true
        value   = 10
      }
    }
  }
}
`, name, presetCustom, applyOnCreation, hclStringList(tags), testAccCostRSPScope, percentile, cpuRequestsEnabled)
}

func testAccCostRSPConfigMultiScope(name string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 100
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true

  scope {
    clusters   = ["cost-tests"]
    namespaces = ["noam"]
    workload_names_patterns {
      include = "tf-acc-*"
    }
  }

  scope {
    clusters   = ["cost-tests"]
    namespaces = ["default"]
    workload_names_patterns {
      include = "tf-acc-*"
    }
  }
}
`, name, presetSandbox, applyOnCreation)
}

func hclStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	out := "["
	for i, s := range items {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%q", s)
	}
	out += "]"
	return out
}
