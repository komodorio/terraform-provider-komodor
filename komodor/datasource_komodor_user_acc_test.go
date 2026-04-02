package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() { registerAccTest("datasource_komodor_user") }

func TestAcc_datasource_komodor_user(t *testing.T) {
	email := accTestPrefix + "ds-user@komodor-test.com"
	resourceAddr := "data.komodor_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserConfig(email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "email", email),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttrSet(resourceAddr, "display_name"),
				),
			},
		},
	})
}

func testAccDatasourceUserConfig(email string) string {
	return fmt.Sprintf(`
resource "komodor_user" "test" {
  email        = %q
  display_name = "Acc Test DS User"
}

data "komodor_user" "test" {
  email        = komodor_user.test.email
  display_name = komodor_user.test.display_name
  depends_on   = [komodor_user.test]
}
`, email)
}
