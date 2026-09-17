package provider

import (
	"context"
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m ogConfigModel) validate(ctx context.Context) diag.Diagnostics {
	var d diag.Diagnostics
	if !m.known() {
		d.AddError("Unknown OG settings", "OG settings must be known during apply.")
		return d
	}
	invalid := func(field, rule string) { d.AddError("Invalid OG configuration", field+" "+rule) }
	if ogTrim(m.Name) == "" {
		invalid("name", "must not be empty after trimming.")
	}
	if !validOrigin(strings.TrimSpace(m.BaseURL.ValueString()), true) {
		invalid("base_url", "must be an HTTPS origin without credentials, path, query, or fragment.")
	}
	if !m.RefreshInterval.IsNull() && m.RefreshInterval.ValueInt64() < 0 {
		invalid("refresh_interval_s", "must be nonnegative.")
	}
	switch m.ConfigType.ValueString() {
	case "html_css":
		for name, v := range m.inputs() {
			switch name {
			case "template_id", "template_version", "template_values_mapping", "headers", "additional_header_origins":
				if !v.IsNull() {
					invalid(name, "is only valid for templated configurations.")
				}
			}
		}
		if !m.DefaultOptions.IsNull() {
			a := m.DefaultOptions.Attributes()
			for _, name := range []string{"max_wait_ms", "ms_delay", "viewport_width", "viewport_height"} {
				v := a[name].(types.Int64)
				if !v.IsNull() && v.ValueInt64() < 0 {
					invalid("default_options."+name, "must be nonnegative.")
				}
				// The SDK uses int32 for these fields. Reject overflow before conversion.
				if !v.IsNull() && v.ValueInt64() > math.MaxInt32 {
					invalid("default_options."+name, "cannot be represented by the SDK's 32-bit integer type.")
				}
			}
			scale := a["device_scale"].(types.Float64)
			if !scale.IsNull() && (math.IsNaN(scale.ValueFloat64()) || math.IsInf(scale.ValueFloat64(), 0) || scale.ValueFloat64() < 0) {
				invalid("default_options.device_scale", "must be finite and nonnegative.")
			}
			if a["viewport_height"].IsNull() != a["viewport_width"].IsNull() {
				invalid("default_options", "must supply both viewport_width and viewport_height together.")
			}
			headers := a["headers"].(types.Map)
			origins := a["additional_header_origins"].(types.Set)
			d.Append(ogValidateHeaders(ctx, headers, origins, "default_options.")...)
			if a["include_headers_on_subrequests"].(types.Bool).ValueBool() && len(headers.Elements()) == 0 {
				invalid("default_options.include_headers_on_subrequests", "requires at least one header.")
			}
		}
	case "templated":
		if !m.DefaultOptions.IsNull() || !m.ExtractValues.IsNull() {
			invalid("default_options and extract_values", "are only valid for html_css configurations.")
		}
		if m.TemplateID.IsNull() || ogTrim(m.TemplateID) == "" {
			invalid("template_id", "must identify a template.")
		}
		if !m.TemplateVersion.IsNull() && m.TemplateVersion.ValueInt64() <= 0 {
			invalid("template_version", "must be positive or omitted to follow the latest version.")
		}
		keys := map[string]bool{}
		for i, v := range m.TemplateValuesMapping.Elements() {
			path := fmt.Sprintf("template_values_mapping[%d]", i)
			if v.IsNull() {
				invalid(path, "cannot be null.")
				continue
			}
			a := v.(types.Object).Attributes()
			key := ogTemplatePath(a["template_key"].(types.String).ValueString())
			meta := a["meta_key"].(types.String)
			fallback := a["fallback"].(types.String)
			if key == "" {
				invalid(path+".template_key", "must not be empty.")
			}
			for existing := range keys {
				if existing == key || strings.HasPrefix(key, existing+".") || strings.HasPrefix(existing, key+".") || strings.HasPrefix(key, existing+"[") || strings.HasPrefix(existing, key+"[") {
					invalid(path+".template_key", "must not duplicate or overlap another mapping.")
				}
			}
			keys[key] = true
			if meta.IsNull() == fallback.IsNull() {
				invalid(path, "must supply exactly one of meta_key or fallback.")
			}
			if !meta.IsNull() && (ogTrim(meta) == "" || strings.ContainsFunc(meta.ValueString(), unicode.IsControl)) {
				invalid(path+".meta_key", "must be nonempty without control characters.")
			}
			if !fallback.IsNull() && fallback.ValueString() != "titles" && fallback.ValueString() != "descriptions" {
				invalid(path+".fallback", "must be titles or descriptions.")
			}
		}
		d.Append(ogValidateHeaders(ctx, m.Headers, m.AdditionalHeaderOrigins, "")...)
	default:
		invalid("config_type", "must be html_css or templated.")
	}
	return d
}

func ogValidateHeaders(ctx context.Context, headers types.Map, origins types.Set, prefix string) diag.Diagnostics {
	var d diag.Diagnostics
	if len(origins.Elements()) > 0 && len(headers.Elements()) == 0 {
		d.AddError("Headers required", prefix+"additional_header_origins requires at least one header.")
	}
	seen := map[string]bool{}
	for k, v := range headers.Elements() {
		if seen[strings.ToLower(k)] {
			d.AddError("Invalid headers", prefix+"headers contains duplicate case-insensitive names.")
		}
		seen[strings.ToLower(k)] = true
		// Diagnostics deliberately omit both names and values from sensitive maps.
		if len(k) == 0 || !ogHeaderName(k) || v.IsNull() || !utf8.ValidString(v.(types.String).ValueString()) || strings.ContainsFunc(v.(types.String).ValueString(), func(r rune) bool { return unicode.IsControl(r) && r != '\t' }) {
			d.AddError("Invalid headers", prefix+"headers requires nonempty ASCII names and non-null UTF-8 values, without line breaks.")
		}
	}
	for _, v := range origins.Elements() {
		if v.IsNull() || !validOrigin(v.(types.String).ValueString(), false) {
			d.AddError("Invalid header origin", prefix+"additional_header_origins must contain exact HTTP(S) origins.")
		}
	}
	return d
}

func ogHeaderName(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", r)) {
			return false
		}
	}
	return s != ""
}
