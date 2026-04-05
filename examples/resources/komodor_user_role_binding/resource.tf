resource "komodor_user_role_binding" "example" {
  name    = "admin-user-binding"
  user_id = "user@example.com" # Can also use user ID
  roles = [
    "role-id-1",
    "role-id-2"
  ]

}
