package datasources

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	autonomisdk "github.com/intercloud/autonomi-sdk"
	autonomisdkmodel "github.com/intercloud/autonomi-sdk/models"
	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/intercloud/terraform-provider-autonomi/internal/data_sources/filters"
)

type physicalPortDataSource struct {
	client *autonomisdk.Client
}

type physicalPortDataSourceModel struct {
	Recent  *bool             `tfsdk:"most_recent"`
	Filters []filters.Filter  `tfsdk:"filters"`
	Port    *physicalPortHits `tfsdk:"port"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &physicalPortDataSource{}
	_ datasource.DataSourceWithConfigure = &physicalPortDataSource{}
)

func NewPhysicalPortDataSource() datasource.DataSource {
	return &physicalPortDataSource{}
}

// Metadata returns the data source type name.
func (d *physicalPortDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_physical_port"
}

// Schema defines the schema for the data source.
func (d *physicalPortDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"most_recent": schema.BoolAttribute{
				MarkdownDescription: "To ensure only one hit is returned we advise to set at true",
				Optional:            true,
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: `List of filters: [id, name, location, bandwidth, priceMrc, priceNrc].
Operators avaiable are [=, IN]`,
				Optional: true,
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
			"port": schema.SingleNestedAttribute{
				MarkdownDescription: `The **ports** attribute contains the list of physical ports available on the accountId.
Each port represents a physical-port that matches the specified search criteria.`,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					// ID attribute for the physical port
					"id": schema.StringAttribute{
						MarkdownDescription: "The **ID** of the physical port.",
						Computed:            true,
					},
					// Name attribute for the physical port
					"name": schema.StringAttribute{
						MarkdownDescription: "The **name** of the physical port.",
						Computed:            true,
					},
					// Account ID attribute
					"account_id": schema.StringAttribute{
						MarkdownDescription: "The **account ID** associated with the physical port.",
						Computed:            true,
					},
					// Product attribute that nests a physicalPortProduct structure
					"product": schema.SingleNestedAttribute{
						MarkdownDescription: "The **product** attribute represents details about the physical port product.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"provider": schema.StringAttribute{
								MarkdownDescription: "The **provider** of the product.",
								Computed:            true,
							},
							"duration": schema.Int64Attribute{
								MarkdownDescription: "The **duration** associated with the product.",
								Computed:            true,
							},
							"location": schema.StringAttribute{
								MarkdownDescription: "The **location** where the product is available.",
								Computed:            true,
							},
							"bandwidth": schema.Int64Attribute{
								MarkdownDescription: "The **bandwidth** offered by the product.",
								Computed:            true,
							},
							"price_nrc": schema.Float64Attribute{
								MarkdownDescription: "The **non-recurring cost** (NRC) of the product.",
								Computed:            true,
							},
							"price_mrc": schema.Float64Attribute{
								MarkdownDescription: "The **monthly recurring cost** (MRC) of the product.",
								Computed:            true,
							},
							"cost_nrc": schema.Float64Attribute{
								MarkdownDescription: "The internal **non-recurring cost** (NRC) for the provider.",
								Computed:            true,
							},
							"cost_mrc": schema.Float64Attribute{
								MarkdownDescription: "The internal **monthly recurring cost** (MRC) for the provider.",
								Computed:            true,
							},
							"sku": schema.StringAttribute{
								MarkdownDescription: "The **SKU** of the product.",
								Computed:            true,
							},
						},
					},
					// Available Bandwidth attribute
					"available_bandwidth": schema.Int64Attribute{
						MarkdownDescription: "The **available bandwidth** on the physical port.",
						Computed:            true,
					},
					// Administrative State attribute
					"administrative_state": schema.StringAttribute{
						MarkdownDescription: "The **administrative state** of the physical port.",
						Computed:            true,
					},
					// Used VLANs attribute as a list of Int64
					"used_vlans": schema.ListAttribute{
						MarkdownDescription: "A list of **used VLANs** on the physical port.",
						Computed:            true,
						ElementType:         types.Int64Type,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *physicalPortDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.AutonomiClient
}

// Read refreshes the Terraform state with the latest data.
func (d *physicalPortDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data physicalPortDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	respPhysicalPorts, err := d.client.ListPort(autonomisdk.WithAdministrativeState(autonomisdkmodel.AdministrativeStateCreated))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Autonomi PhysicalPort Products",
			err.Error(),
		)
		return
	}

	// apply filters on physical-ports list returned from autonomi
	filteredPhysicalPorts, err := filters.Apply(*respPhysicalPorts, data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}

	if len(filteredPhysicalPorts) == 0 {
		resp.Diagnostics.AddError("Not hit found", "")
		return
	}

	// If Meiliesearch return more than one hit, check if `cheapest` filter has been set.
	// If not, an error is returned, otherwise a sort will be done to order the list by price mrc. The first entry will be returned
	if len(filteredPhysicalPorts) > 1 {
		if data.Recent == nil || !*data.Recent {
			resp.Diagnostics.AddError("Request got more than one hit, please set most_recent=true", "")
			return
		}
		// sort slice by price mrc if cheapest=true is set
		sort.Slice(filteredPhysicalPorts, func(i, j int) bool {
			return filteredPhysicalPorts[i].CreatedAt.After(filteredPhysicalPorts[j].CreatedAt)
		})
	}

	physicalPort := filteredPhysicalPorts[0]

	// Map response body to model
	// Create a map to hold attributes of each physicalPort
	physicalPortTF := physicalPortHits{
		ID:                 types.StringValue(physicalPort.ID.String()),
		Name:               types.StringValue(physicalPort.Name),
		AccountID:          types.StringValue(physicalPort.AccountID),
		AvailableBandwidth: types.Int64Value(int64(physicalPort.Product.Bandwidth)),
		Product: physicalPortProduct{
			Provider:  types.StringValue(physicalPort.Product.Provider.String()),
			Duration:  types.Int64Value(int64(physicalPort.Product.Duration)),
			Location:  types.StringValue(physicalPort.Product.Location),
			Bandwidth: types.Int64Value(int64(physicalPort.Product.Bandwidth)),
			PriceNRC:  types.Int64Value(int64(physicalPort.Product.PriceNRC)),
			PriceMRC:  types.Int64Value(int64(physicalPort.Product.PriceMRC)),
			CostNRC:   types.Int64Value(int64(physicalPort.Product.CostNRC)),
			CostMRC:   types.Int64Value(int64(physicalPort.Product.CostMRC)),
			SKU:       types.StringValue(physicalPort.Product.SKU),
		},
		State:     types.StringValue(physicalPort.State),
		UsedVLANs: convertToTerraformList(physicalPort.UsedVLANs),
	}

	state := physicalPortDataSourceModel{
		Port: &physicalPortTF,
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
