terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:3301"
  access_token = "<SIGNOZ-API-KEY>"
}

resource "signoz_notification_channel" "slack_alerts" {
  name = "slack-alerts"
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.example.com/services/TXXXXXXXXX/BXXXXXXXXX/xxxxxxxxxxxxxxxxxxxxxxxx"
    channel = "#alerts"
    title   = "SigNoz Alert"
    text    = "Alert fired: {{ .CommonLabels.alertname }}"
  }
}

output "slack_channel" {
  value = signoz_notification_channel.slack_alerts
}
