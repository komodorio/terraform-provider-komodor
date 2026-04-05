resource "komodor_user_role_binding" "temporary_access" {
  name       = "contractor-access"
  user_id    = "contractor@example.com"
  roles      = ["viewer-role-id"]
  expiration = "2024-12-31T23:59:59Z"
}
