resource "komodor_policy_v2" "pattern_based_policy" {
  name = "pattern-read-policy"
  type = "v2"

  statements {
    actions = ["view:all"]

    resources_scope {
        clusters_patterns {
          include = "prod-*"
          exclude = "prod-legacy"
        }

        clusters_patterns {
          include = "staging-*"
          exclude = "staging-legacy"
        }
      

      namespaces_patterns {
          include = "team-*"
          exclude = "team-internal"
        }
    }
  }
}