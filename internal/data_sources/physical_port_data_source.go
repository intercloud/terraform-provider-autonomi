package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	autonomisdk "github.com/intercloud/autonomi-sdk"
	autonomisdkmodel "github.com/intercloud/autonomi-sdk/models"
	"github.com/intercloud/terraform-provider-autonomi/external/products/models"
	"github.com/intercloud/terraform-provider-autonomi/internal/data_sources/filters"
)

type physicalPortsDataSource struct {
	client *autonomisdk.Client
}

type physicalPortProduct struct {
	Provider  types.String `tfsdk:"provider"`
	Duration  types.Int64  `tfsdk:"duration"`
	Location  types.String `tfsdk:"location"`
	Bandwidth types.Int64  `tfsdk:"bandwidth"`
	PriceNRC  types.Int64  `tfsdk:"price_nrc"`
	PriceMRC  types.Int64  `tfsdk:"price_mrc"`
	CostNRC   types.Int64  `tfsdk:"cost_nrc"`
	CostMRC   types.Int64  `tfsdk:"cost_mrc"`
	SKU       types.String `tfsdk:"sku"`
}

type physicalPortHits struct {
	ID                 types.String        `tfsdk:"id"`
	Name               types.String        `tfsdk:"name"`
	AccountID          types.String        `tfsdk:"account_id"`
	Product            physicalPortProduct `tfsdk:"product"`
	AvailableBandwidth types.Int64         `tfsdk:"available_bandwidth"`
	State              types.String        `tfsdk:"administrative_state"`
	UsedVLANs          []types.Int64       `tfsdk:"used_vlans"`
}

type physicalPortDataSourceModel struct {
	Filters []filters.Filter   `tfsdk:"filters"`
	Ports   []physicalPortHits `tfsdk:"ports"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &physicalPortsDataSource{}
	_ datasource.DataSourceWithConfigure = &physicalPortsDataSource{}
)

func NewPhysicalPortsDataSource() datasource.DataSource {
	return &physicalPortsDataSource{}
}

// Metadata returns the data source type name.
func (d *physicalPortsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_physical_ports"
}

// Schema defines the schema for the data source.
func (d *physicalPortsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: `The **ports** attribute contains the list of physical ports available on the accountId.
Each port represents a physical-port that matches the specified search criteria.`,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
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
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *physicalPortsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *physicalPortsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
	filterPhysicalPort, err := filters.Apply(*respPhysicalPorts, data.Filters)
	if err != nil {
		resp.Diagnostics.AddError("error getting filters", err.Error())
	}

	// Create a slice to hold the state
	physicalPortsTF := []physicalPortHits{}

	// Map response body to model
	for _, physicalPort := range filterPhysicalPort {
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
		physicalPortsTF = append(physicalPortsTF, physicalPortTF)
	}

	state := physicalPortDataSourceModel{
		Ports: physicalPortsTF,
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Helper function to convert a slice of int64 to a list of Int64 values for Terraform
func convertToTerraformList(vlans []int64) []types.Int64 {
	result := make([]types.Int64, len(vlans))
	for i, vlan := range vlans {
		result[i] = types.Int64Value(vlan)
	}
	return result
}
