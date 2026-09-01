package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() { registerAccTest("datasource_komodor_role") }

func TestAcc_datasource_komodor_role(t *testing.T) {
	name := testResourceName(t, "ds-role")
	resourceAddr := "data.komodor_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_at"),
				),
			},
		},
	})
}

func testAccDatasourceRoleConfig(name string) string {
	return fmt.Sprintf(`
resource "komodor_role" "test" {
  name = %q
}

data "komodor_role" "test" {
  name       = komodor_role.test.name
  depends_on = [komodor_role.test]
}
`, name)
}
