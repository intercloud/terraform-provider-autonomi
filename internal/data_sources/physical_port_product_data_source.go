package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/intercloud/terraform-provider-autonomi/internal/data_sources/filters"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/meilisearch/meilisearch-go"
)

type physicalPortProductDataSource struct {
	client *meilisearch.Client
}

type physicalPortsProductDataSourceModel struct {
	Cheapest          *bool                                                `tfsdk:"cheapest"`
	Filters           []filters.Filter                                     `tfsdk:"filters"`
	Hit               *physicalPortProductHits                             `tfsdk:"hit"`
	FacetDistribution *physicalPortProductFacetDistributionDataSourceModel `tfsdk:"facet_distribution"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &physicalPortProductDataSource{}
	_ datasource.DataSourceWithConfigure = &physicalPortProductDataSource{}
)

func NewPhysicalPortProductDataSource() datasource.DataSource {
	return &physicalPortProductDataSource{}
}

// Metadata returns the data source type name.
func (d *physicalPortProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_physical_port_product"
}

// Schema defines the schema for the data source.
func (d *physicalPortProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Datasource to retrieve a single physical port's product by filters.
If zero, or more than one, product are retrieved with the filters, this datasource raises an error.`,
		Attributes: map[string]schema.Attribute{
			"cheapest": schema.BoolAttribute{
				MarkdownDescription: "To ensure only one hit is returned we advise to set at true",
				Optional:            true,
			},
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
			"hit": schema.SingleNestedAttribute{
				MarkdownDescription: `The **hit** attribute contains the physical port products returned by the Meilisearch query.
Each hit represents a physical port's product that matches the specified search criteria.
If no hit is returned, an error will be returned`,
				Computed: true,
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
func (d *physicalPortProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *physicalPortProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data physicalPortsProductDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filtersStrings, err := filters.GetFiltersString(data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filtersStrings,
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
			"Unable to read autonomi physical port's products",
			err.Error(),
		)
		return
	}

	physicalPortProducts := models.PhysicalPortProducts{}
	productsJSON, err := json.Marshal(respProducts) // Marshal the hits to JSON
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read autonomi physical port's products",
			err.Error(),
		)
		return
	}

	err = json.Unmarshal(productsJSON, &physicalPortProducts) // Unmarshal JSON into the PhysicalPortProduct slice
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read autonomi physical port's products",
			err.Error(),
		)
		return
	}

	if len(physicalPortProducts.Hits) == 0 {
		resp.Diagnostics.AddError("Not hit found", "")
		return
	}

	// If Meiliesearch return more than one hit, check if `cheapest` filter has been set.
	// If not, an error is returned, otherwise a sort will be done to order the list by price mrc. The first entry will be returned
	if len(physicalPortProducts.Hits) > 1 {
		if data.Cheapest == nil || !*data.Cheapest {
			resp.Diagnostics.AddError("Request got more than one hit, please set cheapest=true", "")
			return
		}
		// sort slice by price mrc if cheapest=true is set
		sort.Slice(physicalPortProducts.Hits, func(i, j int) bool {
			return physicalPortProducts.Hits[i].PriceMRC < physicalPortProducts.Hits[j].PriceMRC
		})
	}

	state := physicalPortsProductDataSourceModel{
		Cheapest: data.Cheapest,
		Filters:  data.Filters,
	}

	cp := physicalPortProducts.Hits[0]
	state.Hit = &physicalPortProductHits{
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
