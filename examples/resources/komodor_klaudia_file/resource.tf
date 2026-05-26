resource "komodor_klaudia_file" "knowledge_base" {
  type        = "knowledge-base"
  filename    = "platform-runbook.md"
  source_path = "${path.module}/platform-runbook.md"

  clusters {
    include = ["*"]
  }
}

resource "komodor_klaudia_file" "blueprint" {
  type        = "blueprint"
  filename    = "service-blueprint.yaml"
  source_path = "${path.module}/service-blueprint.yaml"
}
