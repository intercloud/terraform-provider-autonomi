package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/intercloud/terraform-provider-autonomi/internal/data_sources/filters"
	"github.com/meilisearch/meilisearch-go"
)

type virtualAccessProductsDataSource struct {
	client *meilisearch.Client
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &virtualAccessProductsDataSource{}
	_ datasource.DataSourceWithConfigure = &virtualAccessProductsDataSource{}
)

func NewVirtualAccessProductsDataSource() datasource.DataSource {
	return &virtualAccessProductsDataSource{}
}

// Metadata returns the data source type name.
func (d *virtualAccessProductsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_access_products"
}

// Schema defines the schema for the data source.
func (d *virtualAccessProductsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Datasource to retrieve a list of virtual access node products by filters.",
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [location, bandwidth, provider]",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the filter among **location**, **bandwidth**, **provider**",
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
				MarkdownDescription: "List of sort: [location, bandwidth, priceNrc, priceMrc]",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the key used for sorting. You can set multiple name among **location**, **bandwidth**, **priceNrc**, **priceMrc**",
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
				MarkdownDescription: `The **hits** attribute contains the list of access products returned by the Meilisearch query.
Each hit represents an access product that matches the specified search criteria.`,
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
						"type":      schema.StringAttribute{Computed: true},
					},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
within the virtual access products returned by the Meilisearch query. This attribute allows you to analyze the frequency of
different categories or attributes in the search results.`,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bandwidth": filters.Int64MapAttr,
					"location":  filters.Int64MapAttr,
					"provider":  filters.Int64MapAttr,
					"type":      filters.Int64MapAttr,
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *virtualAccessProductsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *virtualAccessProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accessProductsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the filter string dynamically
	filterStrings := []string{
		fmt.Sprintf("%s %s \"%s\"", "type", "=", models.VIRTUAL),
	}
	filtersToAdd, err := filters.GetFiltersString(data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}
	filterStrings = append(filterStrings, filtersToAdd...)
	sortStrings := filters.GetSortString(data.Sort)

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filterStrings,
		Sort:   sortStrings,
		Facets: []string{
			"location",
			"bandwidth",
			"provider",
			"type",
		},
	}

	respProducts, err := d.client.Index("accessproduct").Search("", searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Access Products",
			err.Error(),
		)
		return
	}

	accessProducts := models.AccessProducts{}
	productsJSON, err := json.Marshal(respProducts) // Marshal the hits to JSON
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Access Products",
			err.Error(),
		)
		return
	}

	err = json.Unmarshal(productsJSON, &accessProducts) // Unmarshal JSON into the CloudProduct slice
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Access Products",
			err.Error(),
		)
		return
	}

	state := accessProductsDataSourceModel{
		Filters: data.Filters,
		Sort:    data.Sort,
	}

	// Map response body to model
	for _, ap := range accessProducts.Hits {
		accessProductState := accessHits{
			ID:        types.Int64Value(int64(ap.ID)),
			Provider:  types.StringValue(ap.Provider),
			Duration:  types.Int64Value(int64(ap.Duration)),
			Location:  types.StringValue(ap.Location),
			Bandwidth: types.Int64Value(int64(ap.Bandwidth)),
			Date:      types.StringValue(ap.Date),
			PriceNRC:  types.Float64Value(float64(ap.PriceNRC)),
			PriceMRC:  types.Float64Value(float64(ap.PriceMRC)),
			CostNRC:   types.Float64Value(float64(ap.CostNRC)),
			CostMRC:   types.Float64Value(float64(ap.CostMRC)),
			SKU:       types.StringValue(ap.SKU),
			Type:      types.StringValue(ap.Type),
		}
		state.Hits = append(state.Hits, accessProductState)
	}

	// Set the bandwidth map in the state
	state.FacetDistribution = &accessFacetDistributionDataSourceModel{
		Bandwidth: accessProducts.FacetDistribution.Bandwidth,
		Location:  accessProducts.FacetDistribution.Location,
		Provider:  accessProducts.FacetDistribution.Provider,
		Type:      accessProducts.FacetDistribution.Type,
	}
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
