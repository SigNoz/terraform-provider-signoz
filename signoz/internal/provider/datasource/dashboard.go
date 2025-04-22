package datasource

import (
	"context"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Dashboard types.String `tfsdk:"dashboard"`
	Condition types.String `tfsdk:"condition"`
	Labels    types.Map    `tfsdk:"labels"`
	UUID      types.String `tfsdk:"uuid"`
}

// Metadata returns the data source type name.
func (d *dashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = SigNozDashboard
}

// Configure adds the provider configured client to the data source.
func (d *dashboardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			SigNozDashboard,
		)

		return
	}

	d.client = client
}

// Schema defines the schema for the data source.
func (d *dashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// todo: verify the correctness of the description below.
		Description: "Fetches a dashboard from Signoz using its UUID. The UUID can be found in the URL of the alert in the Signoz UI.",
		Attributes: map[string]schema.Attribute{
			attr.Dashboard: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the dashboard.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dashboardModel
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dashboard, err := d.client.GetDashboard(ctx, data.UUID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz alert: %s", err.Error()), SigNozDashboard)
		return
	}

	// Set state values from retrieved data
	data.UUID = types.StringValue(dashboard.UUID)

	resp.Diagnostics.Append(diags...)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
