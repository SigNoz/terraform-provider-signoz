package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboard_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic("tf-acc-test-dashboard"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_dashboard.test", "title", "tf-acc-test-dashboard"),
					resource.TestCheckResourceAttr("signoz_dashboard.test", "name", "tf-acc-test-dashboard"),
					resource.TestCheckResourceAttrSet("signoz_dashboard.test", "id"),
				),
			},
			{
				ResourceName:      "signoz_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDashboard_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic("tf-acc-test-dashboard-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_dashboard.test", "title", "tf-acc-test-dashboard-update"),
				),
			},
			{
				Config: testAccDashboardConfig_updated("tf-acc-test-dashboard-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_dashboard.test", "title", "tf-acc-test-dashboard-updated"),
					resource.TestCheckResourceAttr("signoz_dashboard.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccDataSourceDashboard_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardDataSourceConfig("tf-acc-test-dashboard-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_dashboard.test", "title", "tf-acc-test-dashboard-ds"),
				),
			},
		},
	})
}

func testAccDashboardConfig_basic(title string) string {
	return fmt.Sprintf(`
resource "signoz_dashboard" "test" {
  title                    = %[1]q
  name                     = %[1]q
  description              = "Acceptance test dashboard"
  version                  = "v4"
  collapsable_rows_migrated = false
  uploaded_grafana          = false
  layout                   = jsonencode([])
  widgets                  = jsonencode([])
  variables                = jsonencode({})
}
`, title)
}

func testAccDashboardConfig_updated(title string) string {
	return fmt.Sprintf(`
resource "signoz_dashboard" "test" {
  title                    = %[1]q
  name                     = %[1]q
  description              = "Updated description"
  version                  = "v4"
  collapsable_rows_migrated = false
  uploaded_grafana          = false
  layout                   = jsonencode([])
  widgets                  = jsonencode([])
  variables                = jsonencode({})
}
`, title)
}

func testAccDashboardDataSourceConfig(title string) string {
	return fmt.Sprintf(`
resource "signoz_dashboard" "test" {
  title                    = %[1]q
  name                     = %[1]q
  description              = "Acceptance test dashboard"
  version                  = "v4"
  collapsable_rows_migrated = false
  uploaded_grafana          = false
  layout                   = jsonencode([])
  widgets                  = jsonencode([])
  variables                = jsonencode({})
}

data "signoz_dashboard" "test" {
  id = signoz_dashboard.test.id
}
`, title)
}
