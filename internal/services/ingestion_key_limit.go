package services

import (
	"context"
	"fmt"
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ingestionKeyLimitResource{}
	_ resource.ResourceWithConfigure   = &ingestionKeyLimitResource{}
	_ resource.ResourceWithImportState = &ingestionKeyLimitResource{}
)

// NewIngestionKeyLimitResource is a helper function to simplify the provider implementation.
func NewIngestionKeyLimitResource() resource.Resource {
	return &ingestionKeyLimitResource{}
}

// ingestionKeyLimitResource is the resource implementation.
type ingestionKeyLimitResource struct {
	client *legacyclient.Client
}

// ingestionKeyLimitResourceModel maps the resource schema data.
type ingestionKeyLimitResourceModel struct {
	ID        types.String `tfsdk:"id"`
	KeyID     types.String `tfsdk:"key_id"`
	Signal    types.String `tfsdk:"signal"`
	Config    types.Object `tfsdk:"config"`
	Tags      types.List   `tfsdk:"tags"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Configure adds the provider configured client to the resource.
func (r *ingestionKeyLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, "signoz_ingestion_key_limit",
		)

		return
	}

	legacyClient, err := legacyclient.NewClient(api)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build legacy SigNoz client", err.Error())
		return
	}

	r.client = legacyClient
}

// Metadata returns the resource type name.
func (r *ingestionKeyLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "signoz_ingestion_key_limit"
}

// Schema defines the schema for the resource.
func (r *ingestionKeyLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a single rate limit attached to an ingestion key. Keys and limits are separate server objects: deleting the key removes its limits, but limits are created, updated and deleted through their own endpoints.",
		MarkdownDescription: "Manages a single rate limit attached to an ingestion key. Keys and limits are separate server objects: deleting the key removes its limits, but limits are created, updated and deleted through their own endpoints.",
		Attributes: map[string]schema.Attribute{
			attr.KeyID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the ingestion key the limit applies to. Changing it replaces the limit (the server has no move operation).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			attr.Signal: schema.StringAttribute{
				Required:    true,
				Description: "Telemetry signal the limit applies to. One of `logs`, `traces`, `metrics`. Immutable once set; changing it replaces the limit.",
				Validators: []validator.String{
					stringvalidator.OneOf(model.SignalLogs, model.SignalTraces, model.SignalMetrics),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			attr.Config: schema.SingleNestedAttribute{
				Required:    true,
				Description: "Limit configuration. At least one of `day` or `second` must be set; within an arm any of `size` (bytes) or `count` (requests) may be left out to leave that dimension unbounded.",
				Attributes: map[string]schema.Attribute{
					attr.Day: schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Per-day ceiling. Omit for no daily limit.",
						Attributes: map[string]schema.Attribute{
							attr.Size: schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum bytes per day. Omit for unbounded.",
							},
							attr.Count: schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum requests per day. Omit for unbounded.",
							},
						},
					},
					attr.Second: schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Per-second ceiling. Omit for no rate limit.",
						Attributes: map[string]schema.Attribute{
							attr.Size: schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum bytes per second. Omit for unbounded.",
							},
							attr.Count: schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum requests per second. Omit for unbounded.",
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(
						path.MatchRelative().AtName(attr.Day),
						path.MatchRelative().AtName(attr.Second),
					),
				},
			},
			attr.KeyTags: schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the limit.",
			},
			// computed.
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated unique ID for the limit.",
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the limit (RFC 3339).",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the limit (RFC 3339).",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ingestionKeyLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ingestionKeyLimitResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postable, diags := limitPostableFromPlan(plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateIngestionKeyLimit(ctx, postable)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, "signoz_ingestion_key_limit")
		return
	}

	limit, err := r.client.GetIngestionKeyLimit(ctx, created.ID)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, "signoz_ingestion_key_limit")
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(refreshLimitModel(&plan, limit)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ingestionKeyLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ingestionKeyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ingestion key limit", map[string]any{"id": state.ID.ValueString()})

	limit, err := r.client.GetIngestionKeyLimit(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			tflog.Warn(ctx, "Ingestion key limit not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}

		addErr(&resp.Diagnostics, err, operationRead, "signoz_ingestion_key_limit")
		return
	}

	resp.Diagnostics.Append(refreshLimitModel(&state, limit)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ingestionKeyLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ingestionKeyLimitResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := limitConfigFromTerraform(plan.Config)

	resp.Diagnostics.Append(diags...)

	tags, diags := model.TagsFromTerraform(plan.Tags)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updatable := model.UpdatableIngestionKeyLimit{Config: config, Tags: tags}

	if err := r.client.UpdateIngestionKeyLimit(ctx, state.ID.ValueString(), updatable); err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_ingestion_key_limit")
		return
	}

	limit, err := r.client.GetIngestionKeyLimit(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, "signoz_ingestion_key_limit")
		return
	}

	plan.ID = state.ID
	plan.KeyID = state.KeyID
	plan.Signal = state.Signal

	resp.Diagnostics.Append(refreshLimitModel(&plan, limit)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ingestionKeyLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ingestionKeyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteIngestionKeyLimit(ctx, state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}

		addErr(&resp.Diagnostics, err, operationDelete, "signoz_ingestion_key_limit")
		return
	}
}

// ImportState imports Terraform state into the resource; the follow-up read
// recovers key_id from the server.
func (r *ingestionKeyLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)
}

// limitPostableFromPlan builds the create body from the plan.
func limitPostableFromPlan(plan ingestionKeyLimitResourceModel) (model.PostableIngestionKeyLimit, diag.Diagnostics) {
	config, diags := limitConfigFromTerraform(plan.Config)

	tags, d := model.TagsFromTerraform(plan.Tags)

	diags.Append(d...)

	if diags.HasError() {
		return model.PostableIngestionKeyLimit{}, diags
	}

	return model.PostableIngestionKeyLimit{
		KeyID:  plan.KeyID.ValueString(),
		Signal: plan.Signal.ValueString(),
		Config: config,
		Tags:   tags,
	}, diags
}

// refreshLimitModel overwrites the model's server-owned fields from the
// fetched limit.
func refreshLimitModel(m *ingestionKeyLimitResourceModel, limit *model.Limit) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Signal = types.StringValue(limit.Signal)
	m.CreatedAt = types.StringValue(limit.CreatedAt.Format(time.RFC3339Nano))
	m.UpdatedAt = types.StringValue(limit.UpdatedAt.Format(time.RFC3339Nano))

	if limit.KeyID != "" {
		m.KeyID = types.StringValue(limit.KeyID)
	}

	var d diag.Diagnostics

	m.Config, d = limitConfigToTerraform(limit.Config)

	diags.Append(d...)

	m.Tags = types.ListNull(types.StringType)

	if limit.Tags != nil {
		var tagsValue types.List

		tagsValue, d = types.ListValue(types.StringType, utils.Map(limit.Tags, func(tag string) tfattr.Value {
			return types.StringValue(tag)
		}))

		diags.Append(d...)
		m.Tags = tagsValue
	}

	return diags
}
