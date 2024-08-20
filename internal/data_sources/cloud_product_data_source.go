package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/intercloud/terraform-provider-autonomi/external/products/models"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/meilisearch/meilisearch-go"
)

type cloudProductDataSource struct {
	client *meilisearch.Client
}

type cloudHits struct {
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

type cloudFacetDistributionDataSourceModel struct {
	Bandwidth map[string]int `tfsdk:"bandwidth"`
	CSPCity   map[string]int `tfsdk:"csp_city"`
	CSPName   map[string]int `tfsdk:"csp_name"`
	CSPRegion map[string]int `tfsdk:"csp_region"`
	Location  map[string]int `tfsdk:"location"`
	Provider  map[string]int `tfsdk:"provider"`
}

type cloudsProductDataSourceModel struct {
	CSPName           types.String                           `tfsdk:"csp_name"`
	CSPCity           types.String                           `tfsdk:"csp_city"`
	CSPRegion         types.String                           `tfsdk:"csp_region"`
	UnderlayProvider  types.String                           `tfsdk:"underlay_provider"`
	Location          types.String                           `tfsdk:"location"`
	Bandwidth         types.Int64                            `tfsdk:"bandwidth"`
	Hits              []cloudHits                            `tfsdk:"hits"`
	FacetDistribution *cloudFacetDistributionDataSourceModel `tfsdk:"facet_distribution"`
}

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
			"csp_city": schema.StringAttribute{
				MarkdownDescription: "Name of the CSP city",
				Optional:            true,
			},
			"csp_region": schema.StringAttribute{
				MarkdownDescription: "Name of the CSP region",
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
			"bandwidth": schema.Int64Attribute{
				MarkdownDescription: "Name of the Provider: expected values are [50, 100, 110, 500, 1000, 5000, 10000]",
				Optional:            true,
			},
			"hits": schema.ListNestedAttribute{
				MarkdownDescription: "The **hits** attribute contains the list of cloud products returned by the Meilisearch query. Each hit represents a cloud product that matches the specified search criteria.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.Int64Attribute{Computed: true},
						"provider":  schema.StringAttribute{Computed: true},
						"duration":  schema.Int64Attribute{Computed: true},
						"location":  schema.StringAttribute{Computed: true},
						"bandwidth": schema.Int64Attribute{Computed: true},
						"date":      schema.StringAttribute{Computed: true},
						"price_nrc": schema.Int64Attribute{Computed: true},
						"price_mrc": schema.Int64Attribute{Computed: true},
						"cost_nrc":  schema.Int64Attribute{Computed: true},
						"cost_mrc":  schema.Int64Attribute{Computed: true},
						"sku":       schema.StringAttribute{Computed: true},
						"csp_name":  schema.StringAttribute{Computed: true},
					},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: "The **facet_distribution** attribute provides an overview of the distribution of various facets within the cloud products returned by the Meilisearch query. This attribute allows you to analyze the frequency of different categories or attributes in the search results.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"bandwidth":  int64MapAttr,
					"csp_city":   int64MapAttr,
					"csp_name":   int64MapAttr,
					"csp_region": int64MapAttr,
					"location":   int64MapAttr,
					"provider":   int64MapAttr,
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

	catalogClient, ok := req.ProviderData.(*meilisearch.Client)
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

	var data cloudsProductDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := models.CloudFilters{
		CSPName:   data.CSPName.ValueString(),
		Provider:  models.ProviderType(data.UnderlayProvider.ValueString()),
		Location:  data.Location.ValueString(),
		Bandwidth: int(data.Bandwidth.ValueInt64()),
	}

	// Create the filter string dynamically
	var filterStrings []string

	if filters.CSPName != "" {
		filterStrings = append(filterStrings, fmt.Sprintf("cspName = \"%s\"", filters.CSPName))
	}
	if filters.CSPCity != "" {
		filterStrings = append(filterStrings, fmt.Sprintf("cspCity = \"%s\"", filters.CSPName))
	}
	if filters.CSPRegion != "" {
		filterStrings = append(filterStrings, fmt.Sprintf("cspRegion = \"%s\"", filters.CSPName))
	}
	if filters.Provider != "" {
		filterStrings = append(filterStrings, fmt.Sprintf("provider = \"%s\"", filters.Provider))
	}
	if filters.Location != "" {
		filterStrings = append(filterStrings, fmt.Sprintf("location = \"%s\"", filters.Location))
	}
	if filters.Bandwidth != 0 {
		filterStrings = append(filterStrings, fmt.Sprintf("bandwidth = %d", filters.Bandwidth))
	}

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filterStrings,
		Facets: []string{
			"cspName",
			"cspRegion",
			"cspCity",
			"location",
			"bandwidth",
			"provider",
		},
	}

	respProducts, err := d.client.Index("cloudproduct").Search("", searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Cloud Products",
			err.Error(),
		)
		return
	}

	cloudProducts := models.CloudProducts{}
	productsJSON, err := json.Marshal(respProducts) // Marshal the hits to JSON
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Cloud Products",
			err.Error(),
		)
		return
	}

	err = json.Unmarshal(productsJSON, &cloudProducts) // Unmarshal JSON into the CloudProduct slice
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Cloud Products",
			err.Error(),
		)
		return
	}

	state := cloudsProductDataSourceModel{
		CSPName:          data.CSPName,
		CSPCity:          data.CSPCity,
		CSPRegion:        data.CSPRegion,
		UnderlayProvider: data.UnderlayProvider,
		Bandwidth:        data.Bandwidth,
		Location:         data.Location,
	}

	// Map response body to model
	for _, cp := range cloudProducts.Hits {
		cloudProductState := cloudHits{
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

	// Set the bandwidth map in the state
	state.FacetDistribution = &cloudFacetDistributionDataSourceModel{
		Bandwidth: cloudProducts.FacetDistribution.Bandwidth,
		CSPCity:   cloudProducts.FacetDistribution.CSPCity,
		CSPName:   cloudProducts.FacetDistribution.CSPName,
		CSPRegion: cloudProducts.FacetDistribution.CSPRegion,
		Location:  cloudProducts.FacetDistribution.Location,
		Provider:  cloudProducts.FacetDistribution.Provider,
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
