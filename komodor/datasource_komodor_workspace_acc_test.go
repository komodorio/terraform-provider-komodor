package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() { registerAccTest("datasource_komodor_workspace") }

func TestAcc_datasource_komodor_workspace(t *testing.T) {
	name := testResourceName(t, "ds-workspace")
	resourceAddr := "data.komodor_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceWorkspaceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func testAccDatasourceWorkspaceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodor_workspace" "test" {
  name = %q
  scopes {
    clusters   = ["tf-acc-cluster-1"]
    namespaces = ["default"]
  }
}

data "komodor_workspace" "test" {
  id         = komodor_workspace.test.id
  depends_on = [komodor_workspace.test]
}
`, name)
}
