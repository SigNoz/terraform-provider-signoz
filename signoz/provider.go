package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	signozdatasource "github.com/SigNoz/terraform-provider-signoz/signoz/internal/provider/datasource"
	signozresource "github.com/SigNoz/terraform-provider-signoz/signoz/internal/provider/resource"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
)

const (
	DefaultHTTPTimeout  = "35"
	DefaultHTTPMaxRetry = "10"
	DefaultURL          = "http://localhost:3301"

	// Environment variables.
	EnvAccessToken  = "SIGNOZ_ACCESS_TOKEN" // #nosec G101
	EnvEndpoint     = "SIGNOZ_ENDPOINT"
	EnvHTTPMaxRetry = "SIGNOZ_HTTP_MAX_RETRY"
	EnvHTTPTimeout  = "SIGNOZ_HTTP_TIMEOUT"
)

// signozProviderModel maps provider schema data to a Go type.
type signozProviderModel struct {
	AccessToken  types.String `tfsdk:"access_token"`
	Endpoint     types.String `tfsdk:"endpoint"`
	HTTPMaxRetry types.Int64  `tfsdk:"http_max_retry"`
	HTTPTimeout  types.Int64  `tfsdk:"http_timeout"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &signozProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(terraformAgent, version string) func() provider.Provider {
	return func() provider.Provider {
		return &signozProvider{
			terraformAgent: terraformAgent,
			version:        version,
		}
	}
}

// signozProvider is the provider implementation.
type signozProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	// terraformAgent is the name of the terraform agent to use.
	terraformAgent string
}

// Metadata returns the provider type name.
func (p *signozProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "signoz"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *signozProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			attr.AccessToken: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: fmt.Sprintf("Access token of the SigNoz API. You can retrieve it from SigNoz UI\n"+
					"with Admin Role ([documentation](https://signoz.io/newsroom/launch-week-1-day-5/#using-access-token)).\n"+
					"Also, you can set it using environment variable %s.", EnvAccessToken),
			},
			attr.Endpoint: schema.StringAttribute{
				Optional: true,
				Description: fmt.Sprintf("Endpoint of the SigNoz. It is the root URL of the SigNoz UI.\n"+
					"Also, you can set it using environment variable %s. If not set, it defaults to %s.", EnvEndpoint, DefaultURL),
			},
			attr.HTTPMaxRetry: schema.Int64Attribute{
				Optional: true,
				Description: fmt.Sprintf("Specifies the max retry limit for the HTTP requests made to SigNoz.\n"+
					"Also, you can set it using environment variable %s. If not set, it defaults to %s.", EnvHTTPMaxRetry, DefaultHTTPMaxRetry),
			},
			attr.HTTPTimeout: schema.Int64Attribute{
				Optional: true,
				Description: fmt.Sprintf("Specifies the timeout limit in seconds for the HTTP requests made to SigNoz.\n"+
					"Also, you can set it using environment variable %s. If not set, it defaults to %s.", EnvHTTPTimeout, DefaultHTTPTimeout),
			},
		},
	}
}

// Configure prepares a SigNoz API client for data sources and resources.
func (p *signozProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring SigNoz client")

	// Retrieve provider data from configuration
	var config signozProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	accessToken := os.Getenv(EnvAccessToken)
	endpoint := utils.WithDefault(os.Getenv(EnvEndpoint), DefaultURL)
	httpMaxRetry := utils.MustGetInt(utils.WithDefault(os.Getenv(EnvHTTPMaxRetry), DefaultHTTPMaxRetry))
	httpTimeout := utils.MustGetInt(utils.WithDefault(os.Getenv(EnvHTTPTimeout), DefaultHTTPTimeout))

	accessToken = utils.OverrideStrWithConfig(config.AccessToken, accessToken)
	endpoint = utils.OverrideStrWithConfig(config.Endpoint, endpoint)
	httpMaxRetry = utils.OverrideIntWithConfig(config.HTTPMaxRetry, httpMaxRetry)
	httpTimeout = utils.OverrideIntWithConfig(config.HTTPTimeout, httpTimeout)

	// Check if the SigNoz access token has been set in the configuration or
	// environment variables. If not, return an error.
	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(attr.AccessToken),
			"Missing SigNoz "+attr.AccessToken,
			fmt.Sprintf("The provider cannot create the SigNoz API client as there is a missing or empty value for the SigNoz API %s. "+
				"Set the %s value in the configuration or use the %s environment variable. "+
				"If either is already set, ensure the value is not empty.", attr.AccessToken, attr.AccessToken, EnvAccessToken),
		)

		return
	}

	// Create a new SigNoz client using the configuration values
	client := client.NewClient(
		endpoint,
		accessToken,
		time.Duration(httpTimeout)*time.Second,
		httpMaxRetry,
		p.terraformAgent,
		p.version,
	)

	// Make the SigNoz client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured SigNoz client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *signozProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		signozdatasource.NewAlertDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *signozProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		signozresource.NewAlertResource,
	}
}
