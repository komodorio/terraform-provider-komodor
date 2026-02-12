---
page_title: "komodor_user_role_binding Resource - terraform-provider-komodor"
subcategory: ""
description: |-
  Manages user-role bindings in Komodor
---

# komodor_user_role_binding (Resource)

The `komodor_user_role_binding` resource allows you to assign one or more roles to a Komodor user. This resource manages the relationship between users and roles through the Komodor RBAC API.

## Example Usage

### Basic Usage

```terraform
resource "komodor_user_role_binding" "example" {
  name    = "admin-user-binding"
  user_id = "user@example.com"
  roles = [
    "role-id-1",
    "role-id-2"
  ]
}
```

### With Expiration

```terraform
resource "komodor_user_role_binding" "temporary_access" {
  name       = "contractor-access"
  user_id    = "contractor@example.com"
  roles      = ["viewer-role-id"]
  expiration = "2024-12-31T23:59:59Z"
}
```

### Using with komodor_user and komodor_role Resources

```terraform
resource "komodor_user" "developer" {
  display_name = "John Developer"
  email        = "john.developer@example.com"
}

resource "komodor_role" "developer_role" {
  name = "developer"
}

resource "komodor_user_role_binding" "developer_binding" {
  name    = "john-developer-binding"
  user_id = komodor_user.developer.id
  roles   = [komodor_role.developer_role.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) A unique name for this user-role binding. This is used for Terraform state management and must be unique within your configuration.
* `user_id` - (Required, String, ForceNew) The ID or email address of the user to assign roles to. This can be obtained from the `komodor_user` resource or data source.
* `roles` - (Required, Set of String) A set of role IDs or names to assign to the user. These can be obtained from the `komodor_role` resource or data source.
* `expiration` - (Optional, String) An optional expiration date for the user-role assignments in ISO 8601 format (e.g., "2024-12-31T23:59:59Z"). When set, all role assignments will expire at this time.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Terraform resource ID (same as `name`).

## Import

User-role bindings can be imported using the binding name:

```shell
terraform import komodor_user_role_binding.example admin-user-binding
```

Note: When importing, you must also provide the `user_id` and `roles` in your configuration file.
