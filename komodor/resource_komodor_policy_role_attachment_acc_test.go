package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	registerAccTest("komodor_policy_role_attachment")
}

func TestAcc_komodor_policy_role_attachment_basic(t *testing.T) {
	roleName := testResourceName(t, "attach-role")
	policyName := testResourceName(t, "attach-policy")
	attachName := testResourceName(t, "attachment")
	resourceAddr := "komodor_policy_role_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Destroy is implicit: deleting the role and policy removes the attachment.
		Steps: []resource.TestStep{
			// Step 1: Create role, policy, and attachment
			{
				Config: testAccPolicyRoleAttachmentConfig(roleName, policyName, attachName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", attachName),
					resource.TestCheckResourceAttr(resourceAddr, "policies.#", "1"),
				),
			},
			// Step 2: Update — add a second policy
			{
				Config: testAccPolicyRoleAttachmentConfigTwoPolicies(roleName, policyName, attachName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", attachName),
					resource.TestCheckResourceAttr(resourceAddr, "policies.#", "2"),
				),
			},
		},
	})
}

func testAccPolicyRoleAttachmentConfig(roleName, policyName, attachName string) string {
	return fmt.Sprintf(`
resource "komodor_role" "test" {
  name = %q
}

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

resource "komodor_policy_role_attachment" "test" {
  name     = %q
  role     = komodor_role.test.id
  policies = [komodor_policy_v2.test.id]
}
`, roleName, policyName, attachName)
}

func testAccPolicyRoleAttachmentConfigTwoPolicies(roleName, policyName, attachName string) string {
	return fmt.Sprintf(`
resource "komodor_role" "test" {
  name = %q
}

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

resource "komodor_policy_v2" "test2" {
  name = "%s-2"

  statements {
    actions = ["manage:users"]

    resources_scope {
      clusters_patterns {
        include = "*"
        exclude = ""
      }
      namespaces_patterns {
        include = "*"
        exclude = ""
      }
    }
  }
}

resource "komodor_policy_role_attachment" "test" {
  name     = %q
  role     = komodor_role.test.id
  policies = [komodor_policy_v2.test.id, komodor_policy_v2.test2.id]
}
`, roleName, policyName, policyName, attachName)
}
