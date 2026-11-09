package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_user_role_binding")
}

func TestAcc_komodor_user_role_binding_basic(t *testing.T) {
	userEmail := accTestPrefix + "binding-user@komodor-test.com"
	roleName := testResourceName(t, "binding-role")
	role2Name := testResourceName(t, "binding-role2")
	bindingName := testResourceName(t, "user-role-binding")
	resourceAddr := "komodor_user_role_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserRoleBindingDestroyed(userEmail),
		Steps: []resource.TestStep{
			// Step 1: Create user, role, and binding with 1 role
			{
				Config: testAccUserRoleBindingConfig(userEmail, roleName, bindingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", bindingName),
					resource.TestCheckResourceAttr(resourceAddr, "roles.#", "1"),
				),
			},
			// Step 2: Update binding to include a second role
			{
				Config: testAccUserRoleBindingConfigTwoRoles(userEmail, roleName, role2Name, bindingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", bindingName),
					resource.TestCheckResourceAttr(resourceAddr, "roles.#", "2"),
				),
			},
		},
	})
}

func testAccCheckUserRoleBindingDestroyed(userEmail string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		user, statusCode, err := client.GetUser(userEmail)
		if statusCode == 404 || user == nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error fetching user %q: %s", userEmail, err)
		}
		roles, _, err := client.GetUserRoles(user.Id)
		if err != nil {
			return fmt.Errorf("error fetching roles for user %q: %s", userEmail, err)
		}
		if len(roles) > 0 {
			return fmt.Errorf("user %q still has %d role(s) assigned after destroy", userEmail, len(roles))
		}
		return nil
	}
}

func testAccUserRoleBindingConfig(userEmail, roleName, bindingName string) string {
	return fmt.Sprintf(`
resource "komodor_user" "test" {
  email        = %q
  display_name = "Acc Test Binding User"
}

resource "komodor_role" "test" {
  name = %q
}

resource "komodor_user_role_binding" "test" {
  name    = %q
  user_id = komodor_user.test.id
  roles   = [komodor_role.test.id]
}
`, userEmail, roleName, bindingName)
}

func testAccUserRoleBindingConfigTwoRoles(userEmail, roleName, role2Name, bindingName string) string {
	return fmt.Sprintf(`
resource "komodor_user" "test" {
  email        = %q
  display_name = "Acc Test Binding User"
}

resource "komodor_role" "test" {
  name = %q
}

resource "komodor_role" "test2" {
  name = %q
}

resource "komodor_user_role_binding" "test" {
  name    = %q
  user_id = komodor_user.test.id
  roles   = [komodor_role.test.id, komodor_role.test2.id]
}
`, userEmail, roleName, role2Name, bindingName)
}
