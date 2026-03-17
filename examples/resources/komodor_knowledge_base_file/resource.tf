resource "komodor_knowledge_base_file" "runbook" {
  filename = "deployment-runbook.md"
  content  = <<-EOT
    # Deployment Runbook
    This runbook describes how to handle common deployment issues.
  EOT
}

resource "komodor_knowledge_base_file" "cluster_runbook" {
  filename = "production-cluster-runbook.md"
  content  = <<-EOT
    # Production Cluster Runbook
    This runbook is scoped to production clusters only.
  EOT

  clusters {
    include = ["production-us-east-1", "production-eu-west-1"]
    exclude = ["staging"]
  }
}
