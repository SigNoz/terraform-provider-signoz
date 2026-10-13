# signoz_plugin_dashboard (Resource)

Creates and manages a **SigNoz V2 dashboard** composed of **WebComponent plugin panels**
(`signoz/WebComponentPanel`). This is the Lacrosse / noizu-forks extension for
configuring plugin host pages as code.

## Example

```terraform
resource "signoz_plugin_dashboard" "ops" {
  title       = "Lacrosse Plugins"
  description = "Plugin host pages"
  tags        = ["lacrosse", "plugins"]

  page {
    name = "Overview"

    plugin_panel {
      title     = "Hello WC"
      plugin_id = "hello"
      element   = "lax-hello-wc"
      x         = 0
      y         = 0
      w         = 6
      h         = 4
      attributes = {
        "data-env" = "stage"
      }
    }
  }
}
```

## Schema

### Required

- `title` (String)
- `page` (Block List) — at least one page, each with one or more `plugin_panel`

### Optional

- `description` (String)
- `tags` (List of String)

### Nested plugin_panel

- `title`, `plugin_id`, `element` (Required)
- `id` (Optional/Computed panel UUID)
- `attributes` (Map of String)
- `properties` (String JSON object)
- `config_keys` (List of String)
- `x`, `y`, `w`, `h` (layout; defaults 0,0,6,4)

### Read-only

- `id` (String) dashboard UUID
- `document` (String) rendered V2 JSON
