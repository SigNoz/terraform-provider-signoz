package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	signozdatasource "github.com/SigNoz/terraform-provider-signoz/signoz/internal/provider/datasource"
	signozresource "github.com/SigNoz/terraform-provider-signoz/signoz/internal/provider/resource"
)

const (
	DefaultHTTPTimeout  = "35"
	DefaultHTTPMaxRetry = "10"
	DefaultURL          = "http://localhost:3301"

	// Environment variables
	EnvAccessToken = "SIGNOZ_ACCESS_TOKEN" // #nosec G101
	EnvEndpoint    = "SIGNOZ_ENDPOINT"
)

// signozProviderModel maps provider schema data to a Go type.
type signozProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	AccessToken types.String `tfsdk:"access_token"`
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
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Endpoint of the SigNoz API. " + DefaultURL + " by default.",
			},
			"access_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Access token of the SigNoz API.",
			},
		},
	}
}

// Configure prepares a SigNoz API client for data sources and resources.
func (p *signozProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring SigNoz client")

	// Retrieve provider data from configuration
	var config signozProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown SigNoz API Endpoint",
			"The provider cannot create the SigNoz API client as there is an unknown configuration value for the SigNoz API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SIGNOZ_ENDPOINT environment variable.",
		)
	}
	if config.AccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Unknown SigNoz API Access Token",
			"The provider cannot create the SigNoz API client as there is an unknown configuration value for the SigNoz API access token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SIGNOZ_ACCESS_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	endpoint := os.Getenv(EnvEndpoint)
	accessToken := os.Getenv(EnvAccessToken)
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.AccessToken.IsNull() {
		accessToken = config.AccessToken.ValueString()
	}
	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if endpoint == "" {
		tflog.Warn(ctx, "Missing SigNoz API Endpoint, using default endpoint "+DefaultURL)
		endpoint = DefaultURL
	}
	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing SigNoz API Access Token",
			"The provider cannot create the SigNoz API client as there is a missing or empty value for the SigNoz API access token. "+
				"Set the access_token value in the configuration or use the SIGNOZ_ACCESS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "signoz_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "signoz_access_token", accessToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "signoz_access_token")

	tflog.Debug(ctx, "Creating SigNoz client")

	// Create a new SigNoz client using the configuration values
	client, err := client.NewClient(&endpoint, &accessToken)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create SigNoz API Client",
			"An unexpected error occurred when creating the SigNoz API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"SigNoz Client Error: "+err.Error(),
		)
		return
	}
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
