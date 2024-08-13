package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	autonomisdk "github.com/intercloud/autonomi-sdk"
	"github.com/intercloud/autonomi-sdk/models"
)

// cloudNodeResource is the resource implementation.
type cloudNodeResource struct {
	client *autonomisdk.Client
}

type product struct {
	// Provider types.String `tfsdk:"provider"`
	SKU types.String `tfsdk:"sku"`
}

type providerCloudConfig struct {
	AWSAccountID    types.String `tfsdk:"aws_account_id"`
	GCPPairingKey   types.String `tfsdk:"gcp_pairing_key"`
	AzureServiceKey types.String `tfsdk:"azure_service_key"`
}

type cloudNodeResourceModel struct {
	ID             types.String        `tfsdk:"id"`
	WorkspaceID    types.String        `tfsdk:"workspace_id"`
	CreatedAt      types.String        `tfsdk:"created_at"`
	UpdatedAt      types.String        `tfsdk:"updated_at"`
	Name           types.String        `tfsdk:"name"`
	State          types.String        `tfsdk:"administrative_state"`
	Type           types.String        `tfsdk:"type"`
	Product        product             `tfsdk:"product"`
	ProviderConfig providerCloudConfig `tfsdk:"provider_config"`
	ConnectionID   types.String        `tfsdk:"connection_id"`
	Vlan           types.Int64         `tfsdk:"vlan"`
	DxconID        types.String        `tfsdk:"dxcon_id"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &cloudNodeResource{}
	_ resource.ResourceWithConfigure = &cloudNodeResource{}
)

// NewCloudNodeResource is a helper function to simplify the provider implementation.
func NewCloudNodeResource() resource.Resource {
	return &cloudNodeResource{}
}

// Configure adds the provider configured client to the resource.
func (r *cloudNodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *cloudNodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_node"
}

// Schema defines the schema for the resource.
func (r *cloudNodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the cloud node, set after creation",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the cloud node",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Update date of the cloud node",
				Computed:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace to which the cloud node belongs.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the cloud node",
				Required:            true,
			},
			"administrative_state": schema.StringAttribute{
				MarkdownDescription: "Administrative state of the cloud node [creation_pending, creation_proceed, creation_error, deployed, delete_pending, delete_proceed, delete_error]",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the node [cloud]",
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
			"provider_config": schema.SingleNestedAttribute{
				Required: true, // only for cloud nodes
				Attributes: map[string]schema.Attribute{
					"aws_account_id": schema.StringAttribute{
						MarkdownDescription: "AWS Account ID where the resource will be created",
						Optional:            true,
						Default:             nil,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"gcp_pairing_key": schema.StringAttribute{
						MarkdownDescription: "GCP Pairing Key where the resource will be created",
						Optional:            true,
						Default:             nil,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"azure_service_key": schema.StringAttribute{
						MarkdownDescription: "Azure Service Key where the resource will be created",
						Optional:            true,
						Default:             nil,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
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
			"vlan": schema.Int64Attribute{
				MarkdownDescription: "Vlan of the cloud node",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dxcon_id": schema.StringAttribute{
				MarkdownDescription: "Dxcon ID created and returned by AWS",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// CreateNode creates the resource and sets the initial Terraform state.
func (r *cloudNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan cloudNodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.CreateNode{
		Name: plan.Name.ValueString(),
		Type: models.NodeTypeCloud,
		Product: models.AddProduct{
			SKU: plan.Product.SKU.ValueString(),
		},
		ProviderConfig: &models.ProviderCloudConfig{
			AccountID:  plan.ProviderConfig.AWSAccountID.ValueString(),
			PairingKey: plan.ProviderConfig.GCPPairingKey.ValueString(),
			ServiceKey: plan.ProviderConfig.AzureServiceKey.ValueString(),
		},
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
	plan.ConnectionID = types.StringValue(node.ConnectionID)
	plan.Vlan = types.Int64Value(node.Vlan)
	plan.DxconID = types.StringValue(node.DxconID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cloudNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state cloudNodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed node value from Autonomi
	node, err := r.client.GetNode(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Autonomi cloud node",
			"Could not read Autonomi cloud node ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(node.ID.String())
	state.CreatedAt = types.StringValue(node.CreatedAt.String())
	state.UpdatedAt = types.StringValue(node.UpdatedAt.String())
	state.Name = types.StringValue(node.Name)
	state.State = types.StringValue(node.State.String())
	state.Type = types.StringValue(node.Type.String())
	state.Product = product{
		SKU: types.StringValue(node.Product.SKU),
	}
	state.ProviderConfig = providerCloudConfig{
		AWSAccountID:    types.StringValue(node.ProviderConfig.AccountID),
		GCPPairingKey:   types.StringValue(node.ProviderConfig.PairingKey),
		AzureServiceKey: types.StringValue(node.ProviderConfig.ServiceKey),
	}

	state.ConnectionID = types.StringValue(node.ConnectionID)
	state.Vlan = types.Int64Value(node.Vlan)
	state.DxconID = types.StringValue(node.DxconID)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cloudNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan cloudNodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.UpdateElement{
		Name: plan.Name.ValueString(),
	}

	// Update existing cloud node
	node, err := r.client.UpdateNode(ctx, payload, plan.WorkspaceID.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Cloud Node",
			fmt.Sprintf("Could not update Autonomi cloud node: "+plan.ID.ValueString())+": error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(node.ID.String())
	plan.Name = types.StringValue(node.Name)
	plan.CreatedAt = types.StringValue(node.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(node.UpdatedAt.String())
	plan.State = types.StringValue(node.State.String())
	plan.Type = types.StringValue(node.Type.String())
	plan.Product = product{
		SKU: types.StringValue(node.Product.SKU),
	}

	if node.ProviderConfig != nil {
		if node.ProviderConfig.AccountID != "" {
			plan.ProviderConfig.AWSAccountID = types.StringValue(node.ProviderConfig.AccountID)
		} else if node.ProviderConfig.PairingKey != "" {
			plan.ProviderConfig.GCPPairingKey = types.StringValue(node.ProviderConfig.PairingKey)
		} else if node.ProviderConfig.ServiceKey != "" {
			plan.ProviderConfig.AzureServiceKey = types.StringValue(node.ProviderConfig.ServiceKey)
		}
	}

	plan.ConnectionID = types.StringValue(node.ConnectionID)
	plan.Vlan = types.Int64Value(node.Vlan)
	plan.DxconID = types.StringValue(node.DxconID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cloudNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state cloudNodeResourceModel
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
