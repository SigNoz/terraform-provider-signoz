// Package signoz re-exports the legacy alert and dashboard resources so the
// framework provider can register them. internal/provider sits outside the
// signoz/ tree and so can't import signoz/internal/... directly; these thin
// wrappers bridge that gap.
package signoz

import (
	signozdatasource "github.com/SigNoz/terraform-provider-signoz/signoz/internal/provider/datasource"
	signozresource "github.com/SigNoz/terraform-provider-signoz/signoz/internal/provider/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewAlertResource() resource.Resource {
	return signozresource.NewAlertResource()
}

func NewDashboardResource() resource.Resource {
	return signozresource.NewDashboardResource()
}

func NewAlertDataSource() datasource.DataSource {
	return signozdatasource.NewAlertDataSource()
}

func NewDashboardDataSource() datasource.DataSource {
	return signozdatasource.NewDashboardDataSource()
}
