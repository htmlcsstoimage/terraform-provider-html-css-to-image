package provider

import (
	"context"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func knownValue(v attr.Value) bool {
	if v.IsUnknown() {
		return false
	}
	if v.IsNull() {
		return true
	}
	switch v := v.(type) {
	case types.Object:
		for _, e := range v.Attributes() {
			if !knownValue(e) {
				return false
			}
		}
	case types.List:
		for _, e := range v.Elements() {
			if !knownValue(e) {
				return false
			}
		}
	case types.Set:
		for _, e := range v.Elements() {
			if !knownValue(e) {
				return false
			}
		}
	case types.Map:
		for _, e := range v.Elements() {
			if !knownValue(e) {
				return false
			}
		}
	}
	return true
}

func nullValue(t attr.Type) attr.Value {
	ctx := context.Background()
	v, err := t.ValueFromTerraform(ctx, tftypes.NewValue(t.TerraformType(ctx), nil))
	if err != nil {
		panic(err)
	} // Built-in schema types always accept null.
	return v
}

func validOrigin(s string, httpsOnly bool) bool {
	u, e := url.Parse(s)
	return e == nil && u.Hostname() != "" && (u.Scheme == "https" || (!httpsOnly && u.Scheme == "http")) && u.User == nil && (u.Path == "" || u.Path == "/") && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && !strings.Contains(s, "#") && u.Opaque == ""
}
