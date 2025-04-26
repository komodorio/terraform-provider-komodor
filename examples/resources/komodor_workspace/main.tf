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

# Example 1: Basic workspace with specific clusters and namespaces
resource "komodor_workspace" "basic" {
  name        = "basic-workspace"
  description = "A workspace for monitoring specific resources"

  scopes {
    clusters = ["cluster-1", "cluster-2"]
    namespaces = ["default", "kube-system"]
  }
}

# Example 2: Workspace with pattern-based scopes
resource "komodor_workspace" "pattern" {
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

# Example 3: Workspace with selector-based scopes
resource "komodor_workspace" "selector" {
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