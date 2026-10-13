package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDataRetention_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRetentionConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_data_retention.test", "id", "settings"),
					resource.TestCheckResourceAttr("signoz_data_retention.test", "metrics_ttl_duration_hrs", "720"),
					resource.TestCheckResourceAttr("signoz_data_retention.test", "logs_ttl_duration_hrs", "720"),
					resource.TestCheckResourceAttr("signoz_data_retention.test", "traces_ttl_duration_hrs", "720"),
				),
			},
			{
				ResourceName:      "signoz_data_retention.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "settings",
			},
		},
	})
}

func TestAccResourceDataRetention_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRetentionConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_data_retention.test", "metrics_ttl_duration_hrs", "720"),
				),
			},
			{
				Config: testAccDataRetentionConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_data_retention.test", "metrics_ttl_duration_hrs", "168"),
					resource.TestCheckResourceAttr("signoz_data_retention.test", "logs_ttl_duration_hrs", "336"),
					resource.TestCheckResourceAttr("signoz_data_retention.test", "traces_ttl_duration_hrs", "168"),
				),
			},
		},
	})
}

func TestAccDataSourceDataRetention_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRetentionDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.signoz_data_retention.test", "metrics_ttl_duration_hrs"),
					resource.TestCheckResourceAttrSet("data.signoz_data_retention.test", "logs_ttl_duration_hrs"),
					resource.TestCheckResourceAttrSet("data.signoz_data_retention.test", "traces_ttl_duration_hrs"),
				),
			},
		},
	})
}

func testAccDataRetentionConfig_basic() string {
	return `
resource "signoz_data_retention" "test" {
  metrics_ttl_duration_hrs = 720
  logs_ttl_duration_hrs    = 720
  traces_ttl_duration_hrs  = 720
}
`
}

func testAccDataRetentionConfig_updated() string {
	return `
resource "signoz_data_retention" "test" {
  metrics_ttl_duration_hrs = 168
  logs_ttl_duration_hrs    = 336
  traces_ttl_duration_hrs  = 168
}
`
}

func testAccDataRetentionDataSourceConfig() string {
	return `
resource "signoz_data_retention" "test" {
  metrics_ttl_duration_hrs = 720
  logs_ttl_duration_hrs    = 720
  traces_ttl_duration_hrs  = 720
}

data "signoz_data_retention" "test" {}
`
}
