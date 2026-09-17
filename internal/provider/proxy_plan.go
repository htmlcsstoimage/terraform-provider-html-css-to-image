package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithModifyPlan = (*ProxyResource)(nil)

// Validate complete known inputs before any write. Unknown values remain unknown
// during planning and receive the same validation once they resolve at apply.
func (r *ProxyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan proxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !plan.inputsKnown() {
		return
	}
	var old *proxyModel
	if !req.State.Raw.IsNull() {
		var state proxyModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		old = &state
	}
	_, d := plan.request(ctx, old)
	resp.Diagnostics.Append(d...)
}
func (m proxyModel) inputsKnown() bool {
	if m.Name.IsUnknown() || m.URL.IsUnknown() || m.Port.IsUnknown() || m.Disabled.IsUnknown() || m.BypassHosts.IsUnknown() || m.Authentication.IsUnknown() {
		return false
	}
	for _, v := range m.BypassHosts.Elements() {
		if v.IsUnknown() {
			return false
		}
	}
	for _, v := range m.Authentication.Attributes() {
		if v.IsUnknown() {
			return false
		}
	}
	return true
}
