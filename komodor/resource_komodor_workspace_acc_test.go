package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_workspace")
}

func TestAcc_komodor_workspace_basic(t *testing.T) {
	name := testResourceName(t, "workspace")
	updatedName := name + "-updated"
	resourceAddr := "komodor_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceDestroyed(name),
		Steps: []resource.TestStep{
			// Step 1: Create with specific clusters/namespaces
			{
				Config: testAccWorkspaceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "description", "acceptance test workspace"),
					resource.TestCheckResourceAttr(resourceAddr, "scopes.#", "1"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
			// Step 2: Import
			{
				ResourceName:      resourceAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update — change description and scope
			{
				Config: testAccWorkspaceConfigUpdated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", updatedName),
					resource.TestCheckResourceAttr(resourceAddr, "description", "updated acceptance test workspace"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceDestroyed(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "komodor_workspace" {
				continue
			}
			ws, statusCode, err := client.GetWorkspace(rs.Primary.ID)
			if statusCode == 404 || ws == nil {
				continue
			}
			if err != nil {
				return fmt.Errorf("error checking workspace destruction: %s", err)
			}
			return fmt.Errorf("workspace %q still exists after destroy", name)
		}
		return nil
	}
}

func testAccWorkspaceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodor_workspace" "test" {
  name        = %q
  description = "acceptance test workspace"

  scopes {
    clusters   = ["tf-acc-cluster-1", "tf-acc-cluster-2"]
    namespaces = ["default"]
  }
}
`, name)
}

func testAccWorkspaceConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "komodor_workspace" "test" {
  name        = %q
  description = "updated acceptance test workspace"

  scopes {
    clusters_patterns {
      include = "tf-acc-*"
      exclude = ""
    }
    namespaces_patterns {
      include = "*"
      exclude = ""
    }
  }
}
`, name)
}
