package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/intercloud/terraform-provider-autonomi/external/products/models"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/meilisearch/meilisearch-go"
)

type cloudProductDataSource struct {
	client *meilisearch.Client
}

type cloudsProductDataSourceModel struct {
	Cheapest          *bool                                  `tfsdk:"cheapest"`
	Filters           []filter                               `tfsdk:"filters"`
	Hit               *cloudHits                             `tfsdk:"hit"`
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
	resp.TypeName = req.ProviderTypeName + "_cloud_product"
}

// Schema defines the schema for the data source.
func (d *cloudProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cheapest": schema.BoolAttribute{
				MarkdownDescription: "To ensure only one hit is returned we advise to set at true",
				Optional:            true,
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters: [cspName, cspRegion, cspCity, location, bandwidth, provider]",
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
				Each hit represents a cloud product that matches the specified search criteria.
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
					"csp_name":  schema.StringAttribute{Computed: true},
				},
			},
			"facet_distribution": schema.SingleNestedAttribute{
				MarkdownDescription: `The **facet_distribution** attribute provides an overview of the distribution of various facets
within the cloud products returned by the Meilisearch query. This attribute allows you to analyze the frequency of
different categories or attributes in the search results.`,
				Computed: true,
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
func (d *cloudProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data cloudsProductDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filtersStrings, err := getFiltersString(data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}

	// Define the search request
	searchRequest := &meilisearch.SearchRequest{
		Filter: filtersStrings,
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

	if len(cloudProducts.Hits) == 0 {
		resp.Diagnostics.AddError("Not hit found", "")
		return
	}

	// If Meiliesearch return more than one hit, check if `cheapest` filter has been set.
	// If not, an error is returned, otherwise a sort will be done to order the list by price mrc. The first entry will be returned
	if len(cloudProducts.Hits) > 1 {
		if data.Cheapest == nil || !*data.Cheapest {
			resp.Diagnostics.AddError("Request got more than one hit, please set cheapest=true", "")
			return
		}
		// sort slice by price mrc if cheapest=true is set
		sort.Slice(cloudProducts.Hits, func(i, j int) bool {
			return cloudProducts.Hits[i].PriceMRC < cloudProducts.Hits[j].PriceMRC
		})
	}

	state := cloudsProductDataSourceModel{
		Cheapest: data.Cheapest,
		Filters:  data.Filters,
	}

	cp := cloudProducts.Hits[0]
	state.Hit = &cloudHits{
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
