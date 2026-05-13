# Look up an existing right-sizing policy by name. Only top-level scalar
# attributes are exposed — scope and guardrails are not surfaced here.
data "komodor_cost_right_sizing_policy" "source" {
  name = "production-default"
}

# Create a new policy that copies the source's behavior fields but targets a
# different cluster set. scope is required and must be specified explicitly.
# guardrails is omitted here because the source uses a named preset; switch
# optimization_preset to "custom" and add an explicit guardrails block if you
# want to override the preset's values.
resource "komodor_cost_right_sizing_policy" "production_eu" {
  name        = "production-eu"
  description = "EU mirror of ${data.komodor_cost_right_sizing_policy.source.name}"
  priority    = data.komodor_cost_right_sizing_policy.source.priority + 1

  scope {
    clusters       = ["prod-eu-west-1", "prod-eu-central-1"]
    namespaces     = ["payments", "checkout", "auth"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names_patterns {
      include = "*"
    }
  }

  apply_protocol         = data.komodor_cost_right_sizing_policy.source.apply_protocol
  allow_restart          = data.komodor_cost_right_sizing_policy.source.allow_restart
  allow_hpa_right_sizing = data.komodor_cost_right_sizing_policy.source.allow_hpa_right_sizing

  optimization_preset = data.komodor_cost_right_sizing_policy.source.optimization_preset
}
