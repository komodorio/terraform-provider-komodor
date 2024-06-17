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