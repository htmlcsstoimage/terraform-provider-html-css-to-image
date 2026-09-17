package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

// RenderResource shares lifecycle mechanics while retaining separate public schemas.
type RenderResource struct {
	kind   string
	client *management.Client
}

var _ resource.ResourceWithConfigure = (*RenderResource)(nil)
var _ resource.ResourceWithModifyPlan = (*RenderResource)(nil)
var _ resource.ResourceWithImportState = (*RenderResource)(nil)

func NewTemplateResource() resource.Resource       { return &RenderResource{kind: "template"} }
func NewImageHTMLCSSResource() resource.Resource   { return &RenderResource{kind: "html_css"} }
func NewImageURLResource() resource.Resource       { return &RenderResource{kind: "url"} }
func NewImageTemplatedResource() resource.Resource { return &RenderResource{kind: "templated"} }
func (r *RenderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func renderState(ctx context.Context, s tfsdk.State) (map[string]attr.Value, diag.Diagnostics) {
	var o types.Object
	d := s.Get(ctx, &o)
	return o.Attributes(), d
}
func renderPlan(ctx context.Context, s tfsdk.Plan) (map[string]attr.Value, diag.Diagnostics) {
	var o types.Object
	d := s.Get(ctx, &o)
	return o.Attributes(), d
}
func setRenderState(ctx context.Context, s *tfsdk.State, a map[string]attr.Value) diag.Diagnostics {
	t := s.Schema.Type().(types.ObjectType).AttrTypes
	// Initialize computed fields before saving a newly created identity, even if readback fails.
	for k, v := range t {
		if a[k] == nil || a[k].IsUnknown() {
			a[k] = nullValue(v)
		}
	}
	o, d := types.ObjectValue(t, a)
	if !d.HasError() {
		d.Append(s.Set(ctx, o)...)
	}
	return d
}
func (r *RenderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	a, d := renderPlan(ctx, req.Plan)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	resp.Diagnostics.Append(validateRender(ctx, r.kind, a)...)
	inputs := renderAttributes(r.kind)
	if !req.State.Raw.IsNull() {
		prior, ds := renderState(ctx, req.State)
		resp.Diagnostics.Append(ds...)
		if !ds.HasError() {
			same := true
			for k := range inputs {
				if !renderEquivalent(k, a[k], prior[k]) {
					same = false
				}
				if renderEquivalent(k, a[k], prior[k]) && !a[k].Equal(prior[k]) {
					resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(k), prior[k])...)
					a[k] = prior[k]
				}
			}
			if same {
				for k, v := range prior {
					if _, input := inputs[k]; !input {
						resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(k), v)...)
						a[k] = v
					}
				}
			}
		}
	}
	if req.State.Raw.IsNull() || r.kind == "template" {
		return
	}
	prior, d := renderState(ctx, req.State)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	for k := range inputs {
		if !renderEquivalent(k, a[k], prior[k]) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(k))
		}
	}
	// Imported images retain unpinned intent; an explicit mismatching pin must replace them.
	if r.kind == "templated" && !a["template_version"].IsNull() && !a["template_version"].IsUnknown() && !a["template_version"].Equal(prior["resolved_template_version"]) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("template_version"))
	}
}
func (r *RenderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	a, d := renderPlan(ctx, req.Plan)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	resp.Diagnostics.Append(validateRender(ctx, r.kind, a)...)
	body, d := renderRequest(ctx, a)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result *management.CreatedRender
	var err error
	if r.kind == "template" {
		result, err = r.client.SaveTemplateDefinition(ctx, "", body)
	} else {
		result, err = r.client.CreateImageDefinition(ctx, r.kind, body)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create rendering definition", apiErrorDetail(err, a["headers"], a["template_values"]))
		return
	}
	if r.kind == "template" {
		a["id"] = types.StringValue(result.TemplateID)
		a["version"] = types.Int64Value(result.TemplateVersion)
	} else {
		a["id"] = types.StringValue(result.ID)
		a["image_url"] = types.StringValue(result.URL)
	}
	resp.Diagnostics.Append(setRenderState(ctx, &resp.State, a)...)
	resp.Diagnostics.Append(r.read(ctx, a, false)...)
	resp.Diagnostics.Append(setRenderState(ctx, &resp.State, a)...)
}
func (r *RenderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	a, d := renderState(ctx, req.State)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	id := a["id"].(types.String).ValueString()
	remote, err := r.fetch(ctx, id)
	if missing(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read rendering definition", apiErrorDetail(err, a["headers"], a["template_values"]))
		return
	}
	imported := a["created_at"].IsNull()
	resp.Diagnostics.Append(r.merge(ctx, a, remote, imported)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(setRenderState(ctx, &resp.State, a)...)
	}
}
func (r *RenderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.kind != "template" {
		resp.Diagnostics.AddError("Image replacement required", "Image inputs cannot be updated in place.")
		return
	}
	a, d := renderPlan(ctx, req.Plan)
	resp.Diagnostics.Append(d...)
	prior, d := renderState(ctx, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateRender(ctx, r.kind, a)...)
	body, d := renderRequest(ctx, a)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := prior["id"].(types.String).ValueString()
	result, err := r.client.SaveTemplateDefinition(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update template", apiErrorDetail(err, a["headers"], a["template_values"]))
		return
	}
	a["id"] = types.StringValue(id)
	a["version"] = types.Int64Value(result.TemplateVersion)
	resp.Diagnostics.Append(setRenderState(ctx, &resp.State, a)...)
	resp.Diagnostics.Append(r.read(ctx, a, false)...)
	resp.Diagnostics.Append(setRenderState(ctx, &resp.State, a)...)
}
func (r *RenderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	a, d := renderState(ctx, req.State)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	id := a["id"].(types.String).ValueString()
	var err error
	if r.kind == "template" {
		err = r.client.DeleteTemplateDefinition(ctx, id)
	} else {
		err = r.client.DeleteImageDefinition(ctx, id)
	}
	if err != nil && !missing(err) {
		resp.Diagnostics.AddError("Unable to delete rendering definition", apiErrorDetail(err, a["headers"], a["template_values"]))
	}
}
func (r *RenderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *RenderResource) fetch(ctx context.Context, id string) (*management.SavedRender, error) {
	if r.kind == "template" {
		return r.client.GetTemplateDefinition(ctx, id, nil)
	}
	return r.client.GetImageMetadata(ctx, id)
}
func (r *RenderResource) read(ctx context.Context, a map[string]attr.Value, imported bool) diag.Diagnostics {
	remote, err := r.fetch(ctx, a["id"].(types.String).ValueString())
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Definition saved, readback failed", apiErrorDetail(err, a["headers"], a["template_values"])+". Its ID has been saved in state; refresh before retrying.")
		return d
	}
	return r.merge(ctx, a, remote, imported)
}
func (r *RenderResource) merge(ctx context.Context, a map[string]attr.Value, v *management.SavedRender, imported bool) diag.Diagnostics {
	var d diag.Diagnostics
	if (r.kind == "template" && v.TemplateType != "html_css") || (r.kind != "template" && v.ImageType != r.kind) {
		d.AddError("Unsupported rendering definition", "The remote source type does not match this resource. Only HTML/CSS templates can be managed.")
		return d
	}
	values, ds := renderReadValues(ctx, r.kind, v.RenderDefinition)
	d.Append(ds...)
	for k, x := range values {
		// Format only selects the returned URL; it is not saved in image metadata.
		if k == "format" {
			continue
		}
		if k == "template_version" && r.kind == "templated" {
			// A failed create readback also lacks created_at, but already has
			// the caller's pin. Only initialize an absent/unknown import value.
			if imported && (a[k] == nil || a[k].IsUnknown()) {
				a[k] = types.Int64Null()
			}
			continue
		}
		if old := a[k]; !imported && old != nil && renderEquivalent(k, old, x) {
			continue
		}
		a[k] = x
	}
	a["created_at"] = types.StringValue(v.CreatedAt)
	if r.kind == "template" {
		a["version"] = types.Int64Value(v.Version)
		a["template_type"] = types.StringValue(v.TemplateType)
		a["updated_at"] = types.StringValue(v.UpdatedAt)
		return d
	}
	a["last_render_stored_at"] = types.StringPointerValue(v.LastRenderStoredAt)
	a["saved_to_storage_destination_at"] = types.StringPointerValue(v.SavedToStorageDestinationAt)
	a["og_config_id"] = types.StringPointerValue(v.OGConfigID)
	a["og_config_content_version"] = types.Int64PointerValue(v.OGConfigContentVersion)
	if r.kind == "templated" {
		a["resolved_template_version"] = types.Int64PointerValue(v.TemplateVersion)
	}
	a["storage_destination_hcti_storage_disabled"] = types.BoolPointerValue(v.StorageDestinationHCTIStorageDisabled)
	operation := ""
	if x, ok := a["image_url"].(types.String); ok && !x.IsNull() && !x.IsUnknown() {
		operation = x.ValueString()
	}
	storageOnly := false
	if operation != "" {
		u, err := url.Parse(operation)
		if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
			d.AddError("Invalid image operation URL", "The API returned an invalid rendering URL.")
			return d
		}
		storageOnly = strings.HasPrefix(u.Path, "/v1/store/")
	} else {
		storageOnly = v.StorageDestinationHCTIStorageDisabled != nil && *v.StorageDestinationHCTIStorageDisabled
		format := ""
		if x, ok := a["format"].(types.String); ok {
			format = x.ValueString()
		}
		var err error
		operation, err = r.client.ImageOperationURL(v.ID, format, storageOnly)
		if err != nil {
			d.AddError("Unable to construct rendering URL", fmt.Sprint(err))
			return d
		}
	}
	a["image_url"] = types.StringValue(operation)
	a["render_requires_auth"] = types.BoolValue(storageOnly)
	a["public_url"] = types.StringNull()
	a["render_method"] = types.StringValue("PUT")
	if !storageOnly {
		a["public_url"] = types.StringValue(operation)
		a["render_method"] = types.StringValue("GET")
	}
	return d
}
