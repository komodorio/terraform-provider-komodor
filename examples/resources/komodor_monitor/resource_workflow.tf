resource "komodor_monitor" "example-workflow-monitor" {
  name          = "example-workflow-monitor"
  type          = "workflow"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["workflow-namespace"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["workflow-alerts"],
  "webhook": ["webhook-url"]
}
EOF
}