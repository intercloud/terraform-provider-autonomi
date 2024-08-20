package autonomiresource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	autonomisdk "github.com/intercloud/autonomi-sdk"
	"github.com/intercloud/autonomi-sdk/models"
)

// attachmentResource is the resource implementation.
type attachmentResource struct {
	client *autonomisdk.Client
}

type attachmentResourceModel struct {
	ID          types.String `tfsdk:"id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	NodeID      types.String `tfsdk:"node_id"`
	TransportID types.String `tfsdk:"transport_id"`
	State       types.String `tfsdk:"administrative_state"`
	Side        types.String `tfsdk:"side"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &attachmentResource{}
	_ resource.ResourceWithConfigure = &attachmentResource{}
)

// NewAttachmentResource is a helper function to simplify the provider implementation.
func NewAttachmentResource() resource.Resource {
	return &attachmentResource{}
}

// Configure adds the provider configured client to the resource.
func (r *attachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *attachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attachment"
}

// Schema defines the schema for the resource.
func (r *attachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the attachment, set after creation",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the attachment",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Update date of the attachment",
				Computed:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace to which the attachment belongs.",
				Required:            true,
			},
			"node_id": schema.StringAttribute{
				MarkdownDescription: "ID of the node attached to the transport",
				Required:            true,
			},
			"transport_id": schema.StringAttribute{
				MarkdownDescription: "ID of the transport attached to the node.",
				Required:            true,
			},
			"administrative_state": schema.StringAttribute{
				MarkdownDescription: "Administrative state of the attachment [creation_pending, creation_proceed, creation_error, deployed, delete_pending, delete_proceed, delete_error]",
				Computed:            true,
			},
			"side": schema.StringAttribute{
				MarkdownDescription: "Direction of the attachment",
				Computed:            true,
			},
		},
	}
}

// CreateAttachment creates the resource and sets the initial Terraform state.
func (r *attachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan attachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	payload := models.CreateAttachment{
		NodeID:      plan.NodeID.ValueString(),
		TransportID: plan.TransportID.ValueString(),
	}

	// Create new attachment
	attachment, err := r.client.CreateAttachment(ctx, payload, plan.WorkspaceID.ValueString(), autonomisdk.WithAdministrativeState(models.AdministrativeStateDeployed))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating attachment",
			"Could not create attachment, unexpected error: "+err.Error(),
		)
		return
	}

	if attachment.State != models.AdministrativeStateDeployed {
		resp.Diagnostics.AddError(
			"Error creating attachment",
			"Attachment did not reach 'deployed' state in time.",
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(attachment.ID.String())
	plan.CreatedAt = types.StringValue(attachment.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(attachment.UpdatedAt.String())
	plan.State = types.StringValue(attachment.State.String())
	plan.NodeID = types.StringValue(attachment.NodeID)
	plan.TransportID = types.StringValue(attachment.TransportID)
	plan.Side = types.StringValue(attachment.Side)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *attachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state attachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed attachment value from Autonomi
	attachment, err := r.client.GetAttachment(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Autonomi attachment",
			"Could not read Autonomi attachment ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(attachment.ID.String())
	state.CreatedAt = types.StringValue(attachment.CreatedAt.String())
	state.UpdatedAt = types.StringValue(attachment.UpdatedAt.String())
	state.State = types.StringValue(attachment.State.String())
	state.NodeID = types.StringValue(attachment.NodeID)
	state.TransportID = types.StringValue(attachment.TransportID)
	state.Side = types.StringValue(attachment.Side)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *attachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *attachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state attachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing attachment
	_, err := r.client.DeleteAttachment(ctx, state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting attachment",
			"Could not delete attachment, unexpected error: "+err.Error(),
		)
		return
	}
}
