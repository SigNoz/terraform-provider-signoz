package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceIngestionKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionKeyConfig_basic("tf-acc-test-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_ingestion_key.test", "name", "tf-acc-test-key"),
					resource.TestCheckResourceAttrSet("signoz_ingestion_key.test", "id"),
					resource.TestCheckResourceAttrSet("signoz_ingestion_key.test", "value"),
					resource.TestCheckResourceAttrSet("signoz_ingestion_key.test", "ingestion_url"),
				),
			},
			{
				ResourceName:            "signoz_ingestion_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func TestAccResourceIngestionKey_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionKeyConfig_basic("tf-acc-test-key-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_ingestion_key.test", "name", "tf-acc-test-key-update"),
				),
			},
			{
				Config: testAccIngestionKeyConfig_updated("tf-acc-test-key-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_ingestion_key.test", "name", "tf-acc-test-key-renamed"),
				),
			},
		},
	})
}

func TestAccDataSourceIngestionKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionKeyDataSourceConfig("tf-acc-test-key-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_ingestion_key.test", "name", "tf-acc-test-key-ds"),
					resource.TestCheckResourceAttrSet("data.signoz_ingestion_key.test", "id"),
				),
			},
		},
	})
}

func testAccIngestionKeyConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "signoz_ingestion_key" "test" {
  name = %[1]q
}
`, name)
}

func testAccIngestionKeyConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "signoz_ingestion_key" "test" {
  name = %[1]q
  tags = ["terraform", "acc-test"]
}
`, name)
}

func testAccIngestionKeyDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_ingestion_key" "test" {
  name = %[1]q
}

data "signoz_ingestion_key" "test" {
  id = signoz_ingestion_key.test.id
}
`, name)
}
