resource "signoz_notification_channel" "webhook_pager" {
  name         = "pager-webhook"
  display_name = "Pager Webhook"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://alerts.example.com/signoz"
        send_resolved = true
      }
    }
  }
}
