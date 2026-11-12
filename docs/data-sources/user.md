---
page_title: "jumpserver_user Data Source - terraform-provider-jumpserver"
subcategory: "Users & Groups"
description: |-
  Look up a user in JumpServer by username.
---
> Read-only edition: writes are rejected. Queries use the explicitly configured organization and require a unique exact match.


# `jumpserver_user` Data Source

Use this data source to look up an existing user in JumpServer by their username.

## Example Usage

```hcl
data "jumpserver_user" "admin" {
  username = "admin"
}

output "admin_id" {
  value = data.jumpserver_user.admin.id
}
```

## Argument Reference

* `username` - (Required) The username of the user to look up.

## Attribute Reference

* `id` - The UUID of the user.
* `name` - The display name of the user.
* `email` - The email address of the user.
* `is_active` - Whether the user is active.
