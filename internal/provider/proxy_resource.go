package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/htmlcsstoimage/go-client/management"
)

type ProxyResource struct{ client *management.Client }

var _ resource.ResourceWithConfigure = (*ProxyResource)(nil)
var _ resource.ResourceWithImportState = (*ProxyResource)(nil)

func NewProxyResource() resource.Resource { return &ProxyResource{} }
func (r *ProxyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func missing(err error) bool {
	var e *management.APIError
	return errors.As(err, &e) && e.StatusCode == 404
}
func (r *ProxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan proxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx, nil)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.CreateProxy(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create proxy", apiErrorDetail(err, plan.Authentication))
		return
	}
	resp.Diagnostics.Append(plan.read(ctx, p, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *ProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state proxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.GetProxy(ctx, state.ID.ValueString())
	if missing(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read proxy", apiErrorDetail(err, state.Authentication))
		return
	}
	importing := state.Name.IsNull()
	resp.Diagnostics.Append(state.read(ctx, p, importing)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *ProxyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state proxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx, &state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.UpdateProxy(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update proxy", apiErrorDetail(err, plan.Authentication, state.Authentication))
		return
	}
	resp.Diagnostics.Append(plan.read(ctx, p, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *ProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state proxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProxy(ctx, state.ID.ValueString()); err != nil && !missing(err) {
		resp.Diagnostics.AddError("Unable to delete proxy", apiErrorDetail(err, state.Authentication))
	}
}
func (*ProxyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
