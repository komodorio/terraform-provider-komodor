package komodor

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_klaudia_skill")
}

var accTestKlaudiaSkillID string

func TestAcc_komodor_klaudia_skill_basic(t *testing.T) {
	name := testResourceName("klaudia-skill")
	resourceAddr := "komodor_klaudia_skill.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKlaudiaSkillDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccKlaudiaSkillConfig(name, "acceptance test skill", "Initial instructions for the skill.", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "description", "acceptance test skill"),
					resource.TestCheckResourceAttr(resourceAddr, "is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceAddr, "clusters.#", "1"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					testAccCaptureKlaudiaSkillID(resourceAddr),
				),
			},
			{
				Config: testAccKlaudiaSkillConfig(name, "acceptance test skill updated", "Updated instructions for the skill.", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "description", "acceptance test skill updated"),
					resource.TestCheckResourceAttr(resourceAddr, "instructions", "Updated instructions for the skill."),
					resource.TestCheckResourceAttr(resourceAddr, "is_enabled", "true"),
					testAccCaptureKlaudiaSkillID(resourceAddr),
				),
			},
		},
	})
}

func testAccCaptureKlaudiaSkillID(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("not found: %s", addr)
		}
		accTestKlaudiaSkillID = rs.Primary.ID
		return nil
	}
}

func testAccCheckKlaudiaSkillDestroyed(s *terraform.State) error {
	id := accTestKlaudiaSkillID
	accTestKlaudiaSkillID = ""
	if id == "" {
		return nil
	}
	client := testAccProvider.Meta().(*Client)
	_, sc, _ := client.GetSkill(id)
	if sc == http.StatusNotFound {
		return nil
	}
	if sc == http.StatusOK {
		return fmt.Errorf("Klaudia skill %q still exists after destroy", id)
	}
	return nil
}

func testAccKlaudiaSkillConfig(name, description, instructions string, isEnabled bool) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_skill" "test" {
  name         = %q
  description  = %q
  instructions = %q
  clusters     = ["*"]
  is_enabled   = %t
}
`, name, description, instructions, isEnabled)
}
