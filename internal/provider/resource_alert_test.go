package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceAlert_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertConfig_basic("tf-acc-test-alert"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_alert.test", "alert", "tf-acc-test-alert"),
					resource.TestCheckResourceAttr("signoz_alert.test", "alert_type", "METRIC_BASED_ALERT"),
					resource.TestCheckResourceAttr("signoz_alert.test", "severity", "info"),
					resource.TestCheckResourceAttrSet("signoz_alert.test", "id"),
					resource.TestCheckResourceAttrSet("signoz_alert.test", "state"),
				),
			},
			{
				ResourceName:      "signoz_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAlert_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertConfig_basic("tf-acc-test-alert-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_alert.test", "alert", "tf-acc-test-alert-update"),
					resource.TestCheckResourceAttr("signoz_alert.test", "severity", "info"),
				),
			},
			{
				Config: testAccAlertConfig_updated("tf-acc-test-alert-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("signoz_alert.test", "alert", "tf-acc-test-alert-updated"),
					resource.TestCheckResourceAttr("signoz_alert.test", "severity", "warning"),
					resource.TestCheckResourceAttr("signoz_alert.test", "eval_window", "10m0s"),
				),
			},
		},
	})
}

func TestAccDataSourceAlert_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertDataSourceConfig("tf-acc-test-alert-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.signoz_alert.test", "alert", "tf-acc-test-alert-ds"),
					resource.TestCheckResourceAttr("data.signoz_alert.test", "alert_type", "METRIC_BASED_ALERT"),
				),
			},
		},
	})
}

func testAccAlertConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "signoz_alert" "test" {
  alert      = %[1]q
  alert_type = "METRIC_BASED_ALERT"
  severity   = "info"
  rule_type  = "threshold_rule"
  condition  = jsonencode({
    compositeQuery = {
      builderQueries = {
        A = {
          queryName    = "A"
          dataSource   = "metrics"
          aggregateAttribute = {
            key      = "signoz_calls_total"
            dataType = "float64"
            type     = "Sum"
            isColumn = true
          }
          aggregateOperator = "rate"
          expression        = "A"
        }
      }
      queryType  = "builder"
      panelType  = "graph"
    }
    op        = "1"
    target    = 100
    matchType = "4"
  })
}
`, name)
}

func testAccAlertConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "signoz_alert" "test" {
  alert       = %[1]q
  alert_type  = "METRIC_BASED_ALERT"
  severity    = "warning"
  rule_type   = "threshold_rule"
  eval_window = "10m0s"
  condition   = jsonencode({
    compositeQuery = {
      builderQueries = {
        A = {
          queryName    = "A"
          dataSource   = "metrics"
          aggregateAttribute = {
            key      = "signoz_calls_total"
            dataType = "float64"
            type     = "Sum"
            isColumn = true
          }
          aggregateOperator = "rate"
          expression        = "A"
        }
      }
      queryType  = "builder"
      panelType  = "graph"
    }
    op        = "1"
    target    = 200
    matchType = "4"
  })
}
`, name)
}

func testAccAlertDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "signoz_alert" "test" {
  alert      = %[1]q
  alert_type = "METRIC_BASED_ALERT"
  severity   = "info"
  rule_type  = "threshold_rule"
  condition  = jsonencode({
    compositeQuery = {
      builderQueries = {
        A = {
          queryName    = "A"
          dataSource   = "metrics"
          aggregateAttribute = {
            key      = "signoz_calls_total"
            dataType = "float64"
            type     = "Sum"
            isColumn = true
          }
          aggregateOperator = "rate"
          expression        = "A"
        }
      }
      queryType  = "builder"
      panelType  = "graph"
    }
    op        = "1"
    target    = 100
    matchType = "4"
  })
}

data "signoz_alert" "test" {
  id = signoz_alert.test.id
}
`, name)
}
