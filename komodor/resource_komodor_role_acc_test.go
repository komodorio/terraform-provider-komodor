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

func TestAcc_komodor_role_updatePreservesID(t *testing.T) {
	name := testResourceName(t, "role-persist-id")
	updatedName := name + "-updated"
	resourceAddr := "komodor_role.test"
	var roleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroyed(),
		Steps: []resource.TestStep{
			// Step 1 - Create the role
			{
				Config: testAccRoleConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "is_default", "false"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					testAccCaptureResourceID(resourceAddr, &roleID),
					// Ensure we have a created_at timestamp, which also indicates the role was created successfully in the API and not just in state
					resource.TestCheckResourceAttrSet(resourceAddr, "created_at"),
				),
			},
			// Step 2 - Update the role - Ensure the ID remains the same after update
			{
				// PreConfig: func() { time.Sleep(2 * time.Second) },
				Config: testAccRoleConfig(updatedName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", updatedName),
					resource.TestCheckResourceAttr(resourceAddr, "is_default", "false"),
					testAccCheckResourceIDEquals(resourceAddr, &roleID),
				),
			},
		},
	})
}

func TestAcc_komodor_role_isDefaultLifecycle(t *testing.T) {
	name := testResourceName(t, "role-default")
	resourceAddr := "komodor_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoleDestroyed(),
		Steps: []resource.TestStep{
			// Step 1 - Create the role with is_default = false
			{
				Config: testAccRoleConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "is_default", "false"),
				),
			},
			// Step 2 - Update the role to is_default = true
			{
				Config: testAccRoleConfig(name, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "is_default", "true"),
				),
			},
			// Step 3 - Update the role back to is_default = false
			{
				Config: testAccRoleConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "is_default", "false"),
				),
			},
		},
	})
}

func testAccCaptureResourceID(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %q ID is empty", resourceName)
		}

		*id = rs.Primary.ID
		return nil
	}
}

func testAccCheckResourceIDEquals(resourceName string, expectedID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %q ID is empty", resourceName)
		}
		if *expectedID == "" {
			return fmt.Errorf("expected ID was not captured before update")
		}
		if rs.Primary.ID != *expectedID {
			return fmt.Errorf("resource %q ID changed across update: expected %q, got %q", resourceName, *expectedID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRoleDestroyed() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("[DEBUG] - Checking role destruction....\n")
		client := testAccProvider.Meta().(*Client)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "komodor_role" {
				fmt.Printf("[DEBUG] Resource: %v", rs)
				continue
			}

			role, statusCode, err := client.GetRole(rs.Primary.ID)
			fmt.Printf("[DEBUG] - Role ID: %s, Status Code: %d, Role: %+v\n", rs.Primary.ID, statusCode, role)
			if statusCode == 404 || role == nil {
				continue
			}
			if err != nil {
				return fmt.Errorf("error checking role destruction: %s", err)
			}

			return fmt.Errorf("role %q still exists after destroy", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRoleConfig(name string, isDefault bool) string {
	return fmt.Sprintf(`
resource "komodor_role" "test" {
  name       = %q
  is_default = %t
}
`, name, isDefault)
}
