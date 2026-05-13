resource "komodor_cost_right_sizing_policy" "production" {
  # step 1 - name
  name        = "production-default"
  description = "Default right-sizing policy for production workloads"
  priority    = 300

  # step 2 - scope
  scope {
    clusters       = ["prod-us-east-1", "prod-eu-west-1", "prod-ap-southeast-2"]
    namespaces     = ["payments", "checkout", "auth", "api"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names = ["*"]
  }

  # step 3 - when to apply
  apply_protocol         = "onCreation"
  allow_restart          = true
  allow_hpa_right_sizing = false

  # step 4 - guardrails
  percentile          = 99
  optimization_preset = "production"
  allow_qos_upgrade   = true
  allow_qos_downgrade = false
}
