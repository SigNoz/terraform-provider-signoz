package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	SigNozDataRetention         = "signoz_data_retention"
	dataRetentionSyntheticID    = "settings"
	dataRetentionDefaultTTLHrs  = 720
	dataRetentionDefaultMoveHrs = 0
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dataRetentionResource{}
	_ resource.ResourceWithConfigure   = &dataRetentionResource{}
	_ resource.ResourceWithImportState = &dataRetentionResource{}
)

// NewDataRetentionResource is a helper function to simplify the provider implementation.
func NewDataRetentionResource() resource.Resource {
	return &dataRetentionResource{}
}

// dataRetentionResource is the resource implementation.
type dataRetentionResource struct {
	client *legacyclient.Client
}

// dataRetentionResourceModel maps the resource schema data.
type dataRetentionResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	MetricsTTLDurationHrs     types.Int64  `tfsdk:"metrics_ttl_duration_hrs"`
	MetricsMoveTTLDurationHrs types.Int64  `tfsdk:"metrics_move_ttl_duration_hrs"`
	LogsTTLDurationHrs        types.Int64  `tfsdk:"logs_ttl_duration_hrs"`
	LogsMoveTTLDurationHrs    types.Int64  `tfsdk:"logs_move_ttl_duration_hrs"`
	TracesTTLDurationHrs      types.Int64  `tfsdk:"traces_ttl_duration_hrs"`
	TracesMoveTTLDurationHrs  types.Int64  `tfsdk:"traces_move_ttl_duration_hrs"`
	ColdStorageVolume         types.String `tfsdk:"cold_storage_volume"`
}

// Configure adds the provider configured client to the resource.
func (r *dataRetentionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, SigNozDataRetention,
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
func (r *dataRetentionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "signoz_data_retention"
}

// Schema defines the schema for the resource.
func (r *dataRetentionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages data retention settings in SigNoz. This is a singleton resource — only one instance should exist per SigNoz deployment.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic ID for the data retention settings. Always \"settings\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.MetricsTTLDuration: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(dataRetentionDefaultTTLHrs),
				Description: "Hours to retain metrics data. Default is 720 (30 days).",
			},
			attr.MetricsMoveTTLDuration: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(dataRetentionDefaultMoveHrs),
				Description: "Hours before moving metrics to cold storage. 0 means disabled.",
			},
			attr.LogsTTLDuration: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(dataRetentionDefaultTTLHrs),
				Description: "Hours to retain logs data (whole days; set via the v2 custom-retention API). Default is 720 (30 days).",
			},
			attr.LogsMoveTTLDuration: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(dataRetentionDefaultMoveHrs),
				Description: "Hours before moving logs to cold storage (whole days). 0 means disabled; requires cold_storage_volume.",
			},
			attr.TracesTTLDuration: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(dataRetentionDefaultTTLHrs),
				Description: "Hours to retain traces data. Default is 720 (30 days).",
			},
			attr.TracesMoveTTLDuration: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(dataRetentionDefaultMoveHrs),
				Description: "Hours before moving traces to cold storage. 0 means disabled.",
			},
			attr.ColdStorageVolume: schema.StringAttribute{
				Optional:    true,
				Description: "Cold storage volume applied to every signal with a move TTL set. The server only performs a cold-storage move when a volume is configured; without it move settings are ignored.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dataRetentionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataRetentionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// POST once per signal type.
	err := r.setAllRetention(ctx, &plan)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozDataRetention)
		return
	}

	// Set synthetic ID.
	plan.ID = types.StringValue(dataRetentionSyntheticID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *dataRetentionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataRetentionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading data retention settings")

	retention, err := r.client.GetDataRetention(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "Data retention settings not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err, operationRead, SigNozDataRetention)
		return
	}

	// API returns -1 for "not explicitly configured" — preserve state value in that case.
	setIfPositive := func(current types.Int64, apiVal int) types.Int64 {
		if apiVal >= 0 {
			return types.Int64Value(int64(apiVal))
		}
		return current
	}
	state.ID = types.StringValue(dataRetentionSyntheticID)
	state.MetricsTTLDurationHrs = setIfPositive(state.MetricsTTLDurationHrs, retention.MetricsTTLDurationHrs)
	state.MetricsMoveTTLDurationHrs = setIfPositive(state.MetricsMoveTTLDurationHrs, retention.MetricsMoveTTLDurationHrs)
	state.LogsTTLDurationHrs = setIfPositive(state.LogsTTLDurationHrs, retention.LogsTTLDurationHrs)
	state.LogsMoveTTLDurationHrs = setIfPositive(state.LogsMoveTTLDurationHrs, retention.LogsMoveTTLDurationHrs)
	state.TracesTTLDurationHrs = setIfPositive(state.TracesTTLDurationHrs, retention.TracesTTLDurationHrs)
	state.TracesMoveTTLDurationHrs = setIfPositive(state.TracesMoveTTLDurationHrs, retention.TracesMoveTTLDurationHrs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dataRetentionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dataRetentionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.setAllRetention(ctx, &plan)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDataRetention)
		return
	}

	plan.ID = types.StringValue(dataRetentionSyntheticID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete removes the resource from Terraform state but does not delete retention settings.
func (r *dataRetentionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "Removing signoz_data_retention from Terraform state. "+
		"Data retention settings persist in SigNoz and are not deleted.")
	// State is automatically removed by the framework.
}

// ImportState imports existing data retention settings into Terraform state.
func (r *dataRetentionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Ignore the import ID — this is a singleton. Set the synthetic ID and let Read populate the rest.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), dataRetentionSyntheticID)...)
}

// setAllRetention POSTs retention settings for all three signal types.
// Metrics/traces go to the v1 TTL endpoint (query params); logs go to the v2
// custom-retention endpoint (days), which requires whole-day values.
func (r *dataRetentionResource) setAllRetention(ctx context.Context, plan *dataRetentionResourceModel) error {
	volume := plan.ColdStorageVolume.ValueString()

	signals := []struct {
		signalType  string
		durationHrs int64
		moveHrs     int64
	}{
		{attr.Metrics, plan.MetricsTTLDurationHrs.ValueInt64(), plan.MetricsMoveTTLDurationHrs.ValueInt64()},
		{attr.Logs, plan.LogsTTLDurationHrs.ValueInt64(), plan.LogsMoveTTLDurationHrs.ValueInt64()},
		{attr.Traces, plan.TracesTTLDurationHrs.ValueInt64(), plan.TracesMoveTTLDurationHrs.ValueInt64()},
	}

	for _, s := range signals {
		if s.moveHrs > 0 && volume == "" {
			return fmt.Errorf("%s_move_ttl_duration_hrs is %d but cold_storage_volume is not set: "+
				"the server ignores cold-storage moves without a volume", s.signalType, s.moveHrs)
		}

		if s.signalType == attr.Logs {
			if s.durationHrs%24 != 0 {
				return fmt.Errorf("logs_ttl_duration_hrs must be whole days (a multiple of 24); got %d", s.durationHrs)
			}
			if s.moveHrs%24 != 0 {
				return fmt.Errorf("logs_move_ttl_duration_hrs must be whole days (a multiple of 24); got %d", s.moveHrs)
			}
		}

		payload := &model.DataRetentionUpdate{
			Type:               s.signalType,
			DurationHrs:        s.durationHrs,
			ColdStorageVolume:  volume,
			ColdStorageMoveHrs: s.moveHrs,
		}

		tflog.Debug(ctx, "Setting data retention", map[string]any{
			"type":     s.signalType,
			"ttl_hrs":  payload.DurationHrs,
			"move_hrs": payload.ColdStorageMoveHrs,
		})

		err := r.client.SetDataRetention(ctx, payload)
		if err != nil {
			return fmt.Errorf("failed to set %s retention: %w", s.signalType, err)
		}
	}

	return nil
}
