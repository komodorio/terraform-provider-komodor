resource "komodor_policy_v2" "selector_pattern_policy" {
  name = "selector-pattern-view-policy"

  statements {
    actions = ["view:nodes", "view:namespaces"]

    resources_scope {
      clusters   = ["prod-cluster"]
      namespaces = ["default"]

      selectors_patterns {
        key  = "team"
        type = "annotation"
        value {
          include = "team-*"
          exclude = "team-internal"
        }
      }

      selectors_patterns {
        key  = "env"
        type = "label"
        value {
          include = "production"
          exclude = ""
        }
      }
    }
  }
}
