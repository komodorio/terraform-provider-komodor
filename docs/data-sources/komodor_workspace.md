# Data Source: komodor_workspace

This data source allows you to retrieve information about an existing Komodor workspace.

## Example Usage

```hcl
data "komodor_workspace" "example" {
  id = "workspace-id"
}

# Use the workspace data in other resources
resource "some_other_resource" "example" {
  workspace_id = data.komodor_workspace.example.id
  workspace_name = data.komodor_workspace.example.name
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the workspace to retrieve.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `name` - The name of the workspace (maximum length: 127 characters).
* `description` - The description of the workspace (maximum length: 256 characters).
* `scopes` - A list of scopes that define which resources are included in the workspace. Each scope contains:
  * `clusters` - A list of cluster names included in the scope.
  * `namespaces` - A list of namespace names included in the scope.
  * `clusters_patterns` - A list of patterns used to match cluster names:
    * `include` - The pattern used to include matching clusters (maximum length: 128 characters).
    * `exclude` - The pattern used to exclude matching clusters (maximum length: 128 characters).
  * `namespaces_patterns` - A list of patterns used to match namespace names:
    * `include` - The pattern used to include matching namespaces (maximum length: 63 characters).
    * `exclude` - The pattern used to exclude matching namespaces (maximum length: 63 characters).
  * `selectors` - A list of selectors used to match resources:
    * `key` - The key of the label or annotation.
    * `type` - The type of selector, either "label" or "annotation".
    * `value` - The value to match (maximum length: 63 characters).
  * `selectors_patterns` - A list of pattern-based selectors:
    * `key` - The key of the label or annotation.
    * `type` - The type of selector, either "label" or "annotation".
    * `value` - The pattern used to match values:
      * `include` - The pattern used to include matching values (maximum length: 63 characters).
      * `exclude` - The pattern used to exclude matching values (maximum length: 63 characters).
* `created_at` - The timestamp when the workspace was created.
* `updated_at` - The timestamp when the workspace was last updated.
* `author_email` - The email of the user who created the workspace.
* `last_updated_by_email` - The email of the user who last updated the workspace. 