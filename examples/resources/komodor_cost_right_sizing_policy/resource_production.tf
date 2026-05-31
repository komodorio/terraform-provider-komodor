resource "komodor_cost_right_sizing_policy" "production" {
  # step 1 - name
  name        = "production-default"
  description = "Default right-sizing policy for production workloads"
  priority    = 300

  # step 2 - scope
  #
  # Multiple `scope` blocks are evaluated with OR semantics: a workload is
  # in-scope if it matches ANY block. This lets a single policy cover a broad
  # set AND carve in specific extras that wouldn't be matched by the first.

  # Primary scope — Deployments and StatefulSets in the four service
  # namespaces across all three prod clusters.
  scope {
    clusters       = ["prod-us-east-1", "prod-eu-west-1", "prod-ap-southeast-2"]
    namespaces     = ["payments", "checkout", "auth", "api"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names_patterns {
      include = "*"
    }
  }

  # Carve-in — the `metrics-collector` DaemonSet in `monitoring-prod`,
  # which is outside the namespace list above but should be governed by
  # the same policy.
  scope {
    clusters       = ["prod-us-east-1", "prod-eu-west-1", "prod-ap-southeast-2"]
    namespaces     = ["monitoring-prod"]
    resource_types = ["DaemonSet"]
    workload_names = ["metrics-collector"]
  }

  # step 3 - when to apply
  apply_protocol         = "onCreation"
  allow_restart          = true
  allow_hpa_right_sizing = false

  # step 4 - guardrails
  optimization_preset = "production"

  # `force_delete = true` lets `terraform destroy` cascade-delete the
  # workload override records this policy produces. Recommended for any
  # policy with broad scope or `apply_protocol = "immediate"` — without it,
  # destroy can return `POLICY_HAS_OVERRIDES` (HTTP 409).
  force_delete = true
}
