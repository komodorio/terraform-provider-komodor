resource "komodor_policy_v2" "my-policy" {
  name = "my-policy"

  statements {
    actions = ["get:daemonset", "edit:cronjob", "delete:service", "edit:job"]

    resources_scope {
      clusters   = ["kind-kind"]
      namespaces = ["default", "komodor"]
    }
  }
}

resource "komodor_role" "my-role" {
  name = "my-role"
}

resource "komodor_policy_role_attachment" "my-attachement" {
  name     = "test-attachement"
  policies = [komodor_policy_v2.my-policy.id]
  role     = komodor_role.my-role.id
}
