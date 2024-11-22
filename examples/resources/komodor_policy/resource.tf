resource "komodor_policy" "my-policy" {
  name       = "my-policy"
  statements = <<EOF
[{
  "actions": [
    "get:daemonset",
    "edit:cronjob",
    "delete:service",
    "edit:job"
  ],
  "resources": [{
    "cluster": "kind-kind",
    "namespaces": [
      "default",
      "komodor"
    ]
  }]
}]
EOF
}