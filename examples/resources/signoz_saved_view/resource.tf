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

resource "signoz_saved_view" "error_logs" {
  name        = "Error logs last 24h"
  source_page = "logs"
  tags        = ["production", "errors"]
  category    = ""

  explorer_data = jsonencode(
    {
      query = {
        queryType = "builder"
        builder = {
          queryData = [
            {
              dataSource = "logs"
              queryName  = "A"
              expression = "A"
              filters = {
                items = [
                  {
                    key = {
                      key      = "severity_text"
                      dataType = "string"
                      type     = "tag"
                    }
                    op    = "="
                    value = "ERROR"
                  }
                ]
                op = "AND"
              }
            }
          ]
        }
      }
    }
  )
}

output "saved_view" {
  value = signoz_saved_view.error_logs
}
