package services

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &legacyRoutePolicyDataSource{}
)

// NewLegacyRoutePolicyDataSource is a helper function to simplify the provider implementation.
func NewLegacyRoutePolicyDataSource() datasource.DataSource {
	return &legacyRoutePolicyDataSource{}
}

// legacyRoutePolicyDataSource is the data source implementation.
type legacyRoutePolicyDataSource struct {
	client *legacyclient.Client
}

// routePolicyModel maps route policy schema data.
type routePolicyModel struct {
	ID               types.String `tfsdk:"id"`
	RouteName        types.String `tfsdk:"route_name"`
	Description      types.String `tfsdk:"description"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Matchers         types.List   `tfsdk:"matchers"`
	ChannelIDs       types.List   `tfsdk:"channel_ids"`
	ContinueMatching types.Bool   `tfsdk:"continue_matching"`
	GroupWait        types.String `tfsdk:"group_wait"`
	GroupInterval    types.String `tfsdk:"group_interval"`
	RepeatInterval   types.String `tfsdk:"repeat_interval"`
	CreatedAt        types.String `tfsdk:"created_at"`
	CreatedBy        types.String `tfsdk:"created_by"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	UpdatedBy        types.String `tfsdk:"updated_by"`
}

// Configure adds the provider configured client to the data source.
func (d *legacyRoutePolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			"signoz_route_policy",
		)

		return
	}

	legacyClient, err := legacyclient.NewClient(api)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build legacy SigNoz client", err.Error())
		return
	}

	d.client = legacyClient
}

// Metadata returns the data source type name.
func (d *legacyRoutePolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_route_policy"
}

// Schema defines the schema for the data source.
func (d *legacyRoutePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a route policy from SigNoz using its ID.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the route policy.",
			},
			attr.RouteName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the route policy.",
			},
			attr.Description: schema.StringAttribute{
				Computed:    true,
				Description: "Description of the route policy.",
			},
			attr.Enabled: schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the route policy is enabled.",
			},
			attr.Matchers: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matchers for the route policy.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.MatchName: schema.StringAttribute{
							Computed:    true,
							Description: "Name of the matcher field.",
						},
						attr.MatchValue: schema.StringAttribute{
							Computed:    true,
							Description: "Value to match against.",
						},
						attr.MatchType: schema.StringAttribute{
							Computed:    true,
							Description: "Type of match to perform.",
						},
					},
				},
			},
			attr.ChannelIDs: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of notification channel IDs.",
			},
			attr.ContinueMatching: schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to continue matching subsequent route policies.",
			},
			attr.GroupWait: schema.StringAttribute{
				Computed:    true,
				Description: "How long to wait before sending a notification about new alerts.",
			},
			attr.GroupInterval: schema.StringAttribute{
				Computed:    true,
				Description: "How long to wait before sending a notification about new alerts in a group.",
			},
			attr.RepeatInterval: schema.StringAttribute{
				Computed:    true,
				Description: "How long to wait before re-sending a notification.",
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the route policy.",
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the route policy.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the route policy.",
			},
			attr.UpdatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the route policy.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *legacyRoutePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data routePolicyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	routePolicy, err := d.client.GetRoutePolicy(ctx, data.ID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz route policy: %s", err.Error()), "signoz_route_policy")
		return
	}

	// Set state values from retrieved data.
	data.ID = types.StringValue(routePolicy.ID)
	data.RouteName = types.StringValue(routePolicy.Name)
	data.Description = types.StringValue(routePolicy.Description)
	data.Enabled = types.BoolValue(routePolicy.Enabled)
	data.ContinueMatching = types.BoolValue(routePolicy.Continue)
	data.GroupWait = types.StringValue(routePolicy.GroupWait)
	data.GroupInterval = types.StringValue(routePolicy.GroupInterval)
	data.RepeatInterval = types.StringValue(routePolicy.RepeatInterval)
	data.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	data.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	data.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	data.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)

	var diagMatchers, diagChannelIDs diag.Diagnostics
	data.Matchers, diagMatchers = routePolicy.MatchersToTerraform()
	resp.Diagnostics.Append(diagMatchers...)

	data.ChannelIDs, diagChannelIDs = routePolicy.ChannelIDsToTerraform()
	resp.Diagnostics.Append(diagChannelIDs...)

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
