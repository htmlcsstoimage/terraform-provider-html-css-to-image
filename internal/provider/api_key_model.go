package provider

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type apiKeyModel struct {
	ID                   types.String `tfsdk:"id"`
	APIID                types.String `tfsdk:"api_id"`
	APIKey               types.String `tfsdk:"api_key"`
	Name                 types.String `tfsdk:"name"`
	EffectiveName        types.String `tfsdk:"effective_name"`
	Description          types.String `tfsdk:"description"`
	Disabled             types.Bool   `tfsdk:"disabled"`
	AllFuturePermissions types.Bool   `tfsdk:"all_future_permissions"`
	Permissions          types.Set    `tfsdk:"permissions"`
	EffectivePermissions types.Set    `tfsdk:"effective_permissions"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

func (m apiKeyModel) known() bool {
	if m.Name.IsUnknown() || m.Description.IsUnknown() || m.Disabled.IsUnknown() || m.AllFuturePermissions.IsUnknown() || m.Permissions.IsUnknown() {
		return false
	}
	for _, v := range m.Permissions.Elements() {
		if v.IsUnknown() {
			return false
		}
	}
	return true
}
func (m apiKeyModel) request(ctx context.Context) (*management.APIKeyRequest, diag.Diagnostics) {
	var d diag.Diagnostics
	r := &management.APIKeyRequest{Disabled: m.Disabled.ValueBool(), AllFuturePermissions: m.AllFuturePermissions.ValueBool(), Permissions: []management.Permission{}}
	if !m.known() {
		d.AddError("Unknown API key settings", "API key settings must be known during apply.")
		return r, d
	}
	if !m.Name.IsNull() {
		r.Name = hcti.Ptr(m.Name.ValueString())
		if utf8.RuneCountInString(*r.Name) > 255 {
			d.AddError("Invalid key name", "name must not exceed 255 characters.")
		}
	}
	if !m.Description.IsNull() {
		r.Description = hcti.Ptr(m.Description.ValueString())
		if utf8.RuneCountInString(*r.Description) > 2000 {
			d.AddError("Invalid description", "description must not exceed 2000 characters.")
		}
	}
	var grants []string
	if !m.Permissions.IsNull() {
		d.Append(m.Permissions.ElementsAs(ctx, &grants, false)...)
	}
	if r.AllFuturePermissions {
		if len(grants) > 0 {
			d.AddError("Conflicting permissions", "Omit permissions or supply [] when all_future_permissions is true.")
		}
	} else {
		if m.Permissions.IsNull() {
			d.AddError("Permissions required", "Supply permissions explicitly, including [] to grant no permissions.")
		}
		for _, v := range grants {
			if strings.TrimSpace(v) == "" {
				d.AddError("Invalid permission", "Permission names cannot be blank.")
			}
			r.Permissions = append(r.Permissions, management.Permission(v))
		}
	}
	return r, d
}
func (m *apiKeyModel) read(ctx context.Context, k *management.APIKey, importing bool) diag.Diagnostics {
	var d diag.Diagnostics
	m.ID = types.StringValue(k.ID)
	m.APIID = types.StringValue(k.APIID)
	m.EffectiveName = types.StringValue(k.Name)
	if importing || (!m.Name.IsNull() && strings.TrimSpace(m.Name.ValueString()) != "" && strings.TrimSpace(m.Name.ValueString()) != k.Name) {
		m.Name = types.StringValue(k.Name)
	}
	remoteDescription := ""
	if k.Description != nil {
		remoteDescription = *k.Description
	}
	if importing || strings.TrimSpace(m.Description.ValueString()) != remoteDescription {
		m.Description = types.StringNull()
		if k.Description != nil {
			m.Description = types.StringValue(*k.Description)
		}
	}
	m.Disabled = types.BoolValue(!k.Enabled)
	m.AllFuturePermissions = types.BoolValue(k.AllFuturePermissions)
	grants := make([]string, len(k.Permissions))
	for i, v := range k.Permissions {
		grants[i] = string(v)
	}
	set, ds := types.SetValueFrom(ctx, types.StringType, grants)
	d.Append(ds...)
	m.EffectivePermissions = set
	if !k.AllFuturePermissions {
		m.Permissions = set
	} else if importing || len(m.Permissions.Elements()) > 0 {
		m.Permissions = types.SetNull(types.StringType)
	}
	if importing || m.APIKey.IsUnknown() {
		m.APIKey = types.StringNull()
	}
	m.CreatedAt = types.StringValue(k.CreatedAt.UTC().Format(time.RFC3339Nano))
	m.UpdatedAt = types.StringValue(k.UpdatedAt.UTC().Format(time.RFC3339Nano))
	return d
}
