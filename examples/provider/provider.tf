# See https://github.com/komodorio/terraform-provider-komodor#how-to-use
# for how to get an api key

provider "komodor" {
  api_key = "KOMODOR_API_KEY"
  # Optional: For EU region, uncomment the line below:
  # api_url = "https://api.eu.komodor.com"
}
