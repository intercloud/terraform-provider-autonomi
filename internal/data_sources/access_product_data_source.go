package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/intercloud/terraform-provider-autonomi/internal/data_sources/filters"
	"github.com/meilisearch/meilisearch-go"
)

type accessProductDataSource struct {
	client *meilisearch.Client
}

type accessProductDataSourceModel struct {
	Cheapest          *bool                                   `tfsdk:"cheapest"`
	Filters           []filters.Filter                        `tfsdk:"filters"`
	Hit               *accessHits                             `tfsdk:"hit"`
	FacetDistribution *accessFacetDistributionDataSourceModel `tfsdk:"facet_distribution"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &accessProductDataSource{}
	_ datasource.DataSourceWithConfigure = &accessProductDataSource{}
)

func NewAccessProductDataSource() datasource.DataSource {
	return &accessProductDataSource{}
}

// Metadata returns the data source type name.
func (d *accessProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_product"
}

// Schema defines the schema for the data source.
func (d *accessProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Datasource to retrieve a single access node product by filters.
If zero, or more than one, product are retrieved with the filters, this datasource raises an error.`,
		Attributes: map[string]schema.Attribute{
			"cheapest": schema.BoolAttribute{
				MarkdownDescription: "To ensure only one hit is returned we advise to set at true",
				Optional:            true,
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [location, bandwidth].",
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
				MarkdownDescription: `The **hit** attribute contains the access products returned by the Meilisearch query.
Each hit represents an access product that matches the specified search criteria.
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
					"type":      schema.StringAttribute{Computed: true},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
within the access products returned by the Meilisearch query. This attribute allows you to analyze the frequency of
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
func (d *accessProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *accessProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accessProductDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the filter string dynamically
	filterStrings := []string{
		fmt.Sprintf("%s %s \"%s\"", "provider", "=", models.INTERCLOUD),
		fmt.Sprintf("%s %s \"%s\"", "type", "=", models.PHYSICAL),
	}
	filtersToAdd, err := filters.GetFiltersString(data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}
	filterStrings = append(filterStrings, filtersToAdd...)

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filterStrings,
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

	if err := json.Unmarshal(productsJSON, &accessProducts); err != nil { // Unmarshal JSON into the AccessProduct slice
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Access Products",
			err.Error(),
		)
		return
	}

	if len(accessProducts.Hits) == 0 {
		resp.Diagnostics.AddError("Not hit found", "")
		return
	}

	// If Meiliesearch return more than one hit, check if `cheapest` filter has been set.
	// If not, an error is returned, otherwise a sort will be done to order the list by price mrc. The first entry will be returned
	if len(accessProducts.Hits) > 1 {
		if data.Cheapest == nil || !*data.Cheapest {
			resp.Diagnostics.AddError("Request got more than one hit, please set cheapest=true", "")
			return
		}
		// sort slice by price mrc if cheapest=true is set
		sort.Slice(accessProducts.Hits, func(i, j int) bool {
			return accessProducts.Hits[i].PriceMRC < accessProducts.Hits[j].PriceMRC
		})
	}

	state := accessProductDataSourceModel{
		Cheapest: data.Cheapest,
		Filters:  data.Filters,
	}

	ap := accessProducts.Hits[0]
	state.Hit = &accessHits{
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
