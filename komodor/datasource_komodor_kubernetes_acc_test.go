package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() { registerAccTest("datasource_komodor_kubernetes") }

func TestAcc_datasource_komodor_kubernetes(t *testing.T) {
	clusterName := testResourceName(t, "ds-cluster")
	resourceAddr := "data.komodor_kubernetes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceKubernetesConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "cluster_name", clusterName),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func testAccDatasourceKubernetesConfig(clusterName string) string {
	return fmt.Sprintf(`
resource "komodor_kubernetes" "test" {
  cluster_name = %q
}

data "komodor_kubernetes" "test" {
  cluster_name = komodor_kubernetes.test.cluster_name
  depends_on   = [komodor_kubernetes.test]
}
`, clusterName)
}
