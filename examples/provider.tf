terraform {
  required_providers {
    komodor = {
      version = ">= 2.0.0"
      source  = "komodorio/komodor"
    }
  }
}

provider "komodor" {
  api_key = var.api_key
}