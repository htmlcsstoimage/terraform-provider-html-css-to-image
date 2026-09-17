package provider

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
	"sort"
	"strings"
)

func renderRequest(ctx context.Context, a map[string]attr.Value) (*management.RenderDefinition, diag.Diagnostics) {
	r := &management.RenderDefinition{}
	var d diag.Diagnostics
	if v, ok := a["html"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.HTML = &x
	}
	if v, ok := a["name"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.Name = &x
	}
	if v, ok := a["description"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.Description = &x
	}
	if v, ok := a["css"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.CSS = &x
	}
	if v, ok := a["device_scale"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Float64).ValueFloat64()
		r.DeviceScale = &x
	}
	if v, ok := a["google_fonts"]; ok && !v.IsNull() && !v.IsUnknown() {
		var fonts []string
		d.Append(v.(types.Set).ElementsAs(ctx, &fonts, false)...)
		for i := range fonts {
			fonts[i] = strings.ReplaceAll(strings.TrimSpace(fonts[i]), " ", "+")
		}
		sort.Strings(fonts)
		unique := fonts[:0]
		for _, font := range fonts {
			if len(unique) == 0 || unique[len(unique)-1] != font {
				unique = append(unique, font)
			}
		}
		s := strings.Join(unique, "|")
		r.GoogleFonts = &s
	}
	if v, ok := a["max_wait_ms"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.MaxWaitMS = &x
	}
	if v, ok := a["ms_delay"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.MSDelay = &x
	}
	if v, ok := a["render_when_ready"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.RenderWhenReady = &x
	}
	if v, ok := a["selector"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.Selector = &x
	}
	if v, ok := a["viewport_height"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.ViewportHeight = &x
	}
	if v, ok := a["viewport_width"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.ViewportWidth = &x
	}
	if v, ok := a["disable_twemoji"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.DisableTwemoji = &x
	}
	if v, ok := a["color_scheme"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.ColorScheme = &x
	}
	if v, ok := a["timezone"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.Timezone = &x
	}
	if v, ok := a["viewport_mobile"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.ViewportMobile = &x
	}
	if v, ok := a["viewport_landscape"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.ViewportLandscape = &x
	}
	if v, ok := a["viewport_touch"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.ViewportTouch = &x
	}
	if v, ok := a["media_type"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.MediaType = &x
	}
	if v, ok := a["proxy_id"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.ProxyID = &x
	}
	if v, ok := a["storage_destination_id"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.StorageDestinationID = &x
	}
	if v, ok := a["jumbo_max_height"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.JumboMaxHeight = &x
	}
	if v, ok := a["jumbo_max_width"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.JumboMaxWidth = &x
	}
	if v, ok := a["transparent_background"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.TransparentBackground = &x
	}
	if v, ok := a["metadata"]; ok && !v.IsNull() && !v.IsUnknown() {
		d.Append(v.(types.Map).ElementsAs(ctx, &r.Metadata, false)...)
	}
	if v, ok := a["max_render_once"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.MaxRenderOnce = &x
	}
	if v, ok := a["format"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.Format = &x
	}
	if v, ok := a["url"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.URL = &x
	}
	if v, ok := a["full_screen"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.FullScreen = &x
	}
	if v, ok := a["block_consent_banners"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.BlockConsentBanners = &x
	}
	if v, ok := a["identify_as_hcti"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.IdentifyAsHCTI = &x
	}
	if v, ok := a["headers"]; ok && !v.IsNull() && !v.IsUnknown() {
		d.Append(v.(types.Map).ElementsAs(ctx, &r.Headers, false)...)
	}
	if v, ok := a["additional_header_origins"]; ok && !v.IsNull() && !v.IsUnknown() {
		d.Append(v.(types.Set).ElementsAs(ctx, &r.AdditionalHeaderOrigins, false)...)
	}
	if v, ok := a["include_headers_on_subrequests"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Bool).ValueBool()
		r.IncludeHeadersOnSubrequests = &x
	}
	if v, ok := a["template_values"]; ok && !v.IsNull() && !v.IsUnknown() {
		r.TemplateValues = json.RawMessage(v.(types.String).ValueString())
	}
	if v, ok := a["template_id"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.String).ValueString()
		r.TemplateID = &x
	}
	if v, ok := a["template_version"]; ok && !v.IsNull() && !v.IsUnknown() {
		x := v.(types.Int64).ValueInt64()
		r.TemplateVersion = &x
	}
	if v, ok := a["pdf_options"]; ok && !v.IsNull() && !v.IsUnknown() {
		p := v.(types.Object).Attributes()
		r.PDFOptions = &management.RenderPDFOptions{}
		if !p["page_width"].IsNull() {
			x := p["page_width"].(types.String).ValueString()
			r.PDFOptions.PageWidth = &x
		}
		if !p["page_height"].IsNull() {
			x := p["page_height"].(types.String).ValueString()
			r.PDFOptions.PageHeight = &x
		}
		if !p["scale"].IsNull() {
			x := p["scale"].(types.Float64).ValueFloat64()
			r.PDFOptions.Scale = &x
		}
		if !p["print_background"].IsNull() {
			x := p["print_background"].(types.Bool).ValueBool()
			r.PDFOptions.PrintBackground = &x
		}
		if !p["margins"].IsNull() {
			m := p["margins"].(types.Object).Attributes()
			for _, k := range []string{"top", "right", "bottom", "left"} {
				r.PDFOptions.Margins = append(r.PDFOptions.Margins, m[k].(types.String).ValueString())
			}
		}
	}
	return r, d
}
func renderReadValues(ctx context.Context, kind string, r management.RenderDefinition) (map[string]attr.Value, diag.Diagnostics) {
	a := map[string]attr.Value{}
	var d diag.Diagnostics
	for k, s := range renderAttributes(kind) {
		a[k] = nullValue(s.GetType())
		switch k {
		case "html":
			if r.HTML != nil {
				a[k] = types.StringValue(*r.HTML)
			}
		case "name":
			if r.Name != nil {
				a[k] = types.StringValue(*r.Name)
			}
		case "description":
			if r.Description != nil {
				a[k] = types.StringValue(*r.Description)
			}
		case "css":
			if r.CSS != nil {
				a[k] = types.StringValue(*r.CSS)
			}
		case "device_scale":
			if r.DeviceScale != nil {
				a[k] = types.Float64Value(*r.DeviceScale)
			}
		case "google_fonts":
			if r.GoogleFonts != nil {
				fonts := []string{}
				for _, v := range strings.Split(*r.GoogleFonts, "|") {
					if v = strings.TrimSpace(v); v != "" {
						fonts = append(fonts, strings.ReplaceAll(v, "+", " "))
					}
				}
				x, ds := types.SetValueFrom(ctx, types.StringType, fonts)
				d.Append(ds...)
				a[k] = x
			}
		case "max_wait_ms":
			if r.MaxWaitMS != nil {
				a[k] = types.Int64Value(*r.MaxWaitMS)
			}
		case "ms_delay":
			if r.MSDelay != nil {
				a[k] = types.Int64Value(*r.MSDelay)
			}
		case "render_when_ready":
			if r.RenderWhenReady != nil {
				a[k] = types.BoolValue(*r.RenderWhenReady)
			}
		case "selector":
			if r.Selector != nil {
				a[k] = types.StringValue(*r.Selector)
			}
		case "viewport_height":
			if r.ViewportHeight != nil {
				a[k] = types.Int64Value(*r.ViewportHeight)
			}
		case "viewport_width":
			if r.ViewportWidth != nil {
				a[k] = types.Int64Value(*r.ViewportWidth)
			}
		case "disable_twemoji":
			if r.DisableTwemoji != nil {
				a[k] = types.BoolValue(*r.DisableTwemoji)
			}
		case "color_scheme":
			if r.ColorScheme != nil {
				a[k] = types.StringValue(*r.ColorScheme)
			}
		case "timezone":
			if r.Timezone != nil {
				a[k] = types.StringValue(*r.Timezone)
			}
		case "viewport_mobile":
			if r.ViewportMobile != nil {
				a[k] = types.BoolValue(*r.ViewportMobile)
			}
		case "viewport_landscape":
			if r.ViewportLandscape != nil {
				a[k] = types.BoolValue(*r.ViewportLandscape)
			}
		case "viewport_touch":
			if r.ViewportTouch != nil {
				a[k] = types.BoolValue(*r.ViewportTouch)
			}
		case "media_type":
			if r.MediaType != nil {
				a[k] = types.StringValue(*r.MediaType)
			}
		case "proxy_id":
			if r.ProxyID != nil {
				a[k] = types.StringValue(*r.ProxyID)
			}
		case "storage_destination_id":
			if r.StorageDestinationID != nil {
				a[k] = types.StringValue(*r.StorageDestinationID)
			}
		case "jumbo_max_height":
			if r.JumboMaxHeight != nil {
				a[k] = types.Int64Value(*r.JumboMaxHeight)
			}
		case "jumbo_max_width":
			if r.JumboMaxWidth != nil {
				a[k] = types.Int64Value(*r.JumboMaxWidth)
			}
		case "transparent_background":
			if r.TransparentBackground != nil {
				a[k] = types.BoolValue(*r.TransparentBackground)
			}
		case "metadata":
			x, ds := types.MapValueFrom(ctx, types.StringType, r.Metadata)
			d.Append(ds...)
			a[k] = x
		case "max_render_once":
			if r.MaxRenderOnce != nil {
				a[k] = types.BoolValue(*r.MaxRenderOnce)
			}
		case "format":
			if r.Format != nil {
				a[k] = types.StringValue(*r.Format)
			}
		case "url":
			if r.URL != nil {
				a[k] = types.StringValue(*r.URL)
			}
		case "full_screen":
			if r.FullScreen != nil {
				a[k] = types.BoolValue(*r.FullScreen)
			}
		case "block_consent_banners":
			if r.BlockConsentBanners != nil {
				a[k] = types.BoolValue(*r.BlockConsentBanners)
			}
		case "identify_as_hcti":
			if r.IdentifyAsHCTI != nil {
				a[k] = types.BoolValue(*r.IdentifyAsHCTI)
			}
		case "headers":
			x, ds := types.MapValueFrom(ctx, types.StringType, r.Headers)
			d.Append(ds...)
			a[k] = x
		case "additional_header_origins":
			x, ds := types.SetValueFrom(ctx, types.StringType, r.AdditionalHeaderOrigins)
			d.Append(ds...)
			a[k] = x
		case "include_headers_on_subrequests":
			if r.IncludeHeadersOnSubrequests != nil {
				a[k] = types.BoolValue(*r.IncludeHeadersOnSubrequests)
			}
		case "template_values":
			if len(r.TemplateValues) > 0 {
				a[k] = types.StringValue(string(r.TemplateValues))
			}
		case "template_id":
			if r.TemplateID != nil {
				a[k] = types.StringValue(*r.TemplateID)
			}
		case "template_version":
			if r.TemplateVersion != nil {
				a[k] = types.Int64Value(*r.TemplateVersion)
			}
		case "pdf_options":
			if r.PDFOptions != nil {
				x, ds := renderPDFRead(ctx, r.PDFOptions)
				d.Append(ds...)
				a[k] = x
			}
		}
	}
	return a, d
}
