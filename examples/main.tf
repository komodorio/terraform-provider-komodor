resource "komodor_role" "my-role" {
  name = "my-role"
}

resource "komodor_policy_v2" "my-policy" {
  name = "view-all"

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["management"]
      namespaces = ["default", "komodor"]
    }
  }

resource "komodor_policy_role_attachment" "my-attachement" {
  name     = "test-attachement"
  policies = [komodor_policy_v2.my-policy.id]
  role     = komodor_role.my-role.id
}
