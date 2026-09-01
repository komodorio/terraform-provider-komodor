package komodor

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
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
				Config: testAccCostRSPConfigNamedPreset(name, "initial description", 99, []string{"team:cost"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "optimization_preset", presetSandbox),
					resource.TestCheckResourceAttr(resourceAddr, "priority", "99"),
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
				Config: testAccCostRSPConfigCustomPreset(name, "percentile = 95", true, []string{"team:cost"}),
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
				Config: testAccCostRSPConfigCustomPreset(name, "percentile = 90", false, []string{managedByTag, "team:cost"}),
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
				Config: testAccCostRSPConfigCustomPreset(name, "percentile = 90", false, []string{managedByTag, "team:cost"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRightSizingPolicyDisappears(resourceAddr),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc_komodor_cost_right_sizing_policy_split_percentiles(t *testing.T) {
	name := testResourceName("cost-rsp-split-pct")
	resourceAddr := "komodor_cost_right_sizing_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRightSizingPolicyDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccCostRSPConfigSplitPercentiles(name, 90, 99),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.cpu_percentile", "90"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.memory_percentile", "99"),
				),
			},
			{
				Config: testAccCostRSPConfigSplitPercentiles(name, 95, 80),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.cpu_percentile", "95"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.memory_percentile", "80"),
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

func TestAcc_komodor_cost_right_sizing_policy_percentile_cycle(t *testing.T) {
	name := testResourceName("cost-rsp-cycle")
	resourceAddr := "komodor_cost_right_sizing_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRightSizingPolicyDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccCostRSPConfigCustomPreset(name, "cpu_percentile    = 70\n    memory_percentile = 80", true, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.cpu_percentile", "70"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.memory_percentile", "80"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.percentile", "0"),
				),
			},
			{
				Config: testAccCostRSPConfigCustomPreset(name, "percentile = 70", true, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.percentile", "70"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.cpu_percentile", "0"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.memory_percentile", "0"),
				),
			},
			{
				Config: testAccCostRSPConfigCustomPreset(name, "cpu_percentile    = 90\n    memory_percentile = 95", true, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.cpu_percentile", "90"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.memory_percentile", "95"),
					resource.TestCheckResourceAttr(resourceAddr, "guardrails.0.percentile", "0"),
				),
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
				// Legacy singular include must still plan clean after this change.
				Config:   testAccCostRSPConfigMultiScope(name),
				PlanOnly: true,
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

func TestAcc_komodor_cost_right_sizing_policy_pattern_includes_excludes(t *testing.T) {
	name := testResourceName("cost-rsp-includes")
	resourceAddr := "komodor_cost_right_sizing_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRightSizingPolicyDestroyed(name),
		Steps: []resource.TestStep{
			{
				// Order is deliberately non-alphabetical (web before api) — don't "tidy" this;
				// see the comment on tfToAPIPattern.
				Config: testAccCostRSPConfigPatternIncludesExcludes(name, []string{"tf-acc-web-*", "tf-acc-api-*"}, []string{"tf-acc-web-canary", "tf-acc-api-canary"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.includes.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.includes.0", "tf-acc-web-*"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.includes.1", "tf-acc-api-*"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.excludes.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.excludes.0", "tf-acc-web-canary"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.excludes.1", "tf-acc-api-canary"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.include", ""),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.exclude", ""),
				),
			},
			{
				Config: testAccCostRSPConfigPatternIncludesExcludes(name, []string{"tf-acc-api-*"}, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.includes.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.includes.0", "tf-acc-api-*"),
					resource.TestCheckResourceAttr(resourceAddr, "scope.0.workload_names_patterns.0.excludes.#", "0"),
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

func TestAcc_komodor_cost_right_sizing_policy_pattern_conflicting_fields(t *testing.T) {
	name := testResourceName("cost-rsp-conflict")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCostRSPConfigConflictingIncludeAndIncludes(name),
				ExpectError: regexp.MustCompile(`"include" and "includes" are mutually exclusive`),
			},
			{
				Config:      testAccCostRSPConfigEmptyIncludesList(name),
				ExpectError: regexp.MustCompile(`one of "include" or "includes" is required`),
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

func testAccCostRSPConfigCustomPreset(name, percentileLines string, cpuRequestsEnabled bool, tags []string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 99
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true
  tags                = %s
%s
  guardrails {
    %s
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
`, name, presetCustom, applyOnCreation, hclStringList(tags), testAccCostRSPScope, percentileLines, cpuRequestsEnabled)
}

func testAccCostRSPConfigSplitPercentiles(name string, cpuPercentile, memoryPercentile int) string {
	return testAccCostRSPConfigCustomPreset(name, fmt.Sprintf("cpu_percentile    = %d\n    memory_percentile = %d", cpuPercentile, memoryPercentile), true, nil)
}

func testAccCostRSPConfigMultiScope(name string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 99
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

func testAccCostRSPConfigPatternIncludesExcludes(name string, includes, excludes []string) string {
	excludesBlock := ""
	if len(excludes) > 0 {
		excludesBlock = fmt.Sprintf("\n      excludes = %s", hclStringList(excludes))
	}
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 99
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true

  scope {
    clusters   = ["tf-acc-cluster"]
    namespaces = ["default"]
    workload_names_patterns {
      includes = %s%s
    }
  }
}
`, name, presetSandbox, applyOnCreation, hclStringList(includes), excludesBlock)
}

func testAccCostRSPConfigConflictingIncludeAndIncludes(name string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 99
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true

  scope {
    clusters   = ["tf-acc-cluster"]
    namespaces = ["default"]
    workload_names_patterns {
      include  = "tf-acc-*"
      includes = ["tf-acc-api-*"]
    }
  }
}
`, name, presetSandbox, applyOnCreation)
}

func testAccCostRSPConfigEmptyIncludesList(name string) string {
	return fmt.Sprintf(`
resource "komodor_cost_right_sizing_policy" "test" {
  name                = %q
  priority            = 99
  optimization_preset = %q
  apply_protocol      = %q
  force_delete        = true

  scope {
    clusters   = ["tf-acc-cluster"]
    namespaces = ["default"]
    workload_names_patterns {
      includes = []
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
