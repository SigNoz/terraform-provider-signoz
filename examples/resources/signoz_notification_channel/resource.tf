terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:8080"
  access_token = "<SIGNOZ-API-KEY>"
}

# Phase 2 supports `slack_configs` only. All other *_configs attributes
# exist in the schema but are rejected by the server-side convertor.
resource "signoz_notification_channel" "slack" {
  name = "platform-alerts-slack"

  slack_configs = [{
    channel       = "#alerts"
    api_url       = "<YOUR_SLACK_WEBHOOK_URL>"
    title         = "{{ .CommonAnnotations.summary }}"
    text          = "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"
    send_resolved = true
    icon_emoji    = ":warning:"
  }]
}

output "notification_channel_id" {
  value = signoz_notification_channel.slack.id
}
