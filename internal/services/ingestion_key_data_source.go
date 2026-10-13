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
	_ datasource.DataSource = &ingestionKeyDataSource{}
)

// NewIngestionKeyDataSource is a helper function to simplify the provider implementation.
func NewIngestionKeyDataSource() datasource.DataSource {
	return &ingestionKeyDataSource{}
}

// ingestionKeyDataSource is the data source implementation.
type ingestionKeyDataSource struct {
	client *legacyclient.Client
}

// ingestionKeyModel maps ingestion key schema data.
type ingestionKeyModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Value        types.String `tfsdk:"value"`
	IngestionURL types.String `tfsdk:"ingestion_url"`
	Tags         types.List   `tfsdk:"tags"`
	ExpiresAt    types.String `tfsdk:"expires_at"`
	RateLimits   types.Object `tfsdk:"rate_limits"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// Configure adds the provider configured client to the data source.
func (d *ingestionKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			"signoz_ingestion_key",
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
func (d *ingestionKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_ingestion_key"
}

// Schema defines the schema for the data source.
func (d *ingestionKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an ingestion key from SigNoz using its ID.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the ingestion key.",
			},
			attr.KeyName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the ingestion key.",
			},
			attr.KeyValue: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The ingestion key value. May be masked or empty when read via data source.",
			},
			attr.IngestionURL: schema.StringAttribute{
				Computed:    true,
				Description: "The ingestion URL associated with this key.",
			},
			attr.KeyTags: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the ingestion key.",
			},
			attr.ExpiresAt: schema.StringAttribute{
				Computed:    true,
				Description: "Expiration time of the ingestion key.",
			},
			attr.RateLimits: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Rate limits for the ingestion key.",
				Attributes: map[string]schema.Attribute{
					attr.DayLimit: schema.Int64Attribute{
						Computed:    true,
						Description: "Maximum number of requests per day.",
					},
					attr.SecondLimit: schema.Int64Attribute{
						Computed:    true,
						Description: "Maximum number of requests per second.",
					},
				},
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the ingestion key.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the ingestion key.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *ingestionKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ingestionKeyModel
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key, err := d.client.GetIngestionKey(ctx, data.ID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz ingestion key: %s", err.Error()), "signoz_ingestion_key")
		return
	}

	// Set state values from retrieved data.
	data.ID = types.StringValue(key.ID)
	data.Name = types.StringValue(key.Name)
	data.Value = types.StringValue(key.Value)
	data.IngestionURL = types.StringValue(key.IngestionURL)
	data.ExpiresAt = types.StringValue(key.ExpiresAt)
	data.CreatedAt = types.StringValue(key.CreatedAt)
	data.UpdatedAt = types.StringValue(key.UpdatedAt)

	data.Tags, diags = key.TagsToTerraform()
	resp.Diagnostics.Append(diags...)

	data.RateLimits, diags = key.RateLimitsToTerraform()
	resp.Diagnostics.Append(diags...)

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
