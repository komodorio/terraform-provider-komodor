# This example shows how to create a policy with dynamic tags.
# Dynamic tags feature is not available by default.
# When the feature is disabled, applying this policy will fail with error: `400 Bad Request`

resource "komodor_policy" "komo-example-dynamic-tags-policy" {
  name = "komo-example-dynamic-tags-policy"
  type = "dynamic_tag"
  tags = {
    "team" : "super-heroes"
  }
  statements = <<EOF
[{
  "actions": [
    "view:all"
  ],
  "resources": []
}]
EOF
}