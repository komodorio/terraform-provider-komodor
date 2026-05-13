resource "komodor_cost_right_sizing_policy" "multi_scope" {
  name        = "multi-scope-example"
  description = "Right-sizing policy with multiple OR-evaluated scope statements"
  priority    = 500

  scope {
    # statement 1 — exact production EU clusters + exact namespaces
    clusters       = ["prod-eu-west-1", "prod-eu-central-1"]
    namespaces     = ["payments", "checkout"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names_patterns {
      include = "*"
    }
  }

  scope {
    # statement 2 — cluster + namespace patterns with canary exclusion
    clusters_patterns {
      include = "staging-*"
      exclude = "staging-*-canary"
    }
    namespaces_patterns {
      include = "team-*"
    }
    resource_types = ["Deployment"]
    workload_names_patterns {
      include = "*"
    }
  }

  apply_protocol         = "onCreation"
  allow_restart          = true
  allow_hpa_right_sizing = false

  percentile          = 95
  optimization_preset = "production"
  allow_qos_upgrade   = true
  allow_qos_downgrade = false
}
