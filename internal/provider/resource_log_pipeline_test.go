package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceLogPipeline_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogPipelineConfig_basic("tf-acc-test-pipeline"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_log_pipeline.test", "name", "tf-acc-test-pipeline"),
					resource.TestCheckResourceAttr("signoz_log_pipeline.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("signoz_log_pipeline.test", "id"),
				),
			},
			{
				ResourceName:      "signoz_log_pipeline.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogPipeline_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogPipelineConfig_basic("tf-acc-test-pipeline-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_log_pipeline.test", "name", "tf-acc-test-pipeline-update"),
				),
			},
			{
				Config: testAccLogPipelineConfig_updated("tf-acc-test-pipeline-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_log_pipeline.test", "name", "tf-acc-test-pipeline-updated"),
					resource.TestCheckResourceAttr("signoz_log_pipeline.test", "description", "Updated log pipeline"),
					resource.TestCheckResourceAttr("signoz_log_pipeline.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceLogPipeline_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogPipelineDataSourceConfig("tf-acc-test-pipeline-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_log_pipeline.test", "name", "tf-acc-test-pipeline-ds"),
				),
			},
		},
	})
}

func testAccLogPipelineConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "signoz_log_pipeline" "test" {
  name     = %[1]q
  order_id = 1
  filter   = jsonencode({
    op    = "AND"
    items = [
      {
        key = {
          key      = "service"
          dataType = "string"
          type     = "tag"
        }
        op    = "="
        value = "test-service"
      }
    ]
  })
  config = jsonencode([
    {
      type    = "regex"
      enabled = true
      name    = "extract-level"
      id      = "extract-level"
      orderId = 1
      output  = "level"
      regex   = "level=(?P<level>\\w+)"
      parse_from = "body"
    }
  ])
}
`, name)
}

func testAccLogPipelineConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "signoz_log_pipeline" "test" {
  name        = %[1]q
  description = "Updated log pipeline"
  order_id    = 1
  enabled     = false
  filter      = jsonencode({
    op    = "AND"
    items = [
      {
        key = {
          key      = "service"
          dataType = "string"
          type     = "tag"
        }
        op    = "="
        value = "updated-service"
      }
    ]
  })
  config = jsonencode([
    {
      type    = "regex"
      enabled = true
      name    = "extract-level"
      id      = "extract-level"
      orderId = 1
      output  = "level"
      regex   = "level=(?P<level>\\w+)"
      parse_from = "body"
    }
  ])
}
`, name)
}

func testAccLogPipelineDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_log_pipeline" "test" {
  name     = %[1]q
  order_id = 1
  filter   = jsonencode({
    op    = "AND"
    items = [
      {
        key = {
          key      = "service"
          dataType = "string"
          type     = "tag"
        }
        op    = "="
        value = "test-service"
      }
    ]
  })
  config = jsonencode([
    {
      type    = "regex"
      enabled = true
      name    = "extract-level"
      id      = "extract-level"
      orderId = 1
      output  = "level"
      regex   = "level=(?P<level>\\w+)"
      parse_from = "body"
    }
  ])
}

data "signoz_log_pipeline" "test" {
  id = signoz_log_pipeline.test.id
}
`, name)
}
