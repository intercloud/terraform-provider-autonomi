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

type transportProductDataSource struct {
	client *meilisearch.Client
}

type transportHits struct {
	ID                 types.Int64   `tfsdk:"id"`
	Provider           types.String  `tfsdk:"provider"`
	Duration           types.Int64   `tfsdk:"duration"`
	Location           types.String  `tfsdk:"location"`
	LocationUnderlay   types.String  `tfsdk:"location_underlay"`
	Bandwidth          types.Int64   `tfsdk:"bandwidth"`
	Date               types.String  `tfsdk:"date"`
	PriceNRC           types.Float64 `tfsdk:"price_nrc"`
	PriceMRC           types.Float64 `tfsdk:"price_mrc"`
	CostNRC            types.Float64 `tfsdk:"cost_nrc"`
	CostMRC            types.Float64 `tfsdk:"cost_mrc"`
	SKU                types.String  `tfsdk:"sku"`
	LocationTo         types.String  `tfsdk:"location_to"`
	LocationToUnderlay types.String  `tfsdk:"location_to_underlay"`
}

type transportFacetDistributionDataSourceModel struct {
	Bandwidth  map[string]int `tfsdk:"bandwidth"`
	Location   map[string]int `tfsdk:"location"`
	LocationTo map[string]int `tfsdk:"location_to"`
	Provider   map[string]int `tfsdk:"provider"`
}

type transportsProductDataSourceModel struct {
	Filters           []filter                                   `tfsdk:"filters"`
	Sort              []sort                                     `tfsdk:"sort"`
	Hits              []transportHits                            `tfsdk:"hits"`
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
	resp.TypeName = req.ProviderTypeName + "_transport_products"
}

// Schema defines the schema for the data source.
func (d *transportProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [location, locationTo, bandwidth, provider]",
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
				MarkdownDescription: `List of sort: [location, locationTo, bandwidth, provider,
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
				MarkdownDescription: `The **hits** attribute contains the list of transport products returned by the Meilisearch query.
					Each hit represents a transport product that matches the specified search criteria.`,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
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
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
					within the transport products returned by the Meilisearch query. This attribute allows you to analyze the frequency
					of different categories or attributes in the search results.`,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"provider":    int64MapAttr,
					"bandwidth":   int64MapAttr,
					"location":    int64MapAttr,
					"location_to": int64MapAttr,
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
func (d *transportProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data transportsProductDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filtersStrings, err := getFiltersString(data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}
	sortStrings := getSortString(data.Sort)
	if err != nil {
		resp.Diagnostics.AddError("error getting sort", err.Error())
	}

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filtersStrings,
		Sort:   sortStrings,
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

	state := transportsProductDataSourceModel{
		Filters: data.Filters,
		Sort:    data.Sort,
	}

	// Map response body to model
	for _, cp := range transportProducts.Hits {
		transportProductState := transportHits{
			ID:                 types.Int64Value(int64(cp.ID)),
			Provider:           types.StringValue(cp.Provider),
			Location:           types.StringValue(cp.Location),
			LocationUnderlay:   types.StringValue(cp.LocationUnderlay),
			Bandwidth:          types.Int64Value(int64(cp.Bandwidth)),
			Date:               types.StringValue(cp.Date),
			PriceNRC:           types.Float64Value(float64(cp.PriceNRC)),
			PriceMRC:           types.Float64Value(float64(cp.PriceMRC)),
			CostNRC:            types.Float64Value(float64(cp.CostNRC)),
			CostMRC:            types.Float64Value(float64(cp.CostMRC)),
			SKU:                types.StringValue(cp.SKU),
			LocationTo:         types.StringValue(cp.LocationTo),
			LocationToUnderlay: types.StringValue(cp.LocationToUnderlay),
		}
		state.Hits = append(state.Hits, transportProductState)
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
