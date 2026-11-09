package komodor

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	registerAccTest("komodor_policy_role_attachment")
}

func TestResourcePolicyRoleAttachmentSchema(t *testing.T) {
	resource := resourcePolicyRoleAttachment()

	if resource == nil {
		t.Fatal("resource should not be nil")
	}
	if resource.Importer == nil {
		t.Fatal("resource importer should be configured")
	}
	if schemaField, ok := resource.Schema["name"]; !ok {
		t.Fatal("name field should exist")
	} else {
		if schemaField.Required {
			t.Fatal("name should not be required")
		}
		if schemaField.Deprecated == "" {
			t.Fatal("name should be deprecated")
		}
	}
}

func TestPolicyRoleAttachmentStatePolicies(t *testing.T) {
	apiPolicies := []string{"policy-a", "policy-b", "policy-c"}
	configuredPolicies := []string{"policy-b"}

	if got := statePoliciesForRead(configuredPolicies, apiPolicies); len(got) != 1 || got[0] != "policy-b" {
		t.Fatalf("expected configured subset to be preserved, got %#v", got)
	}

	if got := statePoliciesForRead(nil, apiPolicies); !equalStringSlice(got, apiPolicies) {
		t.Fatalf("expected API policies to be used when no configured policies are present, got %#v", got)
	}

	if got := statePoliciesForRead([]string{"policy-b"}, []string{"policy-a"}); len(got) != 0 {
		t.Fatalf("expected policy detached out-of-band to be dropped from state so drift is detected, got %#v", got)
	}
}

func TestPolicyRoleAttachmentLegacyStateUpgrade(t *testing.T) {
	resource := resourcePolicyRoleAttachment()
	if len(resource.StateUpgraders) == 0 {
		t.Fatal("legacy state upgrader should be configured")
	}

	upgraded, err := resource.StateUpgraders[0].Upgrade(nil, map[string]interface{}{
		"id":       "legacy-name",
		"name":     "legacy-name",
		"role":     "role-123",
		"policies": []interface{}{"policy-1", "policy-2"},
	}, nil)
	if err != nil {
		t.Fatalf("expected upgrade to succeed: %v", err)
	}
	if got := upgraded["id"]; got != "role-123" {
		t.Fatalf("expected state id to be role-123 after upgrade, got %#v", got)
	}
	if got := upgraded["role"]; got != "role-123" {
		t.Fatalf("expected role to remain role-123 after upgrade, got %#v", got)
	}
}

func TestAcc_komodor_policy_role_attachment_basic(t *testing.T) {
	roleName := testResourceName("attach-role")
	policyName := testResourceName("attach-policy")
	resourceAddr := "komodor_policy_role_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Destroy is implicit: deleting the role and policy removes the attachment.
		Steps: []resource.TestStep{
			// Step 1: Create role, policy, and attachment
			{
				Config: testAccPolicyRoleAttachmentConfig(roleName, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "policies.#", "1"),
				),
			},
			// Step 2: Update — add a second policy
			{
				Config: testAccPolicyRoleAttachmentConfigTwoPolicies(roleName, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "policies.#", "2"),
				),
			},
		},
	})
}

func TestAcc_komodor_policy_role_attachment_multiple_resources_same_role(t *testing.T) {
	roleName := testResourceName("multi-role")
	policy1Name := testResourceName("policy-one")
	policy2Name := testResourceName("policy-two")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyRoleAttachmentConfigMultipleResourcesSameRole(roleName, policy1Name, policy2Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("komodor_policy_role_attachment.one", "policies.#", "1"),
					resource.TestCheckResourceAttr("komodor_policy_role_attachment.two", "policies.#", "1"),
				),
			},
		},
	})
}

func TestAcc_komodor_policy_role_attachment_multiple_roles_independent(t *testing.T) {
	role1Name := testResourceName("role-one")
	role2Name := testResourceName("role-two")
	policy1Name := testResourceName("policy-one")
	policy2Name := testResourceName("policy-two")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyRoleAttachmentConfigMultipleRolesIndependent(role1Name, role2Name, policy1Name, policy2Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("komodor_policy_role_attachment.one", "policies.#", "1"),
					resource.TestCheckResourceAttr("komodor_policy_role_attachment.two", "policies.#", "1"),
					resource.TestCheckResourceAttrPair("komodor_policy_role_attachment.one", "role", "komodor_role.one", "id"),
					resource.TestCheckResourceAttrPair("komodor_policy_role_attachment.two", "role", "komodor_role.two", "id"),
				),
			},
		},
	})
}

func testAccPolicyRoleAttachmentConfig(roleName, policyName string) string {
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
  role     = komodor_role.test.id
  policies = [komodor_policy_v2.test.id]
}
`, roleName, policyName)
}

func testAccPolicyRoleAttachmentConfigTwoPolicies(roleName, policyName string) string {
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
  role     = komodor_role.test.id
  policies = [komodor_policy_v2.test.id, komodor_policy_v2.test2.id]
}
`, roleName, policyName, policyName)
}

func testAccPolicyRoleAttachmentConfigMultipleResourcesSameRole(roleName, policy1Name, policy2Name string) string {
	return fmt.Sprintf(`
resource "komodor_role" "test" {
  name = %q
}

resource "komodor_policy_v2" "one" {
  name = %q

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["tf-acc-cluster"]
      namespaces = ["default"]
    }
  }
}

resource "komodor_policy_v2" "two" {
  name = %q

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

resource "komodor_policy_role_attachment" "one" {
  role     = komodor_role.test.id
  policies = [komodor_policy_v2.one.id]
}

resource "komodor_policy_role_attachment" "two" {
  role     = komodor_role.test.id
  policies = [komodor_policy_v2.two.id]
}
`, roleName, policy1Name, policy2Name)
}

func testAccPolicyRoleAttachmentConfigMultipleRolesIndependent(role1Name, role2Name, policy1Name, policy2Name string) string {
	return fmt.Sprintf(`
resource "komodor_role" "one" {
  name = %q
}

resource "komodor_role" "two" {
  name = %q
}

resource "komodor_policy_v2" "one" {
  name = %q

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["tf-acc-cluster"]
      namespaces = ["default"]
    }
  }
}

resource "komodor_policy_v2" "two" {
  name = %q

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

resource "komodor_policy_role_attachment" "one" {
  role     = komodor_role.one.id
  policies = [komodor_policy_v2.one.id]
}

resource "komodor_policy_role_attachment" "two" {
  role     = komodor_role.two.id
  policies = [komodor_policy_v2.two.id]
}
`, role1Name, role2Name, policy1Name, policy2Name)
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
