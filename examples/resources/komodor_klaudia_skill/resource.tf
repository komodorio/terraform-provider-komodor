resource "komodor_klaudia_skill" "example" {
  name         = "my-skill"
  description  = "Explains how to investigate database connection issues."
  instructions = <<-EOT
    You are an expert in diagnosing Kubernetes database connectivity issues.
    When investigating, always check:
    1. Network policies between the app and the database pods.
    2. Secret rotation history for the database credentials.
    3. Recent deployment changes in the same namespace.
  EOT

  clusters   = ["*"]
  is_enabled = true
}
