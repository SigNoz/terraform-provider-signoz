package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceRoutePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutePolicyConfig_basic("tf-acc-test-route"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_route_policy.test", "route_name", "tf-acc-test-route"),
					resource.TestCheckResourceAttr("signoz_route_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("signoz_route_policy.test", "id"),
				),
			},
			{
				ResourceName:      "signoz_route_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRoutePolicy_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutePolicyConfig_basic("tf-acc-test-route-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_route_policy.test", "route_name", "tf-acc-test-route-update"),
					resource.TestCheckResourceAttr("signoz_route_policy.test", "description", ""),
				),
			},
			{
				Config: testAccRoutePolicyConfig_updated("tf-acc-test-route-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_route_policy.test", "route_name", "tf-acc-test-route-updated"),
					resource.TestCheckResourceAttr("signoz_route_policy.test", "description", "Updated route policy"),
					resource.TestCheckResourceAttr("signoz_route_policy.test", "group_wait", "60s"),
				),
			},
		},
	})
}

func TestAccDataSourceRoutePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutePolicyDataSourceConfig("tf-acc-test-route-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_route_policy.test", "route_name", "tf-acc-test-route-ds"),
				),
			},
		},
	})
}

func testAccRoutePolicyConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "signoz_notification_channel" "route_test" {
  name = "tf-acc-route-channel"
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.slack.com/services/test/test/test"
    channel = "#test"
  }
}

resource "signoz_route_policy" "test" {
  route_name  = %[1]q
  channel_ids = [signoz_notification_channel.route_test.id]

  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }
}
`, name)
}

func testAccRoutePolicyConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "signoz_notification_channel" "route_test" {
  name = "tf-acc-route-channel"
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.slack.com/services/test/test/test"
    channel = "#test"
  }
}

resource "signoz_route_policy" "test" {
  route_name  = %[1]q
  description = "Updated route policy"
  channel_ids = [signoz_notification_channel.route_test.id]
  group_wait  = "60s"

  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }

  matchers {
    match_name  = "service"
    match_value = "api"
    match_type  = "equals"
  }
}
`, name)
}

func testAccRoutePolicyDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_notification_channel" "route_test" {
  name = "tf-acc-route-channel-ds"
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.slack.com/services/test/test/test"
    channel = "#test"
  }
}

resource "signoz_route_policy" "test" {
  route_name  = %[1]q
  channel_ids = [signoz_notification_channel.route_test.id]

  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }
}

data "signoz_route_policy" "test" {
  id = signoz_route_policy.test.id
}
`, name)
}
