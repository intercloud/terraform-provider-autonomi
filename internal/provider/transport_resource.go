package provider

import (
	"context"
	"fmt"
	"time"

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
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Update date of the transport",
				Computed:            true,
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
			},
			"administrative_state": schema.StringAttribute{
				MarkdownDescription: "Administrative state of the transport [creation_pending, creation_proceed, creation_error, deployed, delete_pending, delete_proceed, delete_error]",
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
	transport, err := r.client.CreateTransport(ctx, payload, plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating transport",
			"Could not create transport, unexpected error: "+err.Error(),
		)
		return
	}

	// Poll the node status until it's "deployed" - TODO: find better polling system with terraform
	const maxRetries = 30
	const retryInterval = 20 * time.Second

	for i := 0; i < maxRetries; i++ {
		transport, err = r.client.GetTransport(ctx, plan.WorkspaceID.ValueString(), transport.ID.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting transport status",
				"Could not get nodtransporte status, unexpected error: "+err.Error(),
			)
			return
		}

		if transport.State == models.AdministrativeStateCreationError {
			resp.Diagnostics.AddError(
				"Error creating transport",
				"Could not creating transport, unexpected error: code:"+transport.Error.Code+" msg: "+transport.Error.Msg,
			)
			return
		}

		if transport.State == models.AdministrativeStateDeployed {
			break
		}

		time.Sleep(retryInterval)
	}

	if transport.State != models.AdministrativeStateDeployed {
		resp.Diagnostics.AddError(
			"Error creating transport",
			"Node did not reach 'deployed' state in time.",
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(transport.ID.String())
	plan.State = types.StringValue(transport.State.String())
	plan.CreatedAt = types.StringValue(transport.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(transport.UpdatedAt.String())
	if transport.DeployedAt != nil {
		plan.DeployedAt = types.StringValue(transport.DeployedAt.String())
	}
	plan.ConnectionID = types.StringValue(transport.ConnectionID)

	// set trnasportVlans object
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
}

func (r *transportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *transportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
