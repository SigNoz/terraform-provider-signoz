resource "signoz_auth_domain" "example" {
  name = "example.com"
  config = {
    google = {
      kind = "google"
      spec = {
        client_id     = "912345678901-abcdefghijklmnop.apps.googleusercontent.com"
        client_secret = "GOCSPX-example-secret"
      }
    }
  }
}
