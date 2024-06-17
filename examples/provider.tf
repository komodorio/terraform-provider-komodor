terraform {
  required_providers {
    komodor = {
      version = ">= 1.0.8"
      source  = "komodorio/komodor"
    }
  }
}

provider "komodor" {
  api_key = var.api_key
}