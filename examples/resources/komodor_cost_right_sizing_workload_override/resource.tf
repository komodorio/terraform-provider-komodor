resource "komodor_cost_right_sizing_policy" "prod" {
  name                = "prod-defaults"
  priority            = 100
  apply_protocol      = "onCreation"
  optimization_preset = "production"

  scope {
    clusters_patterns {
      include = "prod-*"
    }
    namespaces_patterns {
      include = "*"
    }
    workload_names_patterns {
      include = "*"
    }
  }
}

# include: pin one workload to a specific policy
resource "komodor_cost_right_sizing_workload_override" "pin_checkout" {
  action       = "include"
  cluster_name = "prod-us-east-1"
  namespace    = "payments"
  kind         = "Deployment"
  name         = "checkout-api"
  policy_id    = komodor_cost_right_sizing_policy.prod.id
}

# exclude: remove a workload from right-sizing automation entirely
resource "komodor_cost_right_sizing_workload_override" "exclude_legacy_db" {
  action       = "exclude"
  cluster_name = "prod-us-east-1"
  namespace    = "legacy"
  kind         = "StatefulSet"
  name         = "old-oracle"
}
