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
    namespaces_patterns {
      include = "*"
    }
    resource_types = ["Deployment", "StatefulSet"]
    workload_names_patterns {
      include = "*"
    }
  }

  # step 3 - when to apply
  apply_protocol         = "immediate"
  allow_restart          = false
  allow_hpa_right_sizing = false

  # step 4 - guardrails
  optimization_preset = "sandbox"
}
