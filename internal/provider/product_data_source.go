package provider

import (
	"context"
	"fmt"

	"github.com/intercloud/terraform-provider-autonomi/external/products"
	"github.com/intercloud/terraform-provider-autonomi/external/products/models"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type cloudProductDataSource struct {
	client *products.Client
}

type cloudProductDataSourceModel struct {
	ID        types.Int64   `tfsdk:"id"`
	Provider  types.String  `tfsdk:"provider"`
	Duration  types.Int64   `tfsdk:"duration"`
	Location  types.String  `tfsdk:"location"`
	Bandwidth types.Int64   `tfsdk:"bandwidth"`
	Date      types.String  `tfsdk:"date"`
	PriceNRC  types.Float64 `tfsdk:"price_nrc"`
	PriceMRC  types.Float64 `tfsdk:"price_mrc"`
	CostNRC   types.Float64 `tfsdk:"cost_nrc"`
	CostMRC   types.Float64 `tfsdk:"cost_mrc"`
	SKU       types.String  `tfsdk:"sku"`
	CSPName   types.String  `tfsdk:"csp_name"`
}

type productDataSourceModel struct {
	CSPName          types.String                  `tfsdk:"csp_name"`
	UnderlayProvider types.String                  `tfsdk:"underlay_provider"`
	Location         types.String                  `tfsdk:"location"`
	Bandwidth        types.String                  `tfsdk:"bandwidth"`
	Hits             []cloudProductDataSourceModel `tfsdk:"hits"`
}

// Ensure the implementation satisfies the expected interfaces.
// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProductDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProductDataSource{}
)

func NewCloudProductDataSource() datasource.DataSource {
	return &cloudProductDataSource{}
}

// Metadata returns the data source type name.
func (d *cloudProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_products"
}

// Schema defines the schema for the data source.
func (d *cloudProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"csp_name": schema.StringAttribute{
				MarkdownDescription: "Name of the CSP expected values are [AWS, Azure, GCP]",
				Optional:            true,
			},
			"underlay_provider": schema.StringAttribute{
				MarkdownDescription: "Name of the Provider: expected values are [Equinix, Megaport]",
				Optional:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Name of the Location: expected values are [...]",
				Optional:            true,
			},
			"bandwidth": schema.StringAttribute{
				MarkdownDescription: "Name of the Provider: expected values are [50, 100, 110, 500, 1000, 5000, 10000]",
				Optional:            true,
			},
			"hits": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"provider": schema.StringAttribute{
							Computed: true,
						},
						"duration": schema.Int64Attribute{
							Computed: true,
						},
						"location": schema.StringAttribute{
							Computed: true,
						},
						"bandwidth": schema.Int64Attribute{
							Computed: true,
						},
						"date": schema.StringAttribute{
							Computed: true,
						},
						"price_nrc": schema.Int64Attribute{
							Computed: true,
						},
						"price_mrc": schema.Int64Attribute{
							Computed: true,
						},
						"cost_nrc": schema.Int64Attribute{
							Computed: true,
						},
						"cost_mrc": schema.Int64Attribute{
							Computed: true,
						},
						"sku": schema.StringAttribute{
							Computed: true,
						},
						"csp_name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *cloudProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	catalogClient, ok := req.ProviderData.(*products.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *catalog.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = catalogClient
}

// Read refreshes the Terraform state with the latest data.
func (d *cloudProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data productDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := models.Filters{
		CSPName:   data.CSPName.ValueString(),
		Provider:  data.UnderlayProvider.ValueString(),
		Location:  data.Location.ValueString(),
		Bandwidth: data.Bandwidth.ValueString(),
	}

	var state productDataSourceModel

	products, err := d.client.GetCloudProducts(filters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Cloud Products",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, cp := range products {
		cloudProductState := cloudProductDataSourceModel{
			ID:        types.Int64Value(int64(cp.ID)),
			Provider:  types.StringValue(cp.Provider),
			Duration:  types.Int64Value(int64(cp.Duration)),
			Location:  types.StringValue(cp.Location),
			Bandwidth: types.Int64Value(int64(cp.Bandwidth)),
			Date:      types.StringValue(cp.Date),
			PriceNRC:  types.Float64Value(float64(cp.PriceNRC)),
			PriceMRC:  types.Float64Value(float64(cp.PriceMRC)),
			CostNRC:   types.Float64Value(float64(cp.CostNRC)),
			CostMRC:   types.Float64Value(float64(cp.CostMRC)),
			SKU:       types.StringValue(cp.SKU),
			CSPName:   types.StringValue(cp.CSPName),
		}
		state.Hits = append(state.Hits, cloudProductState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
