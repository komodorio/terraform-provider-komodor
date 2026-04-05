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
