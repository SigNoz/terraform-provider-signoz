data "signoz_rule" "by_id" {
  id = signoz_rule.high_error_rate.id
}

output "rule_state" {
  value = data.signoz_rule.by_id.state
}
