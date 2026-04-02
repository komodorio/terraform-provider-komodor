resource "komodor_monitor" "example-pvc-monitor" {
  name      = "example-pvc-monitor"
  type      = "PVC"
  active    = true
  sensors   = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["storage-namespace"]
}]
EOF
  sinks     = <<EOF
{
  "slack": ["storage-alerts"],
  "teams": ["Storage-Team"],
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