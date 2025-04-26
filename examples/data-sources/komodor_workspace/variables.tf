variable "komodor_api_key" {
  description = "The API key for the Komodor API"
  type        = string
  sensitive   = true
}

variable "workspace_id" {
  description = "The ID of the workspace to retrieve"
  type        = string
} 