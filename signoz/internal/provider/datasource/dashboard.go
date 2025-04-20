package datasource

import (
	"context"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &dashboardDataSource{}
)

// NewDashboardDataSource is a helper function to simplify the provider implementation.
func NewDashboardDataSource() datasource.DataSource {
	return &dashboardDataSource{}
}

// dashboardDataSource is the data source implementation.
type dashboardDataSource struct {
	client *client.Client
}

// dashboardModel maps dashboard schema data.
type dashboardModel struct {
}

// Metadata returns the data source type name.
func (d *dashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = SigNozDashboard
}
