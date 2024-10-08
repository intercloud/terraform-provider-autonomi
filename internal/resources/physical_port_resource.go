package autonomiresource

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	autonomisdk "github.com/intercloud/autonomi-sdk"
	"github.com/intercloud/autonomi-sdk/models"
)

const (
	AUTONOMI_FRONT_URL = "https://autonomi-platform.com/#"
)

// physicalPortResource is the resource implementation.
type physicalPortResource struct {
	client *autonomisdk.Client
}

type physicalPortResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	AccountID          types.String `tfsdk:"account_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	Name               types.String `tfsdk:"name"`
	State              types.String `tfsdk:"administrative_state"`
	Product            product      `tfsdk:"product"`
	AvailableBandwidth types.Int64  `tfsdk:"available_bandwidth"`
	UsedVLANs          types.List   `tfsdk:"used_vlans"`
	LOAAccessURL       types.String `tfsdk:"loa_access_url"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &physicalPortResource{}
	_ resource.ResourceWithConfigure = &physicalPortResource{}
)

// NewPhysicalPortResource is a helper function to simplify the provider implementation.
func NewPhysicalPortResource() resource.Resource {
	return &physicalPortResource{}
}

// Configure adds the provider configured client to the resource.
func (r *physicalPortResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*autonomisdk.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *autonomi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *physicalPortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_physical_port"
}

// Schema defines the schema for the resource.
func (r *physicalPortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a physical port resource.
Physical port resource allows you to create and delete Autonomi physical ports.
Autonomi physical port are shared Autonomi connection instance.
They allow you to connect InterCloud back bone through access node.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the physical port, set after creation",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the physical port",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Update date of the physical port",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "Account ID of the physical port, is determined by the personal access token",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the physical port",
				Required:            true,
			},
			"administrative_state": schema.StringAttribute{
				MarkdownDescription: `Administrative state of the physical port [created, deleted]`,
				Computed:            true,
			},
			"product": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"sku": schema.StringAttribute{
						MarkdownDescription: "ID of the product",
						Required:            true,
					},
				},
			},
			"available_bandwidth": schema.Int64Attribute{
				MarkdownDescription: `Available bandwidth on the physical port`,
				Computed:            true,
			},
			"used_vlans": schema.ListAttribute{
				MarkdownDescription: `Vlan already attributed on the physical port`,
				Computed:            true,
				ElementType:         types.NumberType,
			},
			"loa_access_url": schema.StringAttribute{
				MarkdownDescription: `URL to the physical port page where the LOA is downloadable`,
				Computed:            true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *physicalPortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan physicalPortResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.CreatePhysicalPort{
		Name: plan.Name.ValueString(),
		Product: models.AddProduct{
			SKU: plan.Product.SKU.ValueString(),
		},
	}

	// Create new physical port
	physicalPort, err := r.client.CreatePhysicalPort(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating physical port",
			"Could not create physical port, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(physicalPort.ID.String())
	plan.AccountID = types.StringValue(physicalPort.AccountID)
	plan.State = types.StringValue(physicalPort.State.String())
	plan.CreatedAt = types.StringValue(physicalPort.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(physicalPort.UpdatedAt.String())
	plan.AvailableBandwidth = types.Int64Value(int64(physicalPort.AvailableBandwidth))
	plan.UsedVLANs = types.ListValueMust(types.NumberType, convertInt64ArrayToNumberValues(physicalPort.UsedVLANs))
	plan.LOAAccessURL = types.StringValue(physicalPort.LOAAccessURL)
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *physicalPortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state physicalPortResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed physical port value from Autonomi
	physicalPort, err := r.client.GetPhysicalPort(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Autonomi physical port",
			"Could not read Autonomi physical port ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(physicalPort.ID.String())
	state.CreatedAt = types.StringValue(physicalPort.CreatedAt.String())
	state.UpdatedAt = types.StringValue(physicalPort.UpdatedAt.String())
	state.Name = types.StringValue(physicalPort.Name)
	state.State = types.StringValue(physicalPort.State.String())
	state.Product = product{
		SKU: types.StringValue(physicalPort.Product.SKU),
	}
	state.AccountID = types.StringValue(physicalPort.AccountID)
	state.AvailableBandwidth = types.Int64Value(int64(physicalPort.AvailableBandwidth))
	state.UsedVLANs = types.ListValueMust(types.NumberType, convertInt64ArrayToNumberValues(physicalPort.UsedVLANs))
	state.LOAAccessURL = types.StringValue(fmt.Sprintf("%s/ports/details/port/%s", AUTONOMI_FRONT_URL, physicalPort.ID))
	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *physicalPortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *physicalPortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state physicalPortResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing physical port
	err := r.client.DeletePhysicalPort(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting physical port",
			"Could not delete physical port, unexpected error: "+err.Error(),
		)
		return
	}
}

// Helper function to convert []int64 to []attr.Value for use in ListValueMust
func convertInt64ArrayToNumberValues(input []int64) []attr.Value {
	result := make([]attr.Value, len(input))
	for i, v := range input {
		// We convert each int64 into a Terraform-compatible number value
		result[i] = types.NumberValue(new(big.Float).SetInt64(v))
	}
	return result
}
