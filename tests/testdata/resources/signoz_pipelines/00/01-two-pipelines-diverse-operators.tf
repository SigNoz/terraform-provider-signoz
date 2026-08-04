# Scenario 00 — the full pipeline list as one resource: two pipelines covering
# the parser operators (grok, add, remove, trace, severity with mapping).
# Base-only: create, no-drift, destroy.
resource "signoz_pipelines" "scenario_00" {
  pipelines = [
    {
      name        = "testdata-scenario-00-a"
      alias       = "scenario-00-a"
      description = "Parse nginx access logs."

      filter = {
        items = [
          {
            key = {
              key       = "source"
              data_type = "string"
              type      = "tag"
            }
            op    = "="
            value = jsonencode("nginx")
          }
        ]
      }

      config = [
        {
          type       = "grok_parser"
          name       = "Parse access log line"
          pattern    = "%%{TIMESTAMP_ISO8601:attributes.ts} %%{LOGLEVEL:attributes.level} %%{GREEDYDATA:attributes.msg}"
          parse_from = "body"
          parse_to   = "attributes"
        },
        {
          type  = "add"
          name  = "Tag the environment"
          field = "attributes.env"
          value = "testdata"
        },
        {
          type  = "remove"
          field = "attributes.password"
        }
      ]
    },
    {
      name  = "testdata-scenario-00-b"
      alias = "scenario-00-b"

      filter = {
        op = "OR"
        items = [
          {
            key = {
              key       = "service.name"
              data_type = "string"
              type      = "resource"
            }
            op    = "="
            value = jsonencode("checkout")
          },
          {
            key = {
              key       = "service.name"
              data_type = "string"
              type      = "resource"
            }
            op    = "="
            value = jsonencode("payments")
          }
        ]
      }

      config = [
        {
          type = "trace_parser"
          trace_id = {
            parse_from = "attributes.trace_id"
          }
        },
        {
          type       = "severity_parser"
          parse_from = "attributes.level"
          mapping = {
            error = ["err", "ERROR"]
            info  = ["info", "INFO"]
          }
          overwrite_text = true
        }
      ]
    }
  ]
}
