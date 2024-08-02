package provider

import (
	"context"
	"fmt"
	"terraform-provider-autonomi/external/autonomi"
	"terraform-provider-autonomi/external/autonomi/models"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &workspaceResource{}
	_ resource.ResourceWithConfigure = &workspaceResource{}
)

// NewWorkspaceResource is a helper function to simplify the provider implementation.
func NewWorkspaceResource() resource.Resource {
	return &workspaceResource{}
}

// workspaceResource is the resource implementation.
type workspaceResource struct {
	client *autonomi.Client
}

type workspaceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AccountID   types.String `tfsdk:"account_id"`
}

// Metadata returns the resource type name.
func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

// Configure adds the provider configured client to the resource.
func (r *workspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema defines the schema for the resource.
func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {

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
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"account_id": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan workspaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.CreateWorkspace{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Create new workspace
	workspace, err := r.client.CreateWorkspace(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace",
			"Could not create workspace, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(context.Background(), "workspace created", map[string]interface{}{
		"workspace": workspace,
	})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(workspace.ID.String())
	plan.CreatedAt = types.StringValue(workspace.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(workspace.UpdatedAt.String())
	plan.AccountID = types.StringValue(r.client.AccountID)

	tflog.Debug(context.Background(), "plan updated", map[string]interface{}{
		"plan": workspace,
	})

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	// Get current state
	var state workspaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed workspace value from Autonomi
	workspace, err := r.client.GetWorkspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Autonomi Workspace",
			fmt.Sprintf("Could not read Autonomi workspace:  %s/accounts/%s/workspaces/%s", r.client.HostURL, r.client.AccountID, state.ID.ValueString())+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(workspace.ID.String())
	state.Name = types.StringValue(workspace.Name)
	state.Description = types.StringValue(workspace.Description)
	state.CreatedAt = types.StringValue(workspace.CreatedAt.String())
	state.UpdatedAt = types.StringValue(workspace.UpdatedAt.String())
	state.AccountID = types.StringValue(workspace.AccountID)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state workspaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing workspace
	err := r.client.DeleteWorkspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Workspace",
			"Could not delete workspace, unexpected error: "+err.Error(),
		)
		return
	}
}
