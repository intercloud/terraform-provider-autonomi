package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/intercloud/terraform-provider-autonomi/internal/data_sources/filters"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/meilisearch/meilisearch-go"
)

type cloudProductsDataSource struct {
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

type cloudsProductsDataSourceModel struct {
	Filters           []filters.Filter                       `tfsdk:"filters"`
	Sort              []filters.SortFacet                    `tfsdk:"sort"`
	Hits              []cloudHits                            `tfsdk:"hits"`
	FacetDistribution *cloudFacetDistributionDataSourceModel `tfsdk:"facet_distribution"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProductsDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProductsDataSource{}
)

func NewCloudProductsDataSource() datasource.DataSource {
	return &cloudProductsDataSource{}
}

// Metadata returns the data source type name.
func (d *cloudProductsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_products"
}

// Schema defines the schema for the data source.
func (d *cloudProductsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Datasource to retrieve a list of cloud node products by filters.",
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [cspName, cspRegion, cspCity, location, bandwidth, provider]",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the filter among **cspName**, **cspRegion**, **cspCity**, **location**, **bandwidth**, **provider**",
							Optional:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Comparison operators. You can use the following list: **=**, **!=**, **>**, **>=**, **<**, **<=**, **IN**, **TO**. **IN** will return any products which have the values you passed when **TO** will return any value contained between the two (and only two) values you passed.",
							Optional:            true,
						},
						"values": schema.ListAttribute{
							MarkdownDescription: "Values of the filter",
							ElementType:         types.StringType,
							Optional:            true,
						},
					},
				},
			},
			"sort": schema.ListNestedAttribute{
				MarkdownDescription: `List of sort: [cspName, cspRegion, cspCity, location, bandwidth, provider,
priceNrc, priceMrc]`,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the key used for sorting. You can set multiple name among **cspName**, **cspRegion**, **cspCity**, **location**, **bandwidth**, **provider**, **priceNrc**, **priceMrc**",
							Optional:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "You can sort list ascending using **asc** or descending using **desc**. The order of the values matters as the first entry will be prioritized",
							Optional:            true,
						},
					},
				},
			},
			"hits": schema.ListNestedAttribute{
				MarkdownDescription: `The **hits** attribute contains the list of cloud products returned by the Meilisearch query.
Each hit represents a cloud product that matches the specified search criteria.`,
				Computed: true,
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
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
within the cloud products returned by the Meilisearch query. This attribute allows you to analyze the frequency of
different categories or attributes in the search results.`,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bandwidth":  filters.Int64MapAttr,
					"csp_city":   filters.Int64MapAttr,
					"csp_name":   filters.Int64MapAttr,
					"csp_region": filters.Int64MapAttr,
					"location":   filters.Int64MapAttr,
					"provider":   filters.Int64MapAttr,
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *cloudProductsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(models.Clients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *catalog.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = clients.CatalogClient
}

// Read refreshes the Terraform state with the latest data.
func (d *cloudProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data cloudsProductsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filtersStrings, err := filters.GetFiltersString(data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}
	sortStrings := filters.GetSortString(data.Sort)

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filtersStrings,
		Sort:   sortStrings,
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

	state := cloudsProductsDataSourceModel{
		Filters: data.Filters,
		Sort:    data.Sort,
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
