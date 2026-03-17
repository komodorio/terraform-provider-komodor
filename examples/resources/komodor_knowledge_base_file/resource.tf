resource "komodor_knowledge_base_file" "runbook" {
  filename = "deployment-runbook.md"
  content  = file("${path.module}/runbooks/deployment-runbook.md")
}

resource "komodor_knowledge_base_file" "cluster_runbook" {
  filename = "production-cluster-runbook.md"
  content  = file("${path.module}/runbooks/production-cluster-runbook.md")

  clusters {
    include = ["production-us-east-1", "production-eu-west-1"]
    exclude = ["staging"]
  }
}
