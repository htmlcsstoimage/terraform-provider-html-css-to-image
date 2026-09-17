package provider

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/htmlcsstoimage/go-client/management"
)

type OGConfigResource struct{ client *management.Client }

var _ resource.ResourceWithConfigure = (*OGConfigResource)(nil)
var _ resource.ResourceWithImportState = (*OGConfigResource)(nil)
var _ resource.ResourceWithModifyPlan = (*OGConfigResource)(nil)

func NewOGConfigResource() resource.Resource { return &OGConfigResource{} }
func (r *OGConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *OGConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan ogConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !plan.known() {
		return
	}
	resp.Diagnostics.Append(plan.validate(ctx)...)
}
func (r *OGConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ogConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.CreateOGConfig(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create OG configuration", apiErrorDetail(err, plan.Headers, plan.DefaultOptions))
		return
	}
	want := plan
	resp.Diagnostics.Append(plan.read(ctx, remote, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if fields := ogUnapplied(want, plan); len(fields) > 0 {
		sort.Strings(fields)
		resp.Diagnostics.AddError("OG configuration partially applied", "The API returned different settings for: "+strings.Join(fields, ", ")+". Actual state has been saved; review the configuration before retrying.")
	}
}
func (r *OGConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ogConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.GetOGConfig(ctx, state.ID.ValueString())
	if missing(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read OG configuration", apiErrorDetail(err, state.Headers, state.DefaultOptions))
		return
	}
	resp.Diagnostics.Append(state.read(ctx, remote, state.ConfigType.IsNull())...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
func (r *OGConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ogConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.UpdateOGConfig(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update OG configuration", apiErrorDetail(err, plan.Headers, plan.DefaultOptions, state.Headers, state.DefaultOptions))
		return
	}
	want := plan
	resp.Diagnostics.Append(plan.read(ctx, remote, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if fields := ogUnapplied(want, plan); len(fields) > 0 {
		sort.Strings(fields)
		resp.Diagnostics.AddError("OG configuration partially applied", "The API did not apply: "+strings.Join(fields, ", ")+". Actual state has been saved. After a plan downgrade, the API may only allow disabling the configuration.")
	}
}
func (r *OGConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ogConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOGConfig(ctx, state.ID.ValueString()); err != nil && !missing(err) {
		resp.Diagnostics.AddError("Unable to delete OG configuration", apiErrorDetail(err, state.Headers, state.DefaultOptions))
	}
}
func (*OGConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
