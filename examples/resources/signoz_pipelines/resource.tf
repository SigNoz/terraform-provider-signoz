# Manages the full ordered list of log pipelines. The SigNoz API saves all
# pipelines as one versioned list, so this resource is a singleton: declare
# every pipeline here, in processing order.
resource "signoz_pipelines" "all" {
  pipelines = [
    {
      name        = "parse-nginx-logs"
      alias       = "parse-nginx-logs"
      description = "Parse JSON access logs from the nginx service."

      filter = {
        items = [
          {
            key = {
              key       = "service.name"
              data_type = "string"
              type      = "resource"
            }
            op    = "="
            value = jsonencode("nginx")
          }
        ]
      }

      config = [
        {
          type       = "json_parser"
          name       = "Parse body as JSON"
          parse_from = "body"
          parse_to   = "attributes"
        },
        {
          type  = "remove"
          name  = "Drop the raw password field"
          field = "attributes.password"
        }
      ]
    }
  ]
}
