package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDowntimeSchedule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDowntimeScheduleConfig_basic("tf-acc-test-downtime"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_downtime_schedule.test", "name", "tf-acc-test-downtime"),
					resource.TestCheckResourceAttr("signoz_downtime_schedule.test", "all_alerts", "true"),
					resource.TestCheckResourceAttrSet("signoz_downtime_schedule.test", "id"),
				),
			},
			{
				ResourceName:      "signoz_downtime_schedule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDowntimeSchedule_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDowntimeScheduleConfig_basic("tf-acc-test-downtime-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_downtime_schedule.test", "name", "tf-acc-test-downtime-update"),
				),
			},
			{
				Config: testAccDowntimeScheduleConfig_updated("tf-acc-test-downtime-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_downtime_schedule.test", "name", "tf-acc-test-downtime-updated"),
					resource.TestCheckResourceAttr("signoz_downtime_schedule.test", "description", "Updated downtime schedule"),
				),
			},
		},
	})
}

func TestAccDataSourceDowntimeSchedule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDowntimeScheduleDataSourceConfig("tf-acc-test-downtime-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_downtime_schedule.test", "name", "tf-acc-test-downtime-ds"),
				),
			},
		},
	})
}

func testAccDowntimeScheduleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "signoz_downtime_schedule" "test" {
  name       = %[1]q
  all_alerts = true

  schedule = {
    start_time = "2030-01-01T00:00:00Z"
    end_time   = "2030-01-01T02:00:00Z"
    timezone   = "UTC"

    recurrence = {
      recurrence_type = "weekly"
      days_of_week    = [1, 3, 5]
      duration        = "2h"
    }
  }
}
`, name)
}

func testAccDowntimeScheduleConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "signoz_downtime_schedule" "test" {
  name        = %[1]q
  description = "Updated downtime schedule"
  all_alerts  = true

  schedule = {
    start_time = "2030-01-01T00:00:00Z"
    end_time   = "2030-01-01T04:00:00Z"
    timezone   = "UTC"

    recurrence = {
      recurrence_type = "daily"
      duration        = "4h"
    }
  }
}
`, name)
}

func testAccDowntimeScheduleDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_downtime_schedule" "test" {
  name       = %[1]q
  all_alerts = true

  schedule = {
    start_time = "2030-01-01T00:00:00Z"
    end_time   = "2030-01-01T02:00:00Z"
    timezone   = "UTC"

    recurrence = {
      recurrence_type = "weekly"
      days_of_week    = [1, 3, 5]
      duration        = "2h"
    }
  }
}

data "signoz_downtime_schedule" "test" {
  id = signoz_downtime_schedule.test.id
}
`, name)
}
