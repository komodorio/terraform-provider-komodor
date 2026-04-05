package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_user")
}

func TestAcc_komodor_user_basic(t *testing.T) {
	email := accTestPrefix + "user@komodor-test.com"
	resourceAddr := "komodor_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserDestroyed(email),
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccUserConfig(email, "Acc Test User"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "email", email),
					resource.TestCheckResourceAttr(resourceAddr, "display_name", "Acc Test User"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
			// Step 2: Update display_name (email is ForceNew so must stay the same)
			{
				Config: testAccUserConfig(email, "Acc Test User Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "display_name", "Acc Test User Updated"),
					resource.TestCheckResourceAttr(resourceAddr, "email", email),
				),
			},
		},
	})
}

func testAccCheckUserDestroyed(email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		user, statusCode, err := client.GetUser(email)
		if statusCode == 404 || user == nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error checking user destruction: %s", err)
		}
		return fmt.Errorf("user %q still exists after destroy", email)
	}
}

func testAccUserConfig(email, displayName string) string {
	return fmt.Sprintf(`
resource "komodor_user" "test" {
  email        = %q
  display_name = %q
}
`, email, displayName)
}
