# Look up an existing alert by ID; all other attributes are computed.

data "signoz_alert" "existing" {
  id = signoz_alert.api_error_rate.id
}
