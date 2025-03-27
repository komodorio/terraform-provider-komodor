resource "komodor_policy_v2" "simple_policy" {
  name = "simple-read-policy"
  type = "v2"

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["prod-cluster"]
      namespaces = ["default", "kube-system"]
    }
  }
}

