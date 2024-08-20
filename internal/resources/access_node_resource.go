package autonomiresource

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	autonomisdk "github.com/intercloud/autonomi-sdk"
	"github.com/intercloud/autonomi-sdk/models"
)

// accessNodeResource is the resource implementation.
type accessNodeResource struct {
	client *autonomisdk.Client
}

type accessNodeResourceModel struct {
	ID             types.String `tfsdk:"id"`
	WorkspaceID    types.String `tfsdk:"workspace_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	DeployedAt     types.String `tfsdk:"deployed_at"`
	Name           types.String `tfsdk:"name"`
	State          types.String `tfsdk:"administrative_state"`
	Type           types.String `tfsdk:"type"`
	Product        product      `tfsdk:"product"`
	PhysicalPortID types.String `tfsdk:"physical_port_id"`
	Vlan           types.Int64  `tfsdk:"vlan"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &accessNodeResource{}
	_ resource.ResourceWithConfigure = &accessNodeResource{}
)

// NewCloudNodeResource is a helper function to simplify the provider implementation.
func NewAccessNodeResource() resource.Resource {
	return &accessNodeResource{}
}

// Configure adds the provider configured client to the resource.
func (r *accessNodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *accessNodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_node"
}

// Schema defines the schema for the resource.
func (r *accessNodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the access node, set after creation",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the access node",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Update date of the access node",
				Computed:            true,
			},
			"deployed_at": schema.StringAttribute{
				MarkdownDescription: "Deployment date of the access node",
				Computed:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace to which the access node belongs.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the access node",
				Required:            true,
			},
			"physical_port_id": schema.StringAttribute{
				MarkdownDescription: "ID of the physical port id to which the access node is linked",
				Required:            true,
			},
			"administrative_state": schema.StringAttribute{
				MarkdownDescription: "Administrative state of the access node [creation_pending, creation_proceed, creation_error, deployed, delete_pending, delete_proceed, delete_error]",
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
			"vlan": schema.Int64Attribute{
				MarkdownDescription: "Vlan of the access node",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the node [access]",
				Computed:            true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *accessNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan accessNodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parsePhysicalPortID, err := uuid.Parse(plan.PhysicalPortID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing physical port id",
			"Could not create node, unexpected error: "+err.Error(),
		)
		return
	}

	// Generate API request body from plan
	payload := models.CreateNode{
		Name: plan.Name.ValueString(),
		Type: models.NodeTypeAccess,
		Product: models.AddProduct{
			SKU: plan.Product.SKU.ValueString(),
		},
		Vlan:           plan.Vlan.ValueInt64(),
		PhysicalPortID: &parsePhysicalPortID,
	}

	// Create new node
	node, err := r.client.CreateNode(ctx, payload, plan.WorkspaceID.ValueString(), autonomisdk.WithAdministrativeState(models.AdministrativeStateDeployed))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating node",
			"Could not create node, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(node.ID.String())
	plan.State = types.StringValue(node.State.String())
	plan.Type = types.StringValue(node.Type.String())
	plan.CreatedAt = types.StringValue(node.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(node.UpdatedAt.String())
	plan.DeployedAt = types.StringValue(node.DeployedAt.String())
	plan.Vlan = types.Int64Value(node.Vlan)
	plan.PhysicalPortID = types.StringValue(node.PhysicalPort.ID.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accessNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state accessNodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed node value from Autonomi
	node, err := r.client.GetNode(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Autonomi access node",
			"Could not read Autonomi access node ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(node.ID.String())
	state.CreatedAt = types.StringValue(node.CreatedAt.String())
	state.UpdatedAt = types.StringValue(node.UpdatedAt.String())
	state.DeployedAt = types.StringValue(node.DeployedAt.String())
	state.Name = types.StringValue(node.Name)
	state.State = types.StringValue(node.State.String())
	state.Type = types.StringValue(node.Type.String())
	state.Product = product{
		SKU: types.StringValue(node.Product.SKU),
	}
	state.PhysicalPortID = types.StringValue(node.PhysicalPort.ID.String())
	state.Vlan = types.Int64Value(node.Vlan)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accessNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan accessNodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.UpdateElement{
		Name: plan.Name.ValueString(),
	}

	// Update existing access node
	node, err := r.client.UpdateNode(ctx, payload, plan.WorkspaceID.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Access Node",
			fmt.Sprintf("Could not update Autonomi access node: "+plan.ID.ValueString())+": error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(node.ID.String())
	plan.Name = types.StringValue(node.Name)
	plan.CreatedAt = types.StringValue(node.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(node.UpdatedAt.String())
	plan.DeployedAt = types.StringValue(node.DeployedAt.String())
	plan.State = types.StringValue(node.State.String())
	plan.Type = types.StringValue(node.Type.String())
	plan.Product = product{
		SKU: types.StringValue(node.Product.SKU),
	}
	plan.Vlan = types.Int64Value(node.Vlan)
	plan.PhysicalPortID = types.StringValue(node.PhysicalPort.ID.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accessNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state accessNodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing node
	_, err := r.client.DeleteNode(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString(), autonomisdk.WithAdministrativeState(models.AdministrativeStateDeleted))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting node",
			"Could not delete node, unexpected error: "+err.Error(),
		)
		return
	}
}
