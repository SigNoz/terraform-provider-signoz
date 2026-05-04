data "signoz_notification_channel" "by_id" {
  id = signoz_notification_channel.slack.id
}

output "notification_channel_type" {
  value = data.signoz_notification_channel.by_id.type
}
