package services

import (
	"context"
	"fmt"
	"time"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
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
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Tags        types.List   `tfsdk:"tags"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	Limits      types.List   `tfsdk:"limits"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Configure adds the provider configured client to the data source.
func (d *ingestionKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
		Description: "Fetches an ingestion key from SigNoz using its ID, including the limits attached to it.",
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
				Description: "The ingestion key value. The gateway returns it on reads; null when the deployment masks it.",
			},
			attr.Workspace: schema.StringAttribute{
				Computed:    true,
				Description: "ID of the workspace (organization) the key belongs to.",
			},
			attr.KeyTags: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the ingestion key.",
			},
			attr.ExpiresAt: schema.StringAttribute{
				Computed:    true,
				Description: "Expiration time of the ingestion key (RFC 3339); null for a key that never expires.",
			},
			attr.Limits: schema.ListAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: limitElementAttrTypes()},
				Description: "Rate limits attached to the key. Each element carries `id`, `signal`, `config` (`day`/`second`, each optional `size`/`count`), `metric` (the server's enforcement snapshot of the config), `tags`, `key_id`, `created_at` and `updated_at`.",
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the ingestion key (RFC 3339).",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the ingestion key (RFC 3339).",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *ingestionKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ingestionKeyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := d.client.GetIngestionKey(ctx, data.ID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz ingestion key: %s", err.Error()), "signoz_ingestion_key")
		return
	}

	data.ID = types.StringValue(key.ID)
	data.Name = types.StringValue(key.Name)
	data.ExpiresAt = model.ExpiresAtToTerraform(key.ExpiresAt)
	data.CreatedAt = types.StringValue(key.CreatedAt.Format(time.RFC3339Nano))
	data.UpdatedAt = types.StringValue(key.UpdatedAt.Format(time.RFC3339Nano))

	if key.Value != "" {
		data.Value = types.StringValue(key.Value)
	} else {
		data.Value = types.StringNull()
	}

	if key.WorkspaceID != "" {
		data.WorkspaceID = types.StringValue(key.WorkspaceID)
	} else {
		data.WorkspaceID = types.StringNull()
	}

	var diags diag.Diagnostics

	data.Tags, diags = key.TagsToTerraform()

	resp.Diagnostics.Append(diags...)

	data.Limits, diags = limitsToTerraform(key.Limits)

	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// limitsToTerraform renders the key's limits as the data source list value.
func limitsToTerraform(limits []model.Limit) (types.List, diag.Diagnostics) {
	if limits == nil {
		return types.ListNull(types.ObjectType{AttrTypes: limitElementAttrTypes()}), nil
	}

	elements := make([]tfattr.Value, 0, len(limits))

	var diags diag.Diagnostics

	for _, limit := range limits {
		var value tfattr.Value

		value, diags = limitToTerraformValue(limit)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: limitElementAttrTypes()}), diags
		}

		elements = append(elements, value)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: limitElementAttrTypes()}, elements)

	diags.Append(d...)

	return list, diags
}
