package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	autonomisdk "github.com/intercloud/autonomi-sdk"
	"github.com/intercloud/autonomi-sdk/models"
)

// transportResource is the resource implementation.
type transportResource struct {
	client *autonomisdk.Client
}

var transportVlans = map[string]attr.Type{
	"a_vlan": types.Int64Type,
	"z_vlan": types.Int64Type,
}

type transportResourceModel struct {
	ID           types.String `tfsdk:"id"`
	WorkspaceID  types.String `tfsdk:"workspace_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	DeployedAt   types.String `tfsdk:"deployed_at"`
	Name         types.String `tfsdk:"name"`
	State        types.String `tfsdk:"administrative_state"`
	Product      product      `tfsdk:"product"`
	Vlans        types.Object `tfsdk:"vlans"`
	ConnectionID types.String `tfsdk:"connection_id"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &transportResource{}
	_ resource.ResourceWithConfigure = &transportResource{}
)

// NewTransportResource is a helper function to simplify the provider implementation.
func NewTransportResource() resource.Resource {
	return &transportResource{}
}

// Configure adds the provider configured client to the resource.
func (r *transportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *transportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transport"
}

// Schema defines the schema for the resource.
func (r *transportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the transport, set after creation",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the transport",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Update date of the transport",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployed_at": schema.StringAttribute{
				MarkdownDescription: "Deployment date of the transport",
				Computed:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace to which the transport belongs.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the transport",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"administrative_state": schema.StringAttribute{
				MarkdownDescription: "Administrative state of the transport [creation_pending, creation_proceed, creation_error, deployed, delete_pending, delete_proceed, delete_error]",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"vlans": schema.SingleNestedAttribute{
				MarkdownDescription: "Vlans of the transport",
				Default:             nil,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"a_vlan": schema.Int64Attribute{
						MarkdownDescription: "vlan for A side",
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"z_vlan": schema.Int64Attribute{
						MarkdownDescription: "vlan for Z side",
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "Connection ID created and returned by the cloud provider",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create transport creates the resource and sets the initial Terraform state.
func (r *transportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan transportResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.CreateTransport{
		Name: plan.Name.ValueString(),
		Product: models.AddProduct{
			SKU: plan.Product.SKU.ValueString(),
		},
	}

	// Create new transport
	transport, err := r.client.CreateTransport(ctx, payload, plan.WorkspaceID.ValueString(), autonomisdk.WithAdministrativeState(models.AdministrativeStateDeployed))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating transport",
			"Could not create transport, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(transport.ID.String())
	plan.State = types.StringValue(transport.State.String())
	plan.CreatedAt = types.StringValue(transport.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(transport.UpdatedAt.String())
	plan.DeployedAt = types.StringValue(transport.DeployedAt.String())
	plan.ConnectionID = types.StringValue(transport.ConnectionID)

	// set transportVlans object
	vlansObject, diag := types.ObjectValue(
		transportVlans,
		map[string]attr.Value{
			"a_vlan": types.Int64Value(transport.TransportVlans.AVlan),
			"z_vlan": types.Int64Value(transport.TransportVlans.ZVlan),
		},
	)
	plan.Vlans = vlansObject

	// Check for errors
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state transportResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed transport value from Autonomi
	transport, err := r.client.GetTransport(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Autonomi Transport",
			"Could not read Autonomi Transpor ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(transport.ID.String())
	state.CreatedAt = types.StringValue(transport.CreatedAt.String())
	state.UpdatedAt = types.StringValue(transport.UpdatedAt.String())
	state.DeployedAt = types.StringValue(transport.DeployedAt.String())
	state.Name = types.StringValue(transport.Name)
	state.State = types.StringValue(transport.State.String())
	state.Product = product{
		SKU: types.StringValue(transport.Product.SKU),
	}

	state.ConnectionID = types.StringValue(transport.ConnectionID)
	// set trnasportVlans object
	vlansObject, diag := types.ObjectValue(
		transportVlans,
		map[string]attr.Value{
			"a_vlan": types.Int64Value(transport.TransportVlans.AVlan),
			"z_vlan": types.Int64Value(transport.TransportVlans.ZVlan),
		},
	)
	state.Vlans = vlansObject

	// Check for errors
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan transportResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.UpdateElement{
		Name: plan.Name.ValueString(),
	}

	// Update existing workspace
	transport, err := r.client.UpdateTransport(ctx, payload, plan.WorkspaceID.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Transport",
			fmt.Sprintf("Could not update Autonomi transport: "+plan.ID.ValueString())+": error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(transport.ID.String())
	plan.Name = types.StringValue(transport.Name)
	plan.CreatedAt = types.StringValue(transport.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(transport.UpdatedAt.String())
	plan.DeployedAt = types.StringValue(transport.DeployedAt.String())
	plan.State = types.StringValue(transport.State.String())
	plan.Product = product{
		SKU: types.StringValue(transport.Product.SKU),
	}
	plan.ConnectionID = types.StringValue(transport.ConnectionID)

	// set transportVlans object
	vlansObject, diag := types.ObjectValue(
		transportVlans,
		map[string]attr.Value{
			"a_vlan": types.Int64Value(transport.TransportVlans.AVlan),
			"z_vlan": types.Int64Value(transport.TransportVlans.ZVlan),
		},
	)
	plan.Vlans = vlansObject

	// Check for errors
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state transportResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing node
	_, err := r.client.DeleteTransport(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString(), autonomisdk.WithAdministrativeState(models.AdministrativeStateDeleted))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting transport",
			"Could not delete transport, unexpected error: "+err.Error(),
		)
		return
	}
}
