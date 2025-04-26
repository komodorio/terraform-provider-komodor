terraform {
  required_providers {
    komodor = {
      source = "komodorio/komodor"
    }
  }
}

provider "komodor" {
  api_key = var.komodor_api_key
}

# Create a workspace for a specific cluster and namespace
resource "komodor_workspace" "example" {
  name        = "example-workspace"
  description = "A workspace for monitoring specific resources"

  scopes {
    clusters = ["cluster-1"]
    namespaces = ["default", "kube-system"]
  }
}

# Create a workspace with pattern-based scopes
resource "komodor_workspace" "pattern_based" {
  name        = "pattern-workspace"
  description = "A workspace using pattern-based scopes"

  scopes {
    clusters_patterns {
      include = "prod-*"
      exclude = "prod-backup-*"
    }
    namespaces_patterns {
      include = "app-*"
      exclude = "app-test-*"
    }
  }
}

# Create a workspace with selector-based scopes
resource "komodor_workspace" "selector_based" {
  name        = "selector-workspace"
  description = "A workspace using selector-based scopes"

  scopes {
    selectors {
      key   = "environment"
      type  = "label"
      value = "production"
    }
    selectors_patterns {
      key  = "team"
      type = "label"
      value {
        include = "team-*"
        exclude = "team-ops-*"
      }
    }
  }
} 