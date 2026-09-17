package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

type APIKeyResource struct{ client *management.Client }

var _ resource.ResourceWithConfigure = (*APIKeyResource)(nil)
var _ resource.ResourceWithImportState = (*APIKeyResource)(nil)
var _ resource.ResourceWithModifyPlan = (*APIKeyResource)(nil)

func NewAPIKeyResource() resource.Resource { return &APIKeyResource{} }
func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *APIKeyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan apiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !plan.known() {
		return
	}
	_, d := plan.request(ctx)
	resp.Diagnostics.Append(d...)
}
func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, d := plan.request(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.CreateAPIKey(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create API key", apiErrorDetail(err))
		return
	}
	plan.APIKey = types.StringValue(key.Secret)
	resp.Diagnostics.Append(plan.read(ctx, &key.APIKey, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.GetAPIKey(ctx, state.ID.ValueString())
	if missing(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read API key", apiErrorDetail(err, state.APIKey))
		return
	}
	resp.Diagnostics.Append(state.read(ctx, key, state.APIID.IsNull())...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state apiKeyModel
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
	key, err := r.client.UpdateAPIKey(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update API key", apiErrorDetail(err, state.APIKey))
		return
	}
	plan.APIKey = state.APIKey
	resp.Diagnostics.Append(plan.read(ctx, key, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIKey(ctx, state.ID.ValueString()); err != nil && !missing(err) {
		resp.Diagnostics.AddError("Unable to disable API key", apiErrorDetail(err, state.APIKey))
	}
}
func (*APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
