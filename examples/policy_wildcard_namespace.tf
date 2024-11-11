# This example shows how to create a policy with a wildcard namespace pattern.
# wildcard policy type is not available by default.
# When the feature is disabled, applying this policy will fail with error: `400 Bad Request`

resource "komodor_policy" "komo-example-wildcard-policy" {
  name       = "komo-example-wildcard-policy"
  type       = "wildcard"
  statements = <<EOF
[{
  "actions": [F
    "view:all"
  ],
  "resources": [{
    "cluster": "komo-example-cluster",
    "namespacePattern": "prod-*"
  }]
}]
EOF
}