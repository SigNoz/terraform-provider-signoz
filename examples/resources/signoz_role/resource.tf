resource "signoz_role" "serviceaccount_reader" {
  name        = "serviceaccount-reader"
  description = "Read-only access to service accounts"

  transaction_groups = [
    {
      relation = "read"

      object_group = {
        resource = {
          type = "metaresource"
          kind = "serviceaccount"
        }

        selectors = ["*"]
      }
    },
  ]
}
