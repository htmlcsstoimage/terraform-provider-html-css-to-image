package provider

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Match the API's template field path normalization, including array indices.
func ogTemplatePath(s string) string {
	original := s
	s = strings.TrimSpace(s)
	var out strings.Builder
	expect := true
	for s != "" {
		s = strings.TrimLeftFunc(s, unicode.IsSpace)
		if s == "" {
			break
		}
		switch s[0] {
		case '.':
			if expect {
				return original
			}
			s = s[1:]
			expect = true
		case '[':
			if out.Len() == 0 {
				return original
			}
			end := strings.IndexByte(s, ']')
			if end < 0 {
				return original
			}
			digits := strings.TrimSpace(s[1:end])
			if digits == "" || strings.ContainsFunc(digits, func(r rune) bool { return r < '0' || r > '9' }) {
				return original
			}
			n, e := strconv.ParseUint(digits, 10, 31)
			if e != nil {
				return original
			}
			out.WriteString("[" + strconv.FormatUint(n, 10) + "]")
			s = s[end+1:]
			expect = false
		default:
			end := strings.IndexAny(s, ".[")
			if end < 0 {
				end = len(s)
			}
			part := strings.TrimSpace(s[:end])
			if part == "" {
				return original
			}
			for i, r := range part {
				if r > 0xffff || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || (i > 0 && r == '-')) {
					return original
				}
			}
			if out.Len() > 0 {
				out.WriteByte('.')
			}
			out.WriteString(part)
			s = s[end:]
			expect = false
		}
	}
	if expect {
		return original
	}
	return out.String()
}
func ogOrigin(s string) string {
	u, e := url.Parse(strings.TrimSpace(s))
	if e != nil || u.Host == "" {
		return s
	}
	host := strings.ToLower(u.Host)
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		host = strings.TrimSuffix(host, ":"+u.Port())
	}
	return strings.ToLower(u.Scheme) + "://" + host
}
func ogEquivalent(name string, a, b attr.Value) bool {
	if a.Equal(b) {
		return true
	}
	if !knownValue(a) || !knownValue(b) {
		return false
	}
	if a.IsNull() {
		switch name {
		case "extract_values", "include_headers_on_subrequests":
			return !b.IsNull() && !b.(types.Bool).ValueBool()
		case "optimization_mode":
			return !b.IsNull() && b.(types.String).ValueString() == "post_process"
		case "refresh_interval_s":
			return !b.IsNull() && b.(types.Int64).ValueInt64() == 86400
		}
	}
	switch a := a.(type) {
	case types.String:
		other := b.(types.String)
		switch name {
		case "name", "description", "template_id", "proxy_id", "storage_destination_id", "meta_key":
			return ogTrim(a) == ogTrim(other)
		case "css":
			return (ogTrim(a) == "" && ogTrim(other) == "") || a.Equal(other)
		case "base_url":
			return !a.IsNull() && !other.IsNull() && ogOrigin(a.ValueString()) == ogOrigin(other.ValueString())
		case "template_key":
			return ogTemplatePath(a.ValueString()) == ogTemplatePath(other.ValueString())
		}
	case types.Map:
		return len(a.Elements()) == 0 && len(b.(types.Map).Elements()) == 0
	case types.Set:
		other := b.(types.Set)
		if name == "additional_header_origins" {
			set := func(s types.Set) map[string]bool {
				r := map[string]bool{}
				for _, v := range s.Elements() {
					r[ogOrigin(v.(types.String).ValueString())] = true
				}
				return r
			}
			x, y := set(a), set(other)
			if len(x) != len(y) {
				return false
			}
			for k := range x {
				if !y[k] {
					return false
				}
			}
			return true
		}
	case types.List:
		x, y := a.Elements(), b.(types.List).Elements()
		if len(x) != len(y) {
			return false
		}
		for i := range x {
			if !ogEquivalent("mapping", x[i], y[i]) {
				return false
			}
		}
		return true
	case types.Object:
		other := b.(types.Object)
		// The API returns an empty options object even if default_options was null,
		// and represents include_headers_on_subrequests=null as false.
		x, y := a.Attributes(), other.Attributes()
		for k, t := range a.AttributeTypes(context.Background()) { // Typed null objects still carry their shape.
			av, bv := x[k], y[k]
			if av == nil {
				av = nullValue(t)
			}
			if bv == nil {
				bv = nullValue(t)
			}
			if !ogEquivalent(k, av, bv) {
				return false
			}
		}
		return true
	}
	return false
}
