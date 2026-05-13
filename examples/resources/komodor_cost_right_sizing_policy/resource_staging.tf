resource "komodor_cost_right_sizing_policy" "staging" {
  # step 1 - name
  name        = "staging-default"
  description = "Default right-sizing policy for staging workloads"
  priority    = 200

  # step 2 - scope
  scope {
    clusters_patterns {
      include = "staging-*"
      exclude = "staging-*-canary"
    }
    namespaces     = ["payments", "checkout", "auth"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names_patterns {
      include = "*"
    }
  }

  # step 3 - when to apply
  apply_protocol         = "onCreation"
  allow_restart          = true
  allow_hpa_right_sizing = false

  # step 4 - guardrails
  optimization_preset = "staging"
}
