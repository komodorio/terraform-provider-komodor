# Resource: komodor_workspace

This resource allows you to create and manage Komodor workspaces. Workspaces are used to organize and scope your Kubernetes resources in Komodor.

> **Note**: When using selectors and selectors patterns in workspaces, make sure to configure the tracked keys in the [Settings Page](https://app.komodor.com/settings/tracked-keys). For more information about workspace creation and optimization, see the [Komodor documentation](https://help.komodor.com/hc/en-us/articles/25537329198866-Workspaces-Creation-Optimisation).

## Example Usage

```hcl
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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the workspace. Maximum length is 127 characters.
* `description` - (Optional) A description of the workspace. Maximum length is 256 characters.
* `scopes` - (Required) A list of scopes that define which resources are included in the workspace. Each scope can contain:
  * `clusters` - (Optional) A list of cluster names to include.
  * `namespaces` - (Optional) A list of namespace names to include.
  * `clusters_patterns` - (Optional) A list of patterns to match cluster names:
    * `include` - (Required) A pattern to include matching clusters. Maximum length is 128 characters.
    * `exclude` - (Required) A pattern to exclude matching clusters. Maximum length is 128 characters.
  * `namespaces_patterns` - (Optional) A list of patterns to match namespace names:
    * `include` - (Required) A pattern to include matching namespaces. Maximum length is 63 characters.
    * `exclude` - (Required) A pattern to exclude matching namespaces. Maximum length is 63 characters.
  * `selectors` - (Optional) A list of selectors to match resources:
    * `key` - (Required) The key of the label or annotation.
    * `type` - (Required) The type of selector, either "label" or "annotation".
    * `value` - (Required) The value to match. Maximum length is 63 characters.
  * `selectors_patterns` - (Optional) A list of pattern-based selectors:
    * `key` - (Required) The key of the label or annotation.
    * `type` - (Required) The type of selector, either "label" or "annotation".
    * `value` - (Required) A pattern to match:
      * `include` - (Required) A pattern to include matching values. Maximum length is 63 characters.
      * `exclude` - (Required) A pattern to exclude matching values. Maximum length is 63 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the workspace.
* `created_at` - The timestamp when the workspace was created.
* `updated_at` - The timestamp when the workspace was last updated.
* `author_email` - The email of the user who created the workspace.
* `last_updated_by_email` - The email of the user who last updated the workspace.

## Import

Workspaces can be imported using their ID:

```bash
terraform import komodor_workspace.example workspace-id
``` 