package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	registerAccTest("datasource_komodor_policy_v2")
}

func TestAcc_datasource_komodor_policy_v2(t *testing.T) {
	name := testResourceName(t, "ds-policy")
	resourceAddr := "data.komodor_policy_v2.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourcePolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func testAccDatasourcePolicyConfig(name string) string {
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

data "komodor_policy_v2" "test" {
  name       = komodor_policy_v2.test.name
  depends_on = [komodor_policy_v2.test]
}
`, name)
}
