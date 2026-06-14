provider "signoz" {
  # Root URL of your SigNoz instance: a SigNoz Cloud region URL, or the UI URL of a
  # self-hosted deployment. Can also be set with the SIGNOZ_ENDPOINT environment variable.
  endpoint = "http://localhost:8080"

  # API access token with the Admin role. Prefer the SIGNOZ_ACCESS_TOKEN environment
  # variable to keep the secret out of configuration and state.
  access_token = var.signoz_access_token
}
