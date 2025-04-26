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

# Example 1: Get information about an existing workspace
data "komodor_workspace" "example" {
  id = var.workspace_id
}

# Example 2: Use workspace data in another resource
resource "komodor_monitor" "example" {
  name        = "monitor-in-workspace"
  description = "A monitor in the workspace"
  workspace_id = data.komodor_workspace.example.id
  # ... other monitor configuration
}

# Output the workspace details
output "workspace_name" {
  value = data.komodor_workspace.example.name
}

output "workspace_description" {
  value = data.komodor_workspace.example.description
}

output "workspace_scopes" {
  value = data.komodor_workspace.example.scopes
} 