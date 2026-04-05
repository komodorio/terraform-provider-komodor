resource "komodor_monitor" "example-job-monitor" {
  name      = "example-job-monitor"
  type      = "job"
  active    = true
  sensors   = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["job-namespace"]
}]
EOF
  sinks     = <<EOF
{
  "slack": ["job-alerts"],
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
  "duration": 300
}
EOF
}