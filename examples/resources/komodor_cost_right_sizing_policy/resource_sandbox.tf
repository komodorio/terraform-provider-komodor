resource "komodor_cost_right_sizing_policy" "sandbox" {
  # step 1 - name
  name        = "sandbox-aggressive"
  description = "Aggressive right-sizing for sandbox/PR-preview clusters"
  priority    = 10

  # step 2 - scope
  scope {
    clusters_patterns {
      include = "sandbox-*"
    }
    namespaces     = ["*"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names = ["*"]
  }

  # step 3 - when to apply
  apply_protocol         = "immediate"
  allow_restart          = false
  allow_hpa_right_sizing = false

  # step 4 - guardrails
  percentile          = 70
  optimization_preset = "sandbox"
  allow_qos_upgrade   = false
  allow_qos_downgrade = false
}
