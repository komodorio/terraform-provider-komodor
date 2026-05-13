resource "komodor_cost_right_sizing_policy" "development" {
  # step 1 - name
  name        = "dev-default"
  description = "Default right-sizing policy for development workloads"
  priority    = 100

  # step 2 - scope
  scope {
    clusters_patterns {
      include = "dev-*"
    }
    namespaces_patterns {
      include = "team-*"
      exclude = "team-*-experimental"
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
  percentile          = 80
  optimization_preset = "development"
  allow_qos_upgrade   = false
  allow_qos_downgrade = false
}
