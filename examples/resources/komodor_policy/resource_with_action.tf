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

resource "komodor_policy" "komo-example-policy" {
  name       = "komo-example-policy"
  statements = <<EOF
[{
  "actions": [
    "${komodor_action.komo-example-pod-viewer.action}"
  ],
  "resources": [{
    "cluster": "komo-example-cluster",
    "namespaces": [
      "default",
      "kube-system"
    ]
  }]
}]
EOF
}