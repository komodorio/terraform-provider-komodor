# Example: Complete User Role Binding Setup
# This example demonstrates how to create users, roles, and bind them together

# Create a user
resource "komodor_user" "developer" {
  display_name = "Jane Developer"
  email        = "jane.developer@example.com"
}

# Create a role
resource "komodor_role" "developer_role" {
  name = "developer"
}

# Create a policy
resource "komodor_policy_v2" "developer_policy" {
  name        = "developer-policy"
  description = "Policy for developers"

  statement {
    actions = [
      "k8s:DescribeResource",
      "k8s:GetLogs"
    ]

    resource_scope {
      clusters            = ["production"]
      namespaces          = ["backend", "frontend"]
      clusters_patterns   = []
      namespaces_patterns = []
      selectors           = []
      selectors_patterns  = []
    }
  }
}

# Attach policy to role
resource "komodor_policy_role_attachment" "developer_policy_attachment" {
  name     = "developer-policy-attachment"
  role     = komodor_role.developer_role.id
  policies = [komodor_policy_v2.developer_policy.id]
}

# Bind user to role
resource "komodor_user_role_binding" "developer_binding" {
  name    = "jane-developer-binding"
  user_id = komodor_user.developer.id
  roles   = [komodor_role.developer_role.id]
}

# Example with multiple roles and expiration
resource "komodor_user_role_binding" "contractor_binding" {
  name    = "contractor-multi-role-binding"
  user_id = "contractor@example.com"
  roles = [
    komodor_role.developer_role.id,
    "viewer-role-id" # Can also use existing role IDs
  ]

}
