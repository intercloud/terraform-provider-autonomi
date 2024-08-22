package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/meilisearch/meilisearch-go"
)

type accessProductDataSource struct {
	client *meilisearch.Client
}

type accessHits struct {
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
	Type      types.String  `tfsdk:"type"`
}

type accessFacetDistributionDataSourceModel struct {
	Bandwidth map[string]int `tfsdk:"bandwidth"`
	Location  map[string]int `tfsdk:"location"`
	Provider  map[string]int `tfsdk:"provider"`
	Type      map[string]int `tfsdk:"type"`
}

type accessProductDataSourceModel struct {
	Filters           []filter                                `tfsdk:"filters"`
	Hits              []accessHits                            `tfsdk:"hits"`
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
	resp.TypeName = req.ProviderTypeName + "_access_products"
}

// Schema defines the schema for the data source.
func (d *accessProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [location, bandwidth]",
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
			"hits": schema.ListNestedAttribute{
				MarkdownDescription: "The **hits** attribute contains the list of cloud products returned by the Meilisearch query. Each hit represents an access product that matches the specified search criteria.",
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
						"type":      schema.StringAttribute{Computed: true},
					},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: "The **facet_distribution** attribute provides an overview of the distribution of various facets within the access products returned by the Meilisearch query. This attribute allows you to analyze the frequency of different categories or attributes in the search results.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"bandwidth": int64MapAttr,
					"location":  int64MapAttr,
					"provider":  int64MapAttr,
					"type":      int64MapAttr,
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
	filtersToAdd, err := getFiltersString(data.Filters)
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

	err = json.Unmarshal(productsJSON, &accessProducts) // Unmarshal JSON into the CloudProduct slice
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi Access Products",
			err.Error(),
		)
		return
	}

	state := accessProductDataSourceModel{
		Filters: data.Filters,
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
