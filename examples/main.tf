terraform {
  required_providers {
    komodor = {
      version = ">= 1.0.6"
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

resource "komodor_monitor" "example-deploy-monitor" {
  name          = "example-deploy-monitor"
  type          = "deploy"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "exclude": {
    "namespaces": ["komodor"]
  },
  "namespaces": [
    "default"
  ]
}]
EOF
  sinks         = <<EOF
{
  "slack": [
    "default"
  ],
  "teams": [
    "default"
  ],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-pagerduty-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF 
  sinks_options = <<EOF
{
  "notifyOn": ["Failure"]
}
EOF 
}

resource "komodor_monitor" "example-availability-monitor" {
  name      = "example-availability-monitor"
  type      = "availability"
  active    = true
  sensors   = <<EOF
[{
  "cluster": "kind-kind",
  "exclude": {
    "services": ["default/service-to-exclude"]
  },
  "services": [
    "default/service-to-include"
  ],
  "condition": "and",
  "namespaces": ["default"]
}]
EOF
  sinks     = <<EOF
{
  "slack": [
    "default"
  ],
  "teams": [
    "default"
  ],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-pagerduty-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF  
  variables = <<EOF
{
  "categories": [
    "*"
  ],
  "duration": 30,
  "minAvailable": "100%"
}
EOF 
  sinks_options = <<EOF
{
  "notifyOn": ["*"]
}
EOF 
}

data "komodor_policy" "my-policy" {
  name = "default-read-only"
}


