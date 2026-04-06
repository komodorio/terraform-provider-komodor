package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_monitor")
}

func TestAcc_komodor_monitor_availability(t *testing.T) {
	name := testResourceName("monitor")
	updatedName := name + "-updated"
	resourceAddr := "komodor_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMonitorDestroyed(name),
		Steps: []resource.TestStep{
			// Step 1: Create an availability monitor
			{
				Config: testAccMonitorConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "type", "availability"),
					resource.TestCheckResourceAttr(resourceAddr, "active", "true"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
			// Step 2: Update — rename and disable
			{
				Config: testAccMonitorConfigUpdated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", updatedName),
					resource.TestCheckResourceAttr(resourceAddr, "active", "false"),
				),
			},
			// Step 3: Import
			{
				ResourceName:      resourceAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMonitorDestroyed(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		monitors, err := client.GetMonitors()
		if err != nil {
			return fmt.Errorf("error listing monitors during destroy check: %s", err)
		}
		for _, m := range monitors {
			if m.IsDeleted != nil && *m.IsDeleted {
				continue
			}
			if m.Name != nil && *m.Name == name {
				return fmt.Errorf("monitor %q still exists after destroy", name)
			}
		}
		return nil
	}
}

func testAccMonitorConfig(name string) string {
	return fmt.Sprintf(`
resource "komodor_monitor" "test" {
  name   = %q
  type   = "availability"
  active = true

  sensors = jsonencode([{
    cluster    = "tf-acc-cluster"
    namespaces = ["default"]
  }])

  variables = jsonencode({
    categories   = ["*"]
    duration     = 5
    minAvailable = "1"
  })
}
`, name)
}

func testAccMonitorConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "komodor_monitor" "test" {
  name   = %q
  type   = "availability"
  active = false

  sensors = jsonencode([{
    cluster    = "tf-acc-cluster"
    namespaces = ["default", "kube-system"]
  }])

  variables = jsonencode({
    categories   = ["*"]
    duration     = 10
    minAvailable = "1"
  })
}
`, name)
}
