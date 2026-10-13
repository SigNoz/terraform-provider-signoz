package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceNotificationChannel_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_slack("tf-acc-test-slack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_notification_channel.test", "name", "tf-acc-test-slack"),
					resource.TestCheckResourceAttr("signoz_notification_channel.test", "type", "slack"),
					resource.TestCheckResourceAttrSet("signoz_notification_channel.test", "id"),
				),
			},
			{
				ResourceName:      "signoz_notification_channel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNotificationChannel_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_slack("tf-acc-test-slack-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_notification_channel.test", "name", "tf-acc-test-slack-update"),
				),
			},
			{
				Config: testAccNotificationChannelConfig_slackUpdated("tf-acc-test-slack-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_notification_channel.test", "name", "tf-acc-test-slack-renamed"),
					resource.TestCheckResourceAttr("signoz_notification_channel.test", "slack_configs.channel", "#alerts"),
				),
			},
		},
	})
}

func TestAccDataSourceNotificationChannel_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelDataSourceConfig("tf-acc-test-slack-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_notification_channel.test", "name", "tf-acc-test-slack-ds"),
					resource.TestCheckResourceAttr("data.signoz_notification_channel.test", "type", "slack"),
				),
			},
		},
	})
}

func testAccNotificationChannelConfig_slack(name string) string {
	return fmt.Sprintf(`
resource "signoz_notification_channel" "test" {
  name = %[1]q
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.slack.com/services/test/test/test"
    channel = "#test"
  }
}
`, name)
}

func testAccNotificationChannelConfig_slackUpdated(name string) string {
	return fmt.Sprintf(`
resource "signoz_notification_channel" "test" {
  name = %[1]q
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.slack.com/services/test/test/test"
    channel = "#alerts"
  }
}
`, name)
}

func testAccNotificationChannelDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_notification_channel" "test" {
  name = %[1]q
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.slack.com/services/test/test/test"
    channel = "#test"
  }
}

data "signoz_notification_channel" "test" {
  id = signoz_notification_channel.test.id
}
`, name)
}
