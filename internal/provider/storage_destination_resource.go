package provider

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

type StorageDestinationResource struct{ client *management.Client }

var _ resource.ResourceWithConfigure = (*StorageDestinationResource)(nil)
var _ resource.ResourceWithImportState = (*StorageDestinationResource)(nil)
var _ resource.ResourceWithModifyPlan = (*StorageDestinationResource)(nil)

func NewStorageDestinationResource() resource.Resource { return &StorageDestinationResource{} }
func (r *StorageDestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*management.Client)
	if !ok {
		resp.Diagnostics.AddError("Invalid provider configuration", "Expected an HCTI management client.")
		return
	}
	r.client = c
}
func (r *StorageDestinationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan storageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !plan.known() {
		return
	}
	var old *storageModel
	if !req.State.Raw.IsNull() {
		old = &storageModel{}
		resp.Diagnostics.Append(req.State.Get(ctx, old)...)
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(plan.validate(old)...)
	}
}
func (r *StorageDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx, nil)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.CreateStorageDestination(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create storage destination", apiErrorDetail(err, plan.Connection))
		return
	}
	want := plan
	plan.read(remote, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if fields := storageUnapplied(want, plan); len(fields) > 0 {
		sort.Strings(fields)
		resp.Diagnostics.AddError("Storage destination partially applied", "The API returned different settings for: "+strings.Join(fields, ", ")+". Actual state has been saved.")
	}
}
func (r *StorageDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.GetStorageDestination(ctx, state.ID.ValueString())
	if missing(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read storage destination", apiErrorDetail(err, state.Connection))
		return
	}
	state.read(remote, state.Connection.IsNull())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *StorageDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, old storageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx, &old)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.UpdateStorageDestination(ctx, old.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update storage destination", apiErrorDetail(err, plan.Connection, old.Connection))
		return
	}
	want := plan
	plan.read(remote, false)
	fields := storageUnapplied(want, plan)
	// Secrets are unreadable. A supplied changed secret always triggers an API
	// connection test, including while disabled. A disable-only downgrade skips it.
	secret := storageString(want.Connection, "secret_access_key")
	prior := storageString(old.Connection, "secret_access_key")
	changedSecret := !secret.IsNull() && (!storageSameIdentity(want.Connection, old.Connection) || prior.IsNull() || storageNormalized("secret_access_key", secret.ValueString()) != storageNormalized("secret_access_key", prior.ValueString()))
	if changedSecret && (len(fields) > 0 || (want.Disabled.ValueBool() && !old.Disabled.ValueBool() && plan.LastTestedAt.Equal(old.LastTestedAt))) {
		a := storageFlat(plan.Connection).Attributes()
		a["secret_access_key"] = types.StringNull()
		if storageSameIdentity(plan.Connection, old.Connection) {
			a["secret_access_key"] = prior
		}
		plan.Connection = storageNested(types.ObjectValueMust(storageFlatTypes(), a))
		fields = append(fields, "connection_info."+storageString(want.Connection, "provider").ValueString()+".secret_access_key (not confirmed)")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if len(fields) > 0 {
		sort.Strings(fields)
		resp.Diagnostics.AddError("Storage destination partially applied", "The API did not confirm: "+strings.Join(fields, ", ")+". Actual readable state has been saved. After a plan downgrade, the API may only permit disabling the destination.")
	}
}
func (r *StorageDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteStorageDestination(ctx, state.ID.ValueString()); err != nil && !missing(err) {
		resp.Diagnostics.AddError("Unable to delete storage destination", apiErrorDetail(err, state.Connection))
	}
}
func (*StorageDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
