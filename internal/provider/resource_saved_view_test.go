package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceSavedView_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSavedViewConfig_basic("tf-acc-test-view"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_saved_view.test", "name", "tf-acc-test-view"),
					resource.TestCheckResourceAttr("signoz_saved_view.test", "source_page", "logs"),
					resource.TestCheckResourceAttrSet("signoz_saved_view.test", "uuid"),
				),
			},
			{
				ResourceName:      "signoz_saved_view.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSavedView_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSavedViewConfig_basic("tf-acc-test-view-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_saved_view.test", "name", "tf-acc-test-view-update"),
				),
			},
			{
				Config: testAccSavedViewConfig_updated("tf-acc-test-view-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_saved_view.test", "name", "tf-acc-test-view-updated"),
					resource.TestCheckResourceAttr("signoz_saved_view.test", "source_page", "traces"),
				),
			},
		},
	})
}

func TestAccDataSourceSavedView_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSavedViewDataSourceConfig("tf-acc-test-view-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_saved_view.test", "name", "tf-acc-test-view-ds"),
					resource.TestCheckResourceAttr("data.signoz_saved_view.test", "source_page", "logs"),
				),
			},
		},
	})
}

func testAccSavedViewConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "signoz_saved_view" "test" {
  name          = %[1]q
  source_page   = "logs"
  explorer_data = jsonencode({
    queryData = [
      {
        dataSource = "logs"
        queryName  = "A"
      }
    ]
  })
}
`, name)
}

func testAccSavedViewConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "signoz_saved_view" "test" {
  name          = %[1]q
  source_page   = "traces"
  explorer_data = jsonencode({
    queryData = [
      {
        dataSource = "traces"
        queryName  = "A"
      }
    ]
  })
}
`, name)
}

func testAccSavedViewDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_saved_view" "test" {
  name          = %[1]q
  source_page   = "logs"
  explorer_data = jsonencode({
    queryData = [
      {
        dataSource = "logs"
        queryName  = "A"
      }
    ]
  })
}

data "signoz_saved_view" "test" {
  uuid = signoz_saved_view.test.uuid
}
`, name)
}
