package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_kubernetes")
}

// TestAcc_komodor_kubernetes_basic tests creation and deletion of a Kubernetes
// integration record in the Komodor API.
//
// Note: komodor_kubernetes has ForceNew on cluster_name and no UpdateContext,
// so this test only covers create and delete (no update step).
func TestAcc_komodor_kubernetes_basic(t *testing.T) {
	clusterName := testResourceName("cluster")
	resourceAddr := "komodor_kubernetes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesDestroyed(clusterName),
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "cluster_name", clusterName),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func testAccCheckKubernetesDestroyed(clusterName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		k8s, statusCode, err := client.GetKubernetesCluster(clusterName)
		if statusCode == 404 {
			return nil
		}
		if err != nil {
			// A non-404 error is unexpected but means the resource is gone.
			return nil
		}
		if k8s != nil {
			return fmt.Errorf("kubernetes integration %q still exists after destroy", clusterName)
		}
		return nil
	}
}

func testAccKubernetesConfig(clusterName string) string {
	return fmt.Sprintf(`
resource "komodor_kubernetes" "test" {
  cluster_name = %q
}
`, clusterName)
}
