resource "komodor_monitor" "example-availability-monitor" {
  name          = "example-availability-monitor"
  type          = "availability"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "exclude": {
    "services": ["default/excluded-service"]
  },
  "services": ["default/important-service"],
  "condition": "and",
  "namespaces": ["default"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["availability-alerts"],
  "teams": ["SRE-Team"]
}
EOF
  variables     = <<EOF
{
  "categories": ["Creating/Initializing", "Unhealthy - failed probes"],
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