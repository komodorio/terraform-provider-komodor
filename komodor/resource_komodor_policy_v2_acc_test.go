package komodor

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_policy_v2")
}

func TestAcc_komodor_policy_v2_basic(t *testing.T) {
	name := testResourceName(t, "policy-v2-basic")
	updatedName := name + "-updated"
	resourceAddr := "komodor_policy_v2.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyV2Destroyed(name),
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccPolicyV2Config(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttr(resourceAddr, "statements.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "statements.0.actions.0", "view:all"),
				),
			},
			// Step 2: Import
			{
				ResourceName:      resourceAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update name and add a second statement
			{
				PreConfig: func() { time.Sleep(2 * time.Second) },
				Config:    testAccPolicyV2ConfigUpdated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", updatedName),
					resource.TestCheckResourceAttr(resourceAddr, "statements.#", "2"),
				),
			},
		},
	})
}

func testAccCheckPolicyV2Destroyed(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		policy, statusCode, err := client.GetPolicy(name)
		if err != nil && statusCode == 404 {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error checking policy destruction: %s", err)
		}
		if policy != nil {
			return fmt.Errorf("policy %q still exists after destroy", name)
		}
		return nil
	}
}

func testAccPolicyV2Config(name string) string {
	return fmt.Sprintf(`
resource "komodor_policy_v2" "test" {
  name = %q

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["tf-acc-cluster"]
      namespaces = ["default"]
    }
  }
}
`, name)
}

func testAccPolicyV2ConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "komodor_policy_v2" "test" {
  name = %q

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["tf-acc-cluster"]
      namespaces = ["default"]
    }
  }

  statements {
    actions = ["manage:users"]

    resources_scope {
      clusters_patterns {
        include = "*"
        exclude = ""
      }
      namespaces_patterns {
        include = "*"
        exclude = ""
      }
    }
  }
}
`, name)
}
