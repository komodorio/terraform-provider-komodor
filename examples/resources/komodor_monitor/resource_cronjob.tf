resource "komodor_monitor" "example-cronjob-monitor" {
  name      = "example-cronjob-monitor"
  type      = "cronJob"
  active    = true
  sensors   = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["jobs-namespace"]
}]
EOF
  sinks     = <<EOF
{
  "slack": ["cronjob-alerts"],
  "teams": ["SRE-Team"],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF
  variables = <<EOF
{
  "duration": 120,
  "cronJobCondition": "first"
}
EOF
}