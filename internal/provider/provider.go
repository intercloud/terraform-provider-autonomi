package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	autonomisdk "github.com/intercloud/autonomi-sdk"
	datasources "github.com/intercloud/terraform-provider-autonomi/internal/data_sources"
	autonomiresource "github.com/intercloud/terraform-provider-autonomi/internal/resources"
	"github.com/meilisearch/meilisearch-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &autonomiProvider{}
)

// autonomiProvider is the provider implementation.
type autonomiProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type autonomiProviderModel struct {
	TermsAndConditions types.Bool   `tfsdk:"terms_and_conditions"`
	PAT                types.String `tfsdk:"personal_access_token"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &autonomiProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *autonomiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "autonomi"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *autonomiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"terms_and_conditions": schema.BoolAttribute{
				MarkdownDescription: "Terms and conditions",
				Required:            true,
				Description:         "A boolean variable indicating whether the terms and conditions have been accepted. Must be set to 'true' to run the provider",
			},
			"personal_access_token": schema.StringAttribute{
				MarkdownDescription: "Personal Access Token (PAT) to authenticate through Autonomi API. This token can be obtained from the Autonomi service and is required to access and manage resources via the API. Can be set as variable or in environment as AUTONOMI_PAT",
				Optional:            true,
				Sensitive:           true,
				Description:         "The Personal Access Token (PAT) used to authenticate with the Autonomi API. This token can be obtained from the Autonomi service and is required to access and manage resources via the API. Can be set as variable or in environment as AUTONOMI_PAT",
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *autonomiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config autonomiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	var terms_and_conditions bool
	var personal_access_token string
	var host_url, catalog_url string
	if !config.TermsAndConditions.IsNull() {
		terms_and_conditions = config.TermsAndConditions.ValueBool()
	}
	if !config.PAT.IsNull() {
		personal_access_token = config.PAT.ValueString()
	} else {
		personal_access_token = os.Getenv("AUTONOMI_PAT")
	}
	host_url = os.Getenv("AUTONOMI_HOST_URL")
	catalog_url = os.Getenv("AUTONOMI_CATALOG_URL")

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if !terms_and_conditions {
		resp.Diagnostics.AddAttributeError(
			path.Root("terms_and_conditions"),
			"API Terms and Conditions not accepted",
			"The provider cannot create the Autonomi API client because the terms_and_conditions configuration value is not set to true."+
				"Please explicitly set the terms_and_conditions value to true in your Terraform configuration or use se the AUTONOMI_TERMS_AND_CONDITIONS environment variable and set it to 'true'.",
		)
	}
	if personal_access_token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("personal_access_token"),
			"Empty API Personal Access Token",
			"The provider cannot create the Autonomi API client because the personal access token (PAT) is not set."+
				"Please explicitly set the personal_access_token value in your Terraform configuration or use the AUTONOMI_PAT environment variable to provide the token.",
		)
	}
	if host_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host_url"),
			"Empty API Host URL",
			"The provider cannot create the Autonomi API client because the host url is not set."+
				"Please explicitly set the host_url value in your Terraform configuration or use the HOST_URL environment variable to provide the token.",
		)
	}
	hostURL, err := url.Parse(host_url)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse host URL",
			"An unexpected error occurred when parsing host url "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Autonomi Client Error: "+err.Error(),
		)
		return
	}
	if catalog_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("catalog_url"),
			"Empty Catalog URL",
			"The provider cannot create the Autonomi API client because the catalog url is not set."+
				"Please explicitly set the catalog_url value in your Terraform configuration or use the CATALOG_URL environment variable to provide the token.",
		)
	}

	// Create a new Catalog client using the configuration values
	catalogClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   catalog_url,
		APIKey: personal_access_token,
	})

	// Create a Autonomi client using the configuration values
	client, err := autonomisdk.NewClient(terms_and_conditions,
		autonomisdk.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec //No
				},
			},
		}),
		autonomisdk.WithHostURL(hostURL),
		autonomisdk.WithPersonalAccessToken(personal_access_token),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Autonomi API Client",
			"An unexpected error occurred when creating the Autonomi API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Autonomi Client Error: "+err.Error(),
		)
		return
	}

	// Make the Autonomi client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = catalogClient
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *autonomiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewCloudProductDataSource,
		datasources.NewTransportProductDataSource,
		datasources.NewAccessProductDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *autonomiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		autonomiresource.NewWorkspaceResource,
		autonomiresource.NewCloudNodeResource,
		autonomiresource.NewAccessNodeResource,
		autonomiresource.NewTransportResource,
		autonomiresource.NewAttachmentResource,
	}
}
