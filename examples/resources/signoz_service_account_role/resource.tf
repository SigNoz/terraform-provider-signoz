resource "signoz_service_account" "deployer" {
  name = "deploy-bot"
}

resource "signoz_role" "reader" {
  name        = "service-account-role-reader"
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

resource "signoz_service_account_role" "deployer_reader" {
  service_account_id = signoz_service_account.deployer.id
  role_id            = signoz_role.reader.id
}
