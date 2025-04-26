# Komodor Workspace Examples

This directory contains examples of how to use the Komodor workspace resource in Terraform.

## Examples

### Basic Workspace
The `main.tf` file contains three examples:

1. **Basic Workspace**: Creates a workspace with specific cluster and namespace scopes
   ```hcl
   resource "komodor_workspace" "example" {
     name        = "example-workspace"
     description = "A workspace for monitoring specific resources"
     scopes {
       clusters = ["cluster-1"]
       namespaces = ["default", "kube-system"]
     }
   }
   ```

2. **Pattern-based Workspace**: Creates a workspace using pattern-based scopes
   ```hcl
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
   ```

3. **Selector-based Workspace**: Creates a workspace using selector-based scopes
   ```hcl
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
   ```

## Usage

1. Set your Komodor API key:
   ```bash
   export TF_VAR_komodor_api_key="your-api-key"
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

## Data Source Example

You can also use the workspace data source to reference existing workspaces:

```hcl
data "komodor_workspace" "existing" {
  id = "workspace-id"
}

# Use the workspace data in other resources
resource "some_other_resource" "example" {
  workspace_id = data.komodor_workspace.existing.id
}
``` 