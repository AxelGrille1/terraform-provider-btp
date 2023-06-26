package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp/internal/btpcli"
	"github.com/SAP/terraform-provider-btp/internal/btpcli/types/cis"
	"github.com/SAP/terraform-provider-btp/internal/tfutils"
	"github.com/SAP/terraform-provider-btp/internal/validation/uuidvalidator"
)

func newDirectoryResource() resource.Resource {
	return &directoryResource{}
}

type directoryResource struct {
	cli *btpcli.ClientFacade
}

func (rs *directoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_directory", req.ProviderTypeName)
}

func (rs *directoryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.cli = req.ProviderData.(*btpcli.ClientFacade)
}

func (rs *directoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Directories allow you to organize and manage your subaccounts according to your technical and business needs. The use of directories is optional.

You can create up to five levels of directories in your account hierarchy. If you have directories, you can still create subaccounts directly under your global account.

__Tips:__
* You must be assigned to the global account admin role, or the directory admin if the directory is configured to manage its authorizations.
* A directory path in the account hierarchy can have only one directory that is enabled with the ` + "`ENTITLEMENTS`" + ` or ` + "`AUTHORIZATIONS`" + ` features. If such a directory exists, other directories in that path can only be enabled with the ` + "`DEFAULT`" + ` features.

__Further documentation:__
<https://help.sap.com/docs/btp/sap-business-technology-platform/account-model>`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the directory.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[^\/]{1,255}$`), "must not contain '/', not be empty and not exceed 255 characters"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the directory.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(300),
				},
			},
			"parent_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the directory's parent entity. Typically this is the global account.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: "Applies only to directories that have the user authorization management feature enabled. The subdomain becomes part of the path used to access the authorization tenant of the directory. It has to be unique within the defined region.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[a-z0-9](?:[a-z0-9|-]{0,61}[a-z0-9])?$"), "must only contain letters (a-z), digits (0-9), and hyphens (not at the start or end)"),
				},
			},
			"labels": schema.MapAttribute{
				ElementType: types.SetType{
					ElemType: types.StringType,
				},
				MarkdownDescription: "Contains information about the labels assigned to a specified global account. Labels are represented in a JSON array of key-value pairs; each key has up to 10 corresponding values.",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the directory.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Details of the user that created the directory.",
				Computed:            true,
			},
			"created_date": schema.StringAttribute{
				MarkdownDescription: "The date and time the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
				Computed:            true,
			},

			"features": schema.SetAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "The features that are enabled for the directory. Possible values are: " +
					"\n\t - `DEFAULT` All directories have the following basic feature enabled: " +
					"\n\t\t 1. Group and filter subaccounts for reports and filters " +
					"\n\t\t 2. Monitor usage and costs on a directory level (costs only available for contracts that use the consumption-based commercial model)" +
					"\n\t\t 3. Set custom properties and tags to the directory for identification and reporting purposes." +
					"\n\t - `ENTITLEMENTS` Allows the assignment of a quota for services and applications to the directory from the global account quota for distribution to the subaccounts under this directory." +
					"\n\t - `AUTHORIZATIONS` Allows the assignment of users as administrators or viewers of this directory. You must apply this feature in combination with the `ENTITLEMENTS` feature.",
				Computed: true,
			},
			"last_modified": schema.StringAttribute{
				MarkdownDescription: "The date and time the resource was last modified in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
				Computed:            true,
			},

			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the directory. Possible values are: " +
					"\n\t - `OK` The CRUD operation or series of operations completed successfully." +
					"\n\t - `STARTED` CRUD operation on an entity has started." +
					"\n\t - `CREATING` Creating entity operation is in progress." +
					"\n\t - `UPDATING` Updating entity operation is in progress." +
					"\n\t - `MOVING` Moving entity operation is in progress." +
					"\n\t - `PROCESSING` A series of operations related to the entity is in progress." +
					"\n\t - `DELETING` Deleting entity operation is in progress." +
					"\n\t - `PENDING REVIEW` The processing operation has been stopped for reviewing and can be restarted by the operator." +
					"\n\t - `CANCELLED` The operation or processing was canceled by the operator." +
					"\n\t - `CREATION_FAILED` The creation operation failed, and the entity was not created or was created but cannot be used." +
					"\n\t - `UPDATE_FAILED` The update operation failed, and the entity was not updated." +
					"\n\t - `PROCESSING_FAILED` The processing operations failed." +
					"\n\t - `DELETION_FAILED` The delete operation failed, and the entity was not deleted." +
					"\n\t - `MOVE_FAILED` Entity could not be moved to a different location." +
					"\n\t - `MIGRATING` Migrating entity from Neo to Cloud Foundry.",
				Computed: true,
			},
		},
	}
}

func (rs *directoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state directoryType

	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cliRes, _, err := rs.cli.Accounts.Directory.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error Reading Resource Directory", fmt.Sprintf("%s", err))
		return
	}

	state, diags = directoryValueFrom(ctx, cliRes)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (rs *directoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan directoryType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := btpcli.DirectoryCreateInput{
		DisplayName: plan.Name.ValueString(),
	}

	if !plan.Description.IsUnknown() {
		description := plan.Description.ValueString()
		args.Description = &description
	}

	if !plan.ParentID.IsUnknown() {
		parentID := plan.ParentID.ValueString()
		args.ParentID = &parentID
	}

	if !plan.Subdomain.IsUnknown() {
		subdomain := plan.Subdomain.ValueString()
		args.Subdomain = &subdomain
	}

	if !plan.Labels.IsUnknown() {
		var labels map[string][]string
		plan.Labels.ElementsAs(ctx, &labels, false)
		args.Labels = labels
	}

	cliRes, _, err := rs.cli.Accounts.Directory.Create(ctx, &args)
	if err != nil {
		resp.Diagnostics.AddError("API Error Creating Resource Directory", fmt.Sprintf("%s", err))
		return
	}

	plan, diags = directoryValueFrom(ctx, cliRes)
	resp.Diagnostics.Append(diags...)

	createStateConf := &tfutils.StateChangeConf{
		Pending: []string{cis.StateCreating, cis.StateStarted},
		Target:  []string{cis.StateOK, cis.StateCreationFailed, cis.StateCanceled},
		Refresh: func() (interface{}, string, error) {
			subRes, _, err := rs.cli.Accounts.Directory.Get(ctx, cliRes.Guid)

			if err != nil {
				return subRes, "", err
			}

			return subRes, subRes.EntityState, nil
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	updatedRes, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("API Error Creating Resource Directory", fmt.Sprintf("%s", err))
	}

	plan, diags = directoryValueFrom(ctx, updatedRes.(cis.DirectoryResponseObject))
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (rs *directoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan directoryType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("API Error Updating Resource Directory", "Update is not yet implemented.")

	/*TODO: cliRes, err := rs.cli.Execute(ctx, btpcli.Update, rs.command, plan)
	if err != nil {
		resp.Diagnostics.AddError("API Error Updating Resource Directory", fmt.Sprintf("%s", err))
		return
	}*/

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (rs *directoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state directoryType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cliRes, _, err := rs.cli.Accounts.Directory.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error Deleting Resource Directory", fmt.Sprintf("%s", err))
		return
	}

	deleteStateConf := &tfutils.StateChangeConf{
		Pending: []string{cis.StateDeleting, cis.StateStarted},
		Target:  []string{cis.StateOK, cis.StateDeletionFailed, cis.StateCanceled, "DELETED"},
		Refresh: func() (interface{}, string, error) {
			subRes, comRes, err := rs.cli.Accounts.Directory.Get(ctx, cliRes.Guid)

			if comRes.StatusCode == http.StatusNotFound || comRes.StatusCode == http.StatusForbidden {
				return subRes, "DELETED", nil
			}

			if err != nil {
				return subRes, subRes.EntityState, err
			}

			return subRes, subRes.EntityState, nil
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = deleteStateConf.WaitForStateContext(ctx)

	if err != nil {
		resp.Diagnostics.AddError("API Error Deleting Resource Directory", fmt.Sprintf("%s", err))
		return
	}
}

func (rs *directoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
