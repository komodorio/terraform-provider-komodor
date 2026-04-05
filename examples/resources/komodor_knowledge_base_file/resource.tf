resource "komodor_knowledge_base_file" "runbook" {
  file_type = "knowledge-base"
  filename  = "deployment-runbook.md"
  content   = <<-EOT
    # Deployment Runbook
    This runbook describes how to handle common deployment issues.
  EOT
}

resource "komodor_knowledge_base_file" "blueprint" {
  file_type = "blueprint"
  filename  = "cluster-blueprint.md"
  content   = <<-EOT
    # Cluster Blueprint
    This blueprint provides cluster configuration guidelines.
  EOT
}

resource "komodor_knowledge_base_file" "cluster_runbook" {
  file_type = "knowledge-base"
  filename  = "production-cluster-runbook.md"
  content   = <<-EOT
    # Production Cluster Runbook
    This runbook is scoped to production clusters only.
  EOT

  clusters {
    include = ["production-us-east-1", "production-eu-west-1"]
    exclude = ["staging"]
  }
}

# Load file content from disk using the file() function.
# The referenced file must exist relative to the module root.
resource "komodor_knowledge_base_file" "from_file" {
  file_type = "knowledge-base"
  filename  = "runbook.md"
  content   = file("${path.module}/runbook.md")
}
