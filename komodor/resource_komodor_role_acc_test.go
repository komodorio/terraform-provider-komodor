package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_role")
}

func TestAcc_komodor_role_basic(t *testing.T) {
	name := testResourceName("role")
	updatedName := name + "-updated"
	resourceAddr := "komodor_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroyed(name),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_at"),
				),
			},
			// Update: role update is implemented as delete+recreate; the plan
			// after apply must still show no pending changes.
			{
				Config: testAccRoleConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", updatedName),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func testAccCheckRoleDestroyed(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		role, err := client.GetRoleByName(name)
		if err != nil {
			return fmt.Errorf("error checking role destruction: %s", err)
		}
		if role != nil {
			return fmt.Errorf("role %q still exists after destroy", name)
		}
		return nil
	}
}

func testAccRoleConfig(name string) string {
	return fmt.Sprintf(`
resource "komodor_role" "test" {
  name = %q
}
`, name)
}
