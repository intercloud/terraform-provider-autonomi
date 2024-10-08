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

type transportProductDataSource struct {
	client *meilisearch.Client
}

type transportProductDataSourceModel struct {
	Cheapest          *bool                                      `tfsdk:"cheapest"`
	Filters           []filters.Filter                           `tfsdk:"filters"`
	Hit               *transportHits                             `tfsdk:"hit"`
	FacetDistribution *transportFacetDistributionDataSourceModel `tfsdk:"facet_distribution"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &transportProductDataSource{}
	_ datasource.DataSourceWithConfigure = &transportProductDataSource{}
)

func NewTransportProductDataSource() datasource.DataSource {
	return &transportProductDataSource{}
}

// Metadata returns the data source type name.
func (d *transportProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transport_product"
}

// Schema defines the schema for the data source.
func (d *transportProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Datasource to retrieve a single transport product by filters.
If zero, or more than one, product are retrieved with the filters, this datasource raises an error.`,
		Attributes: map[string]schema.Attribute{
			"cheapest": schema.BoolAttribute{
				MarkdownDescription: "To ensure only one hit is returned we advise to set at true",
				Optional:            true,
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [location, locationTo, bandwidth, provider]",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the filter among **location**, **locationTo**, **bandwidth**, **provider**",
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
			"hit": schema.SingleNestedAttribute{
				MarkdownDescription: `The **hit** attribute contains the transport product returned by the Meilisearch query.
Each hit represents a transport product that matches the specified search criteria.`,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id":                   schema.Int64Attribute{Computed: true},
					"provider":             schema.StringAttribute{Computed: true},
					"duration":             schema.Int64Attribute{Computed: true},
					"location":             schema.StringAttribute{Computed: true},
					"location_underlay":    schema.StringAttribute{Computed: true},
					"location_to":          schema.StringAttribute{Computed: true},
					"location_to_underlay": schema.StringAttribute{Computed: true},
					"bandwidth":            schema.Int64Attribute{Computed: true},
					"date":                 schema.StringAttribute{Computed: true},
					"price_nrc":            schema.Int64Attribute{Computed: true},
					"price_mrc":            schema.Int64Attribute{Computed: true},
					"cost_nrc":             schema.Int64Attribute{Computed: true},
					"cost_mrc":             schema.Int64Attribute{Computed: true},
					"sku":                  schema.StringAttribute{Computed: true},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
within the transport products returned by the Meilisearch query. This attribute allows you to analyze the frequency
of different categories or attributes in the search results.`,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"provider":    filters.Int64MapAttr,
					"bandwidth":   filters.Int64MapAttr,
					"location":    filters.Int64MapAttr,
					"location_to": filters.Int64MapAttr,
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *transportProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *transportProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data transportProductDataSourceModel

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
			"locationTo",
			"bandwidth",
			"provider",
		},
	}

	respProducts, err := d.client.Index("transportproduct").Search("", searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Transport Products",
			err.Error(),
		)
		return
	}

	transportProducts := models.TransportProducts{}
	productsJSON, err := json.Marshal(respProducts) // Marshal the hits to JSON
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Transport Products",
			err.Error(),
		)
		return
	}

	err = json.Unmarshal(productsJSON, &transportProducts) // Unmarshal JSON into the TransportProduct slice
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Transport Products",
			err.Error(),
		)
		return
	}

	if len(transportProducts.Hits) == 0 {
		resp.Diagnostics.AddError("Not hit found", "")
		return
	}

	// If Meiliesearch return more than one hit, check if `cheapest` filter has been set.
	// If not, an error is returned, otherwise a sort will be done to order the list by price mrc. The first entry will be returned
	if len(transportProducts.Hits) > 1 {
		if data.Cheapest == nil || !*data.Cheapest {
			resp.Diagnostics.AddError("Request got more than one hit, please set cheapest=true", "")
			return
		}
		// sort slice by price mrc if cheapest=true is set
		sort.Slice(transportProducts.Hits, func(i, j int) bool {
			return transportProducts.Hits[i].PriceMRC < transportProducts.Hits[j].PriceMRC
		})
	}

	state := transportProductDataSourceModel{
		Cheapest: data.Cheapest,
		Filters:  data.Filters,
	}

	tp := transportProducts.Hits[0]
	state.Hit = &transportHits{
		ID:                 types.Int64Value(int64(tp.ID)),
		Provider:           types.StringValue(tp.Provider),
		Location:           types.StringValue(tp.Location),
		LocationUnderlay:   types.StringValue(tp.LocationUnderlay),
		Bandwidth:          types.Int64Value(int64(tp.Bandwidth)),
		Date:               types.StringValue(tp.Date),
		PriceNRC:           types.Float64Value(float64(tp.PriceNRC)),
		PriceMRC:           types.Float64Value(float64(tp.PriceMRC)),
		CostNRC:            types.Float64Value(float64(tp.CostNRC)),
		CostMRC:            types.Float64Value(float64(tp.CostMRC)),
		SKU:                types.StringValue(tp.SKU),
		LocationTo:         types.StringValue(tp.LocationTo),
		LocationToUnderlay: types.StringValue(tp.LocationToUnderlay),
	}

	// Set the bandwidth map in the state
	state.FacetDistribution = &transportFacetDistributionDataSourceModel{
		Bandwidth:  transportProducts.FacetDistribution.Bandwidth,
		Location:   transportProducts.FacetDistribution.Location,
		LocationTo: transportProducts.FacetDistribution.LocationTo,
		Provider:   transportProducts.FacetDistribution.Provider,
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
