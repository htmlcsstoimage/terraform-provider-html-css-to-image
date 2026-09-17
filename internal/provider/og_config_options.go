package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

// These explicit conversions keep requests typed and avoid JSON round trips.
func ogOptionsRequest(ctx context.Context, v types.Object) (*management.OGDefaultImageOptions, diag.Diagnostics) {
	var d diag.Diagnostics
	if v.IsNull() {
		return nil, d
	}
	a := v.Attributes()
	r := &management.OGDefaultImageOptions{}
	if !a["css"].IsNull() {
		r.CSS = hcti.Ptr(a["css"].(types.String).ValueString())
	}
	if !a["device_scale"].IsNull() {
		r.DeviceScale = hcti.Ptr(a["device_scale"].(types.Float64).ValueFloat64())
	}
	if !a["max_wait_ms"].IsNull() {
		r.MaxWaitMS = hcti.Ptr(int32(a["max_wait_ms"].(types.Int64).ValueInt64()))
	}
	if !a["ms_delay"].IsNull() {
		r.MSDelay = hcti.Ptr(int32(a["ms_delay"].(types.Int64).ValueInt64()))
	}
	if !a["render_when_ready"].IsNull() {
		r.RenderWhenReady = hcti.Ptr(a["render_when_ready"].(types.Bool).ValueBool())
	}
	if !a["selector"].IsNull() {
		r.Selector = hcti.Ptr(a["selector"].(types.String).ValueString())
	}
	if !a["viewport_height"].IsNull() {
		r.ViewportHeight = hcti.Ptr(int32(a["viewport_height"].(types.Int64).ValueInt64()))
	}
	if !a["viewport_width"].IsNull() {
		r.ViewportWidth = hcti.Ptr(int32(a["viewport_width"].(types.Int64).ValueInt64()))
	}
	if !a["disable_twemoji"].IsNull() {
		r.DisableTwemoji = hcti.Ptr(a["disable_twemoji"].(types.Bool).ValueBool())
	}
	if !a["color_scheme"].IsNull() {
		r.ColorScheme = hcti.Ptr(hcti.ColorScheme(a["color_scheme"].(types.String).ValueString()))
	}
	if !a["timezone"].IsNull() {
		r.Timezone = hcti.Ptr(a["timezone"].(types.String).ValueString())
	}
	if !a["block_consent_banners"].IsNull() {
		r.BlockConsentBanners = hcti.Ptr(a["block_consent_banners"].(types.Bool).ValueBool())
	}
	if !a["identify_as_hcti"].IsNull() {
		r.IdentifyAsHCTI = hcti.Ptr(a["identify_as_hcti"].(types.Bool).ValueBool())
	}
	if !a["headers"].IsNull() {
		d.Append(a["headers"].(types.Map).ElementsAs(ctx, &r.Headers, false)...)
	}
	if !a["additional_header_origins"].IsNull() {
		d.Append(a["additional_header_origins"].(types.Set).ElementsAs(ctx, &r.AdditionalHeaderOrigins, false)...)
	}
	if !a["include_headers_on_subrequests"].IsNull() {
		r.IncludeHeadersOnSubrequests = hcti.Ptr(a["include_headers_on_subrequests"].(types.Bool).ValueBool())
	}
	if !a["viewport_mobile"].IsNull() {
		r.ViewportMobile = hcti.Ptr(a["viewport_mobile"].(types.Bool).ValueBool())
	}
	if !a["viewport_landscape"].IsNull() {
		r.ViewportLandscape = hcti.Ptr(a["viewport_landscape"].(types.Bool).ValueBool())
	}
	if !a["viewport_touch"].IsNull() {
		r.ViewportTouch = hcti.Ptr(a["viewport_touch"].(types.Bool).ValueBool())
	}
	if !a["media_type"].IsNull() {
		r.MediaType = hcti.Ptr(hcti.MediaType(a["media_type"].(types.String).ValueString()))
	}
	if !a["proxy_id"].IsNull() {
		r.ProxyID = hcti.Ptr(a["proxy_id"].(types.String).ValueString())
	}
	if !a["storage_destination_id"].IsNull() {
		r.StorageDestinationID = hcti.Ptr(a["storage_destination_id"].(types.String).ValueString())
	}
	if !a["transparent_background"].IsNull() {
		r.TransparentBackground = hcti.Ptr(a["transparent_background"].(types.Bool).ValueBool())
	}
	return r, d
}

func ogOptionsRead(ctx context.Context, r *management.OGDefaultImageOptions) (types.Object, diag.Diagnostics) {
	var d diag.Diagnostics
	if r == nil {
		return types.ObjectNull(ogOptionTypes()), d
	}
	a := map[string]attr.Value{}
	a["css"] = types.StringNull()
	if r.CSS != nil {
		a["css"] = types.StringValue(*r.CSS)
	}
	a["device_scale"] = types.Float64Null()
	if r.DeviceScale != nil {
		a["device_scale"] = types.Float64Value(*r.DeviceScale)
	}
	a["max_wait_ms"] = types.Int64Null()
	if r.MaxWaitMS != nil {
		a["max_wait_ms"] = types.Int64Value(int64(*r.MaxWaitMS))
	}
	a["ms_delay"] = types.Int64Null()
	if r.MSDelay != nil {
		a["ms_delay"] = types.Int64Value(int64(*r.MSDelay))
	}
	a["render_when_ready"] = types.BoolNull()
	if r.RenderWhenReady != nil {
		a["render_when_ready"] = types.BoolValue(*r.RenderWhenReady)
	}
	a["selector"] = types.StringNull()
	if r.Selector != nil {
		a["selector"] = types.StringValue(*r.Selector)
	}
	a["viewport_height"] = types.Int64Null()
	if r.ViewportHeight != nil {
		a["viewport_height"] = types.Int64Value(int64(*r.ViewportHeight))
	}
	a["viewport_width"] = types.Int64Null()
	if r.ViewportWidth != nil {
		a["viewport_width"] = types.Int64Value(int64(*r.ViewportWidth))
	}
	a["disable_twemoji"] = types.BoolNull()
	if r.DisableTwemoji != nil {
		a["disable_twemoji"] = types.BoolValue(*r.DisableTwemoji)
	}
	a["color_scheme"] = types.StringNull()
	if r.ColorScheme != nil {
		a["color_scheme"] = types.StringValue(string(*r.ColorScheme))
	}
	a["timezone"] = types.StringNull()
	if r.Timezone != nil {
		a["timezone"] = types.StringValue(*r.Timezone)
	}
	a["block_consent_banners"] = types.BoolNull()
	if r.BlockConsentBanners != nil {
		a["block_consent_banners"] = types.BoolValue(*r.BlockConsentBanners)
	}
	a["identify_as_hcti"] = types.BoolNull()
	if r.IdentifyAsHCTI != nil {
		a["identify_as_hcti"] = types.BoolValue(*r.IdentifyAsHCTI)
	}
	vHeaders, dsHeaders := types.MapValueFrom(ctx, types.StringType, r.Headers)
	d.Append(dsHeaders...)
	a["headers"] = vHeaders
	vAdditionalHeaderOrigins, dsAdditionalHeaderOrigins := types.SetValueFrom(ctx, types.StringType, r.AdditionalHeaderOrigins)
	d.Append(dsAdditionalHeaderOrigins...)
	a["additional_header_origins"] = vAdditionalHeaderOrigins
	a["include_headers_on_subrequests"] = types.BoolNull()
	if r.IncludeHeadersOnSubrequests != nil {
		a["include_headers_on_subrequests"] = types.BoolValue(*r.IncludeHeadersOnSubrequests)
	}
	a["viewport_mobile"] = types.BoolNull()
	if r.ViewportMobile != nil {
		a["viewport_mobile"] = types.BoolValue(*r.ViewportMobile)
	}
	a["viewport_landscape"] = types.BoolNull()
	if r.ViewportLandscape != nil {
		a["viewport_landscape"] = types.BoolValue(*r.ViewportLandscape)
	}
	a["viewport_touch"] = types.BoolNull()
	if r.ViewportTouch != nil {
		a["viewport_touch"] = types.BoolValue(*r.ViewportTouch)
	}
	a["media_type"] = types.StringNull()
	if r.MediaType != nil {
		a["media_type"] = types.StringValue(string(*r.MediaType))
	}
	a["proxy_id"] = types.StringNull()
	if r.ProxyID != nil {
		a["proxy_id"] = types.StringValue(*r.ProxyID)
	}
	a["storage_destination_id"] = types.StringNull()
	if r.StorageDestinationID != nil {
		a["storage_destination_id"] = types.StringValue(*r.StorageDestinationID)
	}
	a["transparent_background"] = types.BoolNull()
	if r.TransparentBackground != nil {
		a["transparent_background"] = types.BoolValue(*r.TransparentBackground)
	}
	v, ds := types.ObjectValue(ogOptionTypes(), a)
	d.Append(ds...)
	return v, d
}
