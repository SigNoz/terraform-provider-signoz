package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	conv "github.com/SigNoz/terraform-provider-signoz/internal/convertors"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = (*pipelinesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*pipelinesDataSource)(nil)
)

type pipelinesDataSource struct {
	api *apiclients.WrappedClient
}

// NewPipelinesDataSource returns the framework datasource constructor.
func NewPipelinesDataSource() datasource.DataSource {
	return &pipelinesDataSource{}
}

func (d *pipelinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipelines"
}

func (d *pipelinesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schemas.PipelinesDataSourceSchema(ctx)
}

func (d *pipelinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = c
}

func (d *pipelinesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg schemas.PipelinesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := "latest"
	if !cfg.Version.IsNull() && !cfg.Version.IsUnknown() {
		version = strconv.FormatInt(cfg.Version.ValueInt64(), 10)
	}

	raw, err := d.api.Do(ctx, http.MethodGet, pipelinesPath+"/"+version, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read pipelines datasource", err.Error())
		return
	}

	var envelope apitypes.LogparsingpipelineApiResponse
	if err := json.Unmarshal(raw, &envelope); err != nil {
		resp.Diagnostics.AddError("Read pipelines datasource", err.Error())
		return
	}

	next, diags := conv.FlattenLogparsingpipelinePipelinesResponse(ctx, &envelope.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
}
