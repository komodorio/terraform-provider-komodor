resource "komodor_monitor" "example-node-monitor" {
  name      = "example-node-monitor"
  type      = "node"
  active    = true
  sensors   = <<EOF
[{
  "cluster": "kind-kind"
}]
EOF
  sinks     = <<EOF
{
  "slack": ["node-alerts"]
}
EOF
  variables = <<EOF
{
  "duration": 60,
  "nodeCreationThreshold": "10m"
}
EOF
}