package provider

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

func renderPDFRead(ctx context.Context, p *management.RenderPDFOptions) (types.Object, diag.Diagnostics) {
	t := map[string]attr.Type{}
	for k, s := range renderPDFAttributes() {
		t[k] = s.GetType()
	}
	a := map[string]attr.Value{"page_width": types.StringPointerValue(p.PageWidth), "page_height": types.StringPointerValue(p.PageHeight), "scale": types.Float64PointerValue(p.Scale), "print_background": types.BoolPointerValue(p.PrintBackground), "margins": nullValue(t["margins"])}
	var d diag.Diagnostics
	if len(p.Margins) != 0 {
		if len(p.Margins) != 4 {
			d.AddError("Invalid PDF metadata", "The API returned a margins array with a length other than four.")
		} else {
			m := map[string]attr.Value{}
			for i, k := range []string{"top", "right", "bottom", "left"} {
				m[k] = types.StringValue(p.Margins[i])
			}
			v, ds := types.ObjectValue(t["margins"].(types.ObjectType).AttrTypes, m)
			d.Append(ds...)
			a["margins"] = v
		}
	}
	v, ds := types.ObjectValue(t, a)
	d.Append(ds...)
	return v, d
}
func jsonTemplateValues(s string) (map[string]any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v map[string]any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return v, nil
}
func jsonEqual(a, b any) bool {
	switch x := a.(type) {
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		xx, ok := new(big.Rat).SetString(string(x))
		if !ok {
			return false
		}
		yy, ok := new(big.Rat).SetString(string(y))
		return ok && xx.Cmp(yy) == 0
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			w, ok := y[k]
			if !ok || !jsonEqual(v, w) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !jsonEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}
func renderEquivalent(k string, a, b attr.Value) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Equal(b) {
		return true
	}
	if !knownValue(a) || !knownValue(b) {
		return false
	}
	if k == "template_values" && !a.IsNull() && !b.IsNull() {
		x, e := jsonTemplateValues(a.(types.String).ValueString())
		y, f := jsonTemplateValues(b.(types.String).ValueString())
		return e == nil && f == nil && jsonEqual(x, y)
	}
	if k == "google_fonts" {
		norm := func(v attr.Value) map[string]bool {
			m := map[string]bool{}
			for _, x := range v.(types.Set).Elements() {
				s := strings.ReplaceAll(strings.TrimSpace(x.(types.String).ValueString()), "+", " ")
				if s != "" {
					m[s] = true
				}
			}
			return m
		}
		return reflect.DeepEqual(norm(a), norm(b))
	}
	if k == "additional_header_origins" {
		return ogEquivalent(k, a, b)
	}
	if k == "format" && !a.IsNull() && !b.IsNull() {
		return strings.ReplaceAll(a.(types.String).ValueString(), "jpeg", "jpg") == strings.ReplaceAll(b.(types.String).ValueString(), "jpeg", "jpg")
	}
	if k == "print_background" && a.IsNull() && !b.IsNull() && !b.(types.Bool).ValueBool() {
		return true
	}
	if !a.IsNull() && !b.IsNull() {
		if k == "scale" {
			return float32(a.(types.Float64).ValueFloat64()) == float32(b.(types.Float64).ValueFloat64())
		}
		switch k {
		case "page_width", "page_height", "top", "right", "bottom", "left":
			x, y := a.(types.String).ValueString(), b.(types.String).ValueString()
			if len(x) > 2 && len(y) > 2 && x[len(x)-2:] == y[len(y)-2:] {
				xx, e := strconv.ParseFloat(x[:len(x)-2], 32)
				yy, f := strconv.ParseFloat(y[:len(y)-2], 32)
				return e == nil && f == nil && xx == yy
			}
		}
	}
	if a.IsNull() || b.IsNull() {
		// The API normalizes empty optional strings and collections. No numeric defaults are guessed.
		switch k {
		case "css", "name", "description":
			return strings.TrimSpace(a.(types.String).ValueString()) == "" && strings.TrimSpace(b.(types.String).ValueString()) == ""
		case "headers", "metadata":
			return len(a.(types.Map).Elements()) == 0 && len(b.(types.Map).Elements()) == 0
		case "additional_header_origins":
			return len(a.(types.Set).Elements()) == 0 && len(b.(types.Set).Elements()) == 0
		}
		return false
	}
	if k == "pdf_options" || k == "margins" {
		x, y := a.(types.Object).Attributes(), b.(types.Object).Attributes()
		for n, v := range x {
			if !renderEquivalent(n, v, y[n]) {
				return false
			}
		}
		return true
	}
	return false
}

var pdfLengthPattern = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:px|in|cm|mm|pt)$`)

func validateRender(ctx context.Context, kind string, a map[string]attr.Value) diag.Diagnostics {
	var d diag.Diagnostics
	fail := func(k, msg string) { d.AddError("Invalid "+k, msg) }
	known := func(k string) bool { v, ok := a[k]; return ok && knownValue(v) && !v.IsNull() }
	text := func(k string) string {
		if !known(k) {
			return ""
		}
		return a[k].(types.String).ValueString()
	}
	for _, k := range []string{"html", "url", "template_id"} {
		if known(k) && strings.TrimSpace(text(k)) == "" {
			fail(k, "Must not be empty.")
		}
	}
	if known("template_id") && !strings.HasPrefix(text("template_id"), "t-") {
		fail("template_id", "Must include the t- prefix.")
	}
	if known("template_version") && a["template_version"].(types.Int64).ValueInt64() <= 0 {
		fail("template_version", "Must be a positive int64.")
	}
	if known("url") {
		u, e := url.Parse(text("url"))
		if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			fail("url", "Must be an absolute HTTP or HTTPS URL.")
		}
	}
	if known("template_values") {
		v, e := jsonTemplateValues(text("template_values"))
		if e != nil || len(v) == 0 {
			fail("template_values", "Must be a nonempty JSON object. Use jsonencode to preserve nested values.")
		}
	}
	if kind == "template" {
		if known("html") && !strings.Contains(text("html"), "{{") {
			fail("html", "Template HTML must contain a Handlebars placeholder.")
		}
		if known("css") && strings.Contains(text("css"), "{{") {
			fail("css", "Place Handlebars expressions in the HTML, not the CSS.")
		}
	}
	// API policy limits belong to the server; only reject negative numbers here.
	for _, k := range []string{"viewport_width", "viewport_height", "max_wait_ms", "ms_delay", "jumbo_max_width", "jumbo_max_height"} {
		if known(k) {
			v := a[k].(types.Int64).ValueInt64()
			if v < 0 {
				fail(k, "Must be nonnegative.")
			}
		}
	}
	if known("device_scale") {
		v := a["device_scale"].(types.Float64).ValueFloat64()
		if v < 0 {
			fail("device_scale", "Must be nonnegative.")
		}
	}
	for _, pair := range [][2]string{{"viewport_width", "viewport_height"}, {"jumbo_max_width", "jumbo_max_height"}} {
		x, xok := a[pair[0]]
		y, yok := a[pair[1]]
		if xok && yok && knownValue(x) && knownValue(y) && x.IsNull() != y.IsNull() {
			fail(pair[0], "Both dimensions must be supplied together.")
		}
	}
	if h, ok := a["headers"]; ok && knownValue(h) && knownValue(a["additional_header_origins"]) {
		d.Append(ogValidateHeaders(ctx, h.(types.Map), a["additional_header_origins"].(types.Set), "")...)
		if known("include_headers_on_subrequests") && a["include_headers_on_subrequests"].(types.Bool).ValueBool() && len(h.(types.Map).Elements()) == 0 {
			fail("include_headers_on_subrequests", "Requires at least one header.")
		}
	}
	if known("google_fonts") {
		for _, v := range a["google_fonts"].(types.Set).Elements() {
			s := strings.TrimSpace(v.(types.String).ValueString())
			if s == "" || strings.Contains(s, "|") {
				fail("google_fonts", "Each element must be one nonempty font family without a pipe separator.")
			}
		}
	}
	if known("pdf_options") {
		p := a["pdf_options"].(types.Object).Attributes()
		for _, k := range []string{"page_width", "page_height"} {
			if !p[k].IsNull() && !pdfLengthPattern.MatchString(p[k].(types.String).ValueString()) {
				fail("pdf_options."+k, "Supply a nonnegative length with px, in, cm, mm, or pt units.")
			}
		}
		if !p["scale"].IsNull() {
			s := p["scale"].(types.Float64).ValueFloat64()
			if s < 0 {
				fail("pdf_options.scale", "Must be nonnegative.")
			}
		}
		if !p["margins"].IsNull() {
			for k, v := range p["margins"].(types.Object).Attributes() {
				if !pdfLengthPattern.MatchString(v.(types.String).ValueString()) {
					fail("pdf_options.margins."+k, "Supply a nonnegative length with supported units.")
				}
			}
		}
	}
	return d
}
