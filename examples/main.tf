terraform {
  required_providers {
    komodor = {
      version = "1.0.0"
      source  = "komodorio/komodor"
    }
  }
}

provider "komodor" {
  api_key = var.api_key
}

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

resource "komodor_role" "my-role" {
  name = "my-role"
}

resource "komodor_policy_role_attachment" "my-attachement" {
  name     = "test-attachement"
  policies = [komodor_policy.my-policy.id]
  role     = komodor_role.my-role.id
}
