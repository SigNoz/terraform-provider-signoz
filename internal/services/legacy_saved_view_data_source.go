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
	_ datasource.DataSource = &legacySavedViewDataSource{}
)

// NewLegacySavedViewDataSource is a helper function to simplify the provider implementation.
func NewLegacySavedViewDataSource() datasource.DataSource {
	return &legacySavedViewDataSource{}
}

// legacySavedViewDataSource is the data source implementation.
type legacySavedViewDataSource struct {
	client *legacyclient.Client
}

// savedViewModel maps saved view schema data.
type savedViewModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Name         types.String `tfsdk:"name"`
	Category     types.String `tfsdk:"category"`
	SourcePage   types.String `tfsdk:"source_page"`
	ExplorerData types.String `tfsdk:"explorer_data"`
	ExtraData    types.String `tfsdk:"extra_data"`
	Tags         types.List   `tfsdk:"tags"`
	CreatedAt    types.String `tfsdk:"created_at"`
	CreatedBy    types.String `tfsdk:"created_by"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// Configure adds the provider configured client to the data source.
func (d *legacySavedViewDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			"signoz_saved_view",
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
func (d *legacySavedViewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_saved_view"
}

// Schema defines the schema for the data source.
func (d *legacySavedViewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a saved view from SigNoz using its UUID.",
		Attributes: map[string]schema.Attribute{
			attr.UUID: schema.StringAttribute{
				Required:    true,
				Description: "UUID of the saved view.",
			},
			attr.ViewName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the saved view.",
			},
			attr.Category: schema.StringAttribute{
				Computed:    true,
				Description: "Category of the saved view.",
			},
			attr.SourcePage: schema.StringAttribute{
				Computed:    true,
				Description: "Source page of the saved view. Possible values are: logs, traces, and metrics.",
			},
			attr.ExplorerData: schema.StringAttribute{
				Computed:    true,
				Description: "Explorer data of the saved view as a JSON string.",
			},
			attr.ExtraData: schema.StringAttribute{
				Computed:    true,
				Description: "Extra data associated with the saved view.",
			},
			attr.ViewTags: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the saved view.",
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the saved view.",
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the saved view.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the saved view.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *legacySavedViewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data savedViewModel
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	savedView, err := d.client.GetSavedView(ctx, data.UUID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz saved view: %s", err.Error()), "signoz_saved_view")
		return
	}

	// Set state values from retrieved data.
	data.UUID = types.StringValue(savedView.UUID)
	data.Name = types.StringValue(savedView.Name)
	data.Category = types.StringValue(savedView.Category)
	data.SourcePage = types.StringValue(savedView.SourcePage)
	data.ExtraData = types.StringValue(savedView.ExtraData)
	data.CreatedAt = types.StringValue(savedView.CreatedAt)
	data.CreatedBy = types.StringValue(savedView.CreatedBy)
	data.UpdatedAt = types.StringValue(savedView.UpdatedAt)

	data.ExplorerData, err = savedView.ExplorerDataToTerraform()
	if err != nil {
		addReadErr(&resp.Diagnostics, err, "signoz_saved_view")
		return
	}

	data.Tags, diags = savedView.TagsToTerraform()
	resp.Diagnostics.Append(diags...)

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
