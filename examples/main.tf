resource "komodor_role" "my-role" {
  name = "my-role"
}

resource "komodor_policy_role_attachment" "my-attachement" {
  name     = "test-attachement"
  policies = []
  role     = komodor_role.my-role.id
}

data "komodor_policy_v2" "my-policy" {
  name = "default-read-only"
}
