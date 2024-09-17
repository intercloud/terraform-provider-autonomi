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

type physicalPortProductsDataSource struct {
	client *meilisearch.Client
}

type physicalPortProductHits struct {
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
}

type physicalPortProductFacetDistributionDataSourceModel struct {
	Bandwidth map[string]int `tfsdk:"bandwidth"`
	Location  map[string]int `tfsdk:"location"`
	Provider  map[string]int `tfsdk:"provider"`
	Duration  map[string]int `tfsdk:"duration"`
}

type physicalPortsProductsDataSourceModel struct {
	Filters           []filters.Filter                                     `tfsdk:"filters"`
	Sort              []filters.SortFacet                                  `tfsdk:"sort"`
	Hits              []physicalPortProductHits                            `tfsdk:"hits"`
	FacetDistribution *physicalPortProductFacetDistributionDataSourceModel `tfsdk:"facet_distribution"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &physicalPortProductsDataSource{}
	_ datasource.DataSourceWithConfigure = &physicalPortProductsDataSource{}
)

func NewPhysicalPortProductsDataSource() datasource.DataSource {
	return &physicalPortProductsDataSource{}
}

// Metadata returns the data source type name.
func (d *physicalPortProductsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_physical_port_products"
}

// Schema defines the schema for the data source.
func (d *physicalPortProductsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Datasource to retrieve a list of physical port's products by filters.",
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [location, bandwidth, provider, duration]",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"operator": schema.StringAttribute{
							Optional: true,
						},
						"values": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
			"sort": schema.ListNestedAttribute{
				MarkdownDescription: `List of sort: [location, bandwidth, provider, duration,
priceNrc, priceMrc, costNrc, costMrc]`,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"value": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"hits": schema.ListNestedAttribute{
				MarkdownDescription: `The **hits** attribute contains the list of physical port's products returned by the Meilisearch query.
Each hit represents a physical port product that matches the specified search criteria.`,
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
					},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
within the physical port's products returned by the Meilisearch query. This attribute allows you to analyze the frequency of
different categories or attributes in the search results.`,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bandwidth": filters.Int64MapAttr,
					"location":  filters.Int64MapAttr,
					"provider":  filters.Int64MapAttr,
					"duration":  filters.Int64MapAttr,
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *physicalPortProductsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *physicalPortProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data physicalPortsProductsDataSourceModel

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
			"location",
			"bandwidth",
			"provider",
			"duration",
		},
	}

	respProducts, err := d.client.Index("portproduct").Search("", searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Autonomi physical port's products",
			err.Error(),
		)
		return
	}

	physicalPortProducts := models.PhysicalPortProducts{}
	productsJSON, err := json.Marshal(respProducts) // Marshal the hits to JSON
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Autonomi physical port's products",
			err.Error(),
		)
		return
	}

	err = json.Unmarshal(productsJSON, &physicalPortProducts) // Unmarshal JSON into the PhysicalPortProduct slice
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Autonomi physical port's products",
			err.Error(),
		)
		return
	}

	state := physicalPortsProductsDataSourceModel{
		Filters: data.Filters,
		Sort:    data.Sort,
	}

	// Map response body to model
	for _, cp := range physicalPortProducts.Hits {
		physicalPortProductState := physicalPortProductHits{
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
		}
		state.Hits = append(state.Hits, physicalPortProductState)
	}

	// Set the bandwidth map in the state
	state.FacetDistribution = &physicalPortProductFacetDistributionDataSourceModel{
		Bandwidth: physicalPortProducts.FacetDistribution.Bandwidth,
		Location:  physicalPortProducts.FacetDistribution.Location,
		Provider:  physicalPortProducts.FacetDistribution.Provider,
		Duration:  physicalPortProducts.FacetDistribution.Duration,
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
