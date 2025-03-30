resource "komodor_policy_v2" "simple_policy" {
  name = "simple-read-policy"

  statements {
    actions = ["view:all"]

    resources_scope {
      clusters   = ["prod-cluster"]
      namespaces = ["default", "kube-system"]
    }
  }
}

resource "komodor_policy_v2" "admin_policy" {
  name = "admin-policy"

  statements {
    actions = [
      "manage:kubeconfig",
      "view:audit",
      "manage:users",
      "manage:agents",
      "manage:account-access",
      "manage:trackedkeys"
    ]

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
