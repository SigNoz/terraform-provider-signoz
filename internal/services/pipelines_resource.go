package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	conv "github.com/SigNoz/terraform-provider-signoz/internal/convertors"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The logs pipelines API manages the org's full pipeline list as one versioned
// object (POST replaces the list, GET /latest reads it) — there is no
// per-pipeline CRUD, so the resource is a singleton with a constant id. The
// API has no generated client (see apitypes/pipelines.go), so requests go
// through apiclients.WrappedClient.Do.
const (
	pipelinesID   = "pipelines"
	pipelinesPath = "/api/v1/logs/pipelines"
)

var (
	_ resource.Resource                = (*pipelinesResource)(nil)
	_ resource.ResourceWithConfigure   = (*pipelinesResource)(nil)
	_ resource.ResourceWithImportState = (*pipelinesResource)(nil)
)

type pipelinesResource struct {
	api *apiclients.WrappedClient
}

// NewPipelinesResource returns the framework resource constructor.
func NewPipelinesResource() resource.Resource {
	return &pipelinesResource{}
}

func (r *pipelinesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipelines"
}

func (r *pipelinesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schemas.PipelinesResourceSchema(ctx)
}

func (r *pipelinesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *apiclients.WrappedClient, got %T. This is a provider bug.", req.ProviderData),
		)
		return
	}
	r.api = c
}

func (r *pipelinesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan schemas.PipelinesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, plan, "Create pipelines")...)
	if resp.Diagnostics.HasError() {
		return
	}

	next, diags := r.readLatest(ctx, "Refresh pipelines after create")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, pipelinesResourceFromDS(next))...)
}

func (r *pipelinesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state schemas.PipelinesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	next, diags := r.readLatest(ctx, "Read pipelines")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, pipelinesResourceFromDS(next))...)
}

func (r *pipelinesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan schemas.PipelinesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, plan, "Update pipelines")...)
	if resp.Diagnostics.HasError() {
		return
	}

	next, diags := r.readLatest(ctx, "Refresh pipelines after update")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, pipelinesResourceFromDS(next))...)
}

// Delete posts an empty list — the API has no delete endpoint; an empty
// full-replacement save is how a pipeline set is removed.
func (r *pipelinesResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	body, err := json.Marshal(apitypes.PipelinetypesPostablePipelines{
		Pipelines: []apitypes.PipelinetypesPostablePipeline{},
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete pipelines", err.Error())
		return
	}

	if _, err := r.api.Do(ctx, http.MethodPost, pipelinesPath, bytes.NewReader(body)); err != nil {
		resp.Diagnostics.AddError("Delete pipelines", err.Error())
	}
}

func (r *pipelinesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// apply posts the planned pipeline list, replacing the server's list as a new
// config version.
func (r *pipelinesResource) apply(ctx context.Context, plan schemas.PipelinesModel, summary string) diag.Diagnostics {
	in, diags := conv.ExpandPipelinetypesPostablePipelines(ctx, plan)
	if diags.HasError() {
		return diags
	}

	body, err := json.Marshal(in)
	if err != nil {
		diags.AddError(summary, err.Error())
		return diags
	}

	if _, err := r.api.Do(ctx, http.MethodPost, pipelinesPath, bytes.NewReader(body)); err != nil {
		diags.AddError(summary, err.Error())
	}

	return diags
}

// readLatest fetches the latest pipelines config version — the create/update
// follow-up read of this resource, so state always lands via one Flatten path.
func (r *pipelinesResource) readLatest(ctx context.Context, summary string) (*schemas.PipelinesDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	raw, err := r.api.Do(ctx, http.MethodGet, pipelinesPath+"/latest", nil)
	if err != nil {
		diags.AddError(summary, err.Error())
		return nil, diags
	}

	var envelope apitypes.LogparsingpipelineApiResponse
	if err := json.Unmarshal(raw, &envelope); err != nil {
		diags.AddError(summary, err.Error())
		return nil, diags
	}

	next, d := conv.FlattenLogparsingpipelinePipelinesResponse(ctx, &envelope.Data)
	diags.Append(d...)

	return next, diags
}

// pipelinesResourceFromDS narrows the wide datasource model down to the
// resource model — drops the server-set config version.
func pipelinesResourceFromDS(ds *schemas.PipelinesDataSourceModel) *schemas.PipelinesModel {
	return &schemas.PipelinesModel{
		Id:        types.StringValue(pipelinesID),
		Pipelines: ds.Pipelines,
	}
}
