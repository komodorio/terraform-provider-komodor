resource "komodor_action" "komo-example-pod-viewer" {
  action      = "pod-viewer"
  description = "View pods"
  ruleset     = <<EOF
[
  {
    "apiGroups": [
      "apps"
    ],
    "resources": [
      "pods"
    ],
    "verbs": [
      "get",
      "list"
    ]
  }
]
EOF
}

resource "komodor_policy_v2" "komo-example-policy" {
  name = "komo-example-policy"

  statements {
    actions = [komodor_action.komo-example-pod-viewer.action]

    resources_scope {
      clusters   = ["komo-example-cluster"]
      namespaces = ["default", "kube-system"]
    }
  }
}
