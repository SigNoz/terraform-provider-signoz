resource "signoz_user" "engineer" {
  email = "engineer@example.com"
}

resource "signoz_role" "serviceaccount_reader" {
  name        = "serviceaccount-reader"
  description = "Read-only access to service accounts"

  transaction_groups = [
    {
      relation = "read"

      object_group = {
        resource = {
          type = "serviceaccount"
          kind = "serviceaccount"
        }

        selectors = ["*"]
      }
    },
  ]
}

resource "signoz_user_role" "engineer_reader" {
  user_id = signoz_user.engineer.id
  role_id = signoz_role.serviceaccount_reader.id
}
