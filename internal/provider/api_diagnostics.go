package provider

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

// apiErrorDetail exposes structured API failures to IaC users. It excludes raw
// response bodies and headers and redacts known sensitive request/state values.
func apiErrorDetail(err error, sensitive ...attr.Value) string {
	var secrets []string
	var collect func(attr.Value)
	collect = func(v attr.Value) {
		if v == nil || v.IsNull() || v.IsUnknown() {
			return
		}
		switch v := v.(type) {
		case types.String:
			if s := v.ValueString(); s != "" {
				secrets = append(secrets, s)
			}
		case types.Map:
			for _, x := range v.Elements() {
				collect(x)
			}
		case types.Object:
			for name, x := range v.Attributes() {
				switch name {
				case "password", "secret_access_key", "access_key_id", "headers":
					collect(x)
				default:
					if _, ok := x.(types.Object); ok {
						collect(x)
					}
				}
			}
		}
	}
	for _, v := range sensitive {
		collect(v)
	}
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	redact := func(value string) string {
		for _, secret := range secrets {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
		return value
	}
	var api *management.APIError
	if !errors.As(err, &api) {
		return redact(err.Error())
	}
	var b strings.Builder
	fmt.Fprintf(&b, "HCTI API request failed (HTTP %d)", api.StatusCode)
	if api.Code != "" {
		fmt.Fprintf(&b, ": %s", redact(api.Code))
	}
	if api.Message != "" {
		fmt.Fprintf(&b, "\n\n%s", redact(api.Message))
	}
	for _, v := range api.ValidationErrors {
		// Paths identify fields, not submitted values. Keep them readable even
		// when a credential happens to overlap a field name.
		if v.Path != "" {
			fmt.Fprintf(&b, "\n- %s: %s", v.Path, redact(v.Message))
		} else {
			fmt.Fprintf(&b, "\n- %s", redact(v.Message))
		}
	}
	return b.String()
}
