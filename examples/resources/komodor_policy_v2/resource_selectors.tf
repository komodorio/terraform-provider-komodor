resource "komodor_policy_v2" "selector_based_policy" {
  name = "selector-view-policy"

  statements {
    actions = ["view:nodes", "view:namespaces"]

    resources_scope {
      clusters   = ["prod-cluster"]
      namespaces = ["default"]

      selectors {
        key   = "team"
        type  = "annotation"
        value = "platform"
      }

      selectors {
        key   = "env"
        type  = "label"
        value = "production"
      }
    }
  }
}
