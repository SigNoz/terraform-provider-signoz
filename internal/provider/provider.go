// Package provider exposes the SigNoz Terraform provider built on top of
// terraform-plugin-framework. The provider configures a single
// `*client.Client` and ships it to every resource and datasource via
// `resp.ResourceData` / `resp.DataSourceData`.
package provider

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/services"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	envEndpoint     = "SIGNOZ_ENDPOINT"
	envAccessToken  = "SIGNOZ_ACCESS_TOKEN"
	envHTTPMaxRetry = "SIGNOZ_HTTP_MAX_RETRY"
	envHTTPTimeout  = "SIGNOZ_HTTP_TIMEOUT"
)

var _ provider.Provider = (*signozProvider)(nil)

type signozProvider struct {
	terraformAgent string
	version        string
}

type signozProviderModel struct {
	AccessToken  types.String `tfsdk:"access_token"`
	Endpoint     types.String `tfsdk:"endpoint"`
	HTTPMaxRetry types.Int64  `tfsdk:"http_max_retry"`
	HTTPTimeout  types.String `tfsdk:"http_timeout"`
}

// New returns the framework provider constructor used by `main.go`.
func New(terraformAgent, version string) func() provider.Provider {
	return func() provider.Provider {
		return &signozProvider{terraformAgent: terraformAgent, version: version}
	}
}

func (p *signozProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "signoz"
	resp.Version = p.version
}

func (p *signozProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SigNoz API access token. Read from " + envAccessToken + " when unset.",
			},
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "SigNoz API endpoint URL. Defaults to " + apiclients.DefaultEndpoint + " or " + envEndpoint + ".",
			},
			"http_max_retry": schema.Int64Attribute{
				Optional:    true,
				Description: "Max HTTP retries on transient failures. Read from " + envHTTPMaxRetry + " when unset.",
			},
			"http_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP request timeout (Go duration, e.g. \"35s\"). Read from " + envHTTPTimeout + " when unset.",
			},
		},
	}
}

func (p *signozProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg signozProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := firstNonEmpty(cfg.Endpoint.ValueString(), os.Getenv(envEndpoint))
	token := firstNonEmpty(cfg.AccessToken.ValueString(), os.Getenv(envAccessToken))
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing SigNoz access token",
			"Set provider.access_token, or the "+envAccessToken+" env var.",
		)
		return
	}

	retryMax := 0
	switch {
	case !cfg.HTTPMaxRetry.IsNull() && !cfg.HTTPMaxRetry.IsUnknown():
		retryMax = int(cfg.HTTPMaxRetry.ValueInt64())
	default:
		if v := os.Getenv(envHTTPMaxRetry); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				retryMax = n
			}
		}
	}

	var timeout time.Duration
	if v := firstNonEmpty(cfg.HTTPTimeout.ValueString(), os.Getenv(envHTTPTimeout)); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("http_timeout"),
				"Invalid http_timeout",
				err.Error(),
			)
			return
		}
		timeout = d
	}

	c, err := apiclients.New(endpoint, token, p.terraformAgent, p.version, timeout, retryMax)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SigNoz client", err.Error())
		return
	}

	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *signozProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		services.NewAuthDomainResource,
		services.NewNotificationChannelResource,
		services.NewPlannedMaintenanceResource,
		services.NewRoutePolicyResource,
		services.NewRuleResource,
	}
}

func (p *signozProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		services.NewAuthDomainDataSource,
		services.NewNotificationChannelDataSource,
		services.NewPlannedMaintenanceDataSource,
		services.NewRoutePolicyDataSource,
		services.NewRuleDataSource,
	}
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}
