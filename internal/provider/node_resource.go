package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-autonomi/external/autonomi"
	"terraform-provider-autonomi/external/autonomi/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// nodeResource is the resource implementation.
type nodeResource struct {
	client *autonomi.Client
}

type product struct {
	// Provider types.String `tfsdk:"provider"`
	SKU types.String `tfsdk:"sku"`
}

type providerCloudConfig struct {
	AWSAccountID types.String `tfsdk:"aws_account_id"`
}

type nodeResourceModel struct {
	ID             types.String        `tfsdk:"id"`
	AccountID      types.String        `tfsdk:"account_id"`
	WorkspaceID    types.String        `tfsdk:"workspace_id"`
	CreatedAt      types.String        `tfsdk:"created_at"`
	UpdatedAt      types.String        `tfsdk:"updated_at"`
	Name           types.String        `tfsdk:"name"`
	State          types.String        `tfsdk:"administrative_state"`
	Type           types.String        `tfsdk:"type"`
	Product        product             `tfsdk:"product"`
	ProviderConfig providerCloudConfig `tfsdk:"provider_config"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &nodeResource{}
	_ resource.ResourceWithConfigure = &nodeResource{}
)

// NewCloudNodeResource is a helper function to simplify the provider implementation.
func NewCloudNodeResource() resource.Resource {
	return &nodeResource{}
}

// Configure adds the provider configured client to the resource.
func (r *nodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*autonomi.Client)

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
func (r *nodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

// Schema defines the schema for the resource.
func (r *nodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"account_id": schema.StringAttribute{
				Required: true,
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"administrative_state": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"product": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"sku": schema.StringAttribute{
						Required: true,
					},
				},
			},
			"provider_config": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"aws_account_id": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

// CreateNode creates the resource and sets the initial Terraform state.
func (r *nodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan nodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.CreateNode{
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Name:        plan.Name.ValueString(),
		Type:        plan.Type.ValueString(),
		Product: models.NodeProduct{
			Product: models.Product{
				SKU: plan.Product.SKU.ValueString(),
			},
		},
		ProviderConfig: &models.ProviderCloudConfig{
			AccountID: plan.ProviderConfig.AWSAccountID.ValueString(),
		},
	}

	// Create new node
	node, err := r.client.CreateNode(payload, plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating node",
			"Could not create node, unexpected error: "+err.Error(),
		)
		return
	}

	// Poll the node status until it's "deployed" - TODO: find better polling system with terraform
	const maxRetries = 30
	const retryInterval = 20 * time.Second

	for i := 0; i < maxRetries; i++ {
		node, err = r.client.GetNode(plan.WorkspaceID.ValueString(), node.ID.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting node status",
				"Could not get node status, unexpected error: "+err.Error(),
			)
			return
		}

		if node.State == "deployed" {
			break
		}

		time.Sleep(retryInterval)
	}

	if node.State != "deployed" {
		resp.Diagnostics.AddError(
			"Error creating node",
			"Node did not reach 'deployed' state in time.",
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(node.ID.String())
	plan.State = types.StringValue(node.State)
	plan.CreatedAt = types.StringValue(node.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(node.UpdatedAt.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	tflog.Info(ctx, "=============>>>>>>>>>  WELCOME_TO_READ <<<<<<<<<<<=============")

	// Get current state
	var state nodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed node value from Autonomi
	node, err := r.client.GetNode(state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HashiCups Order",
			"Could not read HashiCups order ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(node.ID.String())
	state.CreatedAt = types.StringValue(node.CreatedAt.String())
	state.UpdatedAt = types.StringValue(node.UpdatedAt.String())
	state.AccountID = types.StringValue(node.AccountID)
	state.Name = types.StringValue(node.Name)
	state.State = types.StringValue(node.State)
	state.Type = types.StringValue(node.Type)
	state.Product = product{
		SKU: types.StringValue(node.Product.SKU),
	}
	state.ProviderConfig = providerCloudConfig{
		AWSAccountID: types.StringValue(node.ProviderConfig.AccountID),
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *nodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state nodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing node
	err := r.client.DeleteNode(state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting node",
			"Could not delete node, unexpected error: "+err.Error(),
		)
		return
	}

	// Poll the node until it's deleted - TODO: use terraform function if possible
	const maxRetries = 30
	const retryInterval = 20 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err = r.client.GetNode(state.WorkspaceID.ValueString(), state.ID.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "status: 404") {
				// Handle 404 error specifically
				break
			}
			resp.Diagnostics.AddError(
				"Error getting node status",
				"Could not get node status, unexpected error: "+err.Error(),
			)
			return
		}

		time.Sleep(retryInterval)
	}
}
