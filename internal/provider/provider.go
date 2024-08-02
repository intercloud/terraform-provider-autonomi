package provider

import (
	"context"
	"terraform-provider-autonomi/external/autonomi"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Host               types.String `tfsdk:"host"`
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
			"host": schema.StringAttribute{
				Required: true,
			},
			"terms_and_conditions": schema.BoolAttribute{
				Required: true,
			},
			"personal_access_token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
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
	if !config.TermsAndConditions.IsNull() {
		terms_and_conditions = config.TermsAndConditions.ValueBool()
	}
	if !config.PAT.IsNull() {
		personal_access_token = config.PAT.ValueString()
	}

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
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Autonomi client using the configuration values
	client, err := autonomi.NewClient(personal_access_token, terms_and_conditions)
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
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *autonomiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *autonomiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWorkspaceResource,
		NewCloudNodeResource,
	}
}
