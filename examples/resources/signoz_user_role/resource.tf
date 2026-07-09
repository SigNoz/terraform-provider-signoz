resource "signoz_user_role" "engineer_reader" {
  user_id = signoz_user.engineer.id
  role_id = signoz_role.serviceaccount_reader.id
}
