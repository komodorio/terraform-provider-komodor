package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_action")
}

func TestAcc_komodor_action_basic(t *testing.T) {
	actionName := testResourceName("action")
	resourceAddr := "komodor_action.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckActionDestroyed(actionName),
		Steps: []resource.TestStep{
			// Step 1: Create + no-drift
			{
				Config: testAccActionConfig(actionName, "View pods in default namespace"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "action", actionName),
					resource.TestCheckResourceAttr(resourceAddr, "description", "View pods in default namespace"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
			// Step 2: Update description
			{
				Config: testAccActionConfig(actionName, "View pods and deployments in default namespace"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "description", "View pods and deployments in default namespace"),
				),
			},
		},
	})
}

func testAccCheckActionDestroyed(actionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		actions, err := client.GetCustomK8sActions()
		if err != nil {
			return fmt.Errorf("error listing actions during destroy check: %s", err)
		}
		for _, a := range actions {
			if a.Action == actionName {
				return fmt.Errorf("action %q still exists after destroy", actionName)
			}
		}
		return nil
	}
}

func testAccActionConfig(actionName, description string) string {
	return fmt.Sprintf(`
resource "komodor_action" "test" {
  action      = %q
  description = %q

  ruleset = jsonencode([{
    apiGroups = [""]
    resources = ["pods"]
    verbs     = ["get", "list"]
  }])
}
`, actionName, description)
}
