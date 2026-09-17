package provider

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type proxyModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	URL            types.String `tfsdk:"url"`
	Port           types.Int64  `tfsdk:"port"`
	EffectivePort  types.Int64  `tfsdk:"effective_port"`
	Disabled       types.Bool   `tfsdk:"disabled"`
	BypassHosts    types.Set    `tfsdk:"bypass_hosts"`
	Authentication types.Object `tfsdk:"authentication"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

var authenticationTypes = map[string]attr.Type{"username": types.StringType, "password": types.StringType}

func authentication(v types.Object) (types.String, types.String) {
	if v.IsNull() || v.IsUnknown() {
		return types.StringNull(), types.StringNull()
	}
	a := v.Attributes()
	return a["username"].(types.String), a["password"].(types.String)
}
func normalizeHosts(hosts []string) []string {
	unique := map[string]bool{}
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if u, e := url.Parse(h); e == nil && u.IsAbs() && u.Hostname() != "" {
			h = u.Hostname()
		}
		if h != "" {
			unique[strings.ToLower(h)] = true
		}
	}
	out := make([]string, 0, len(unique))
	for h := range unique {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}
func (m proxyModel) request(ctx context.Context, old *proxyModel) (*management.ProxyRequest, diag.Diagnostics) {
	var d diag.Diagnostics
	req := &management.ProxyRequest{Name: m.Name.ValueString(), URL: m.URL.ValueString(), Disabled: m.Disabled.ValueBool()}
	if m.Name.IsUnknown() || m.URL.IsUnknown() || m.Port.IsUnknown() || m.Disabled.IsUnknown() || m.Authentication.IsUnknown() || m.BypassHosts.IsUnknown() {
		d.AddError("Unknown proxy settings", "Proxy settings must be known during apply.")
		return req, d
	}
	if strings.TrimSpace(req.Name) == "" {
		d.AddError("Invalid proxy name", "name must not be empty after trimming.")
	}
	u, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Port() != "" || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		d.AddError("Invalid proxy URL", "url must be an HTTP(S) origin without credentials, a port, query, fragment, or non-root path.")
	}
	if !m.Port.IsNull() {
		n := m.Port.ValueInt64()
		if n < 1 || n > 65535 {
			d.AddError("Invalid proxy port", "port must be between 1 and 65535.")
		} else {
			req.Port = hcti.Ptr(uint16(n))
		}
	}
	if !m.BypassHosts.IsNull() {
		d.Append(m.BypassHosts.ElementsAs(ctx, &req.BypassHosts, false)...)
	}
	req.BypassHosts = normalizeHosts(req.BypassHosts)
	if !m.Authentication.IsNull() {
		username, password := authentication(m.Authentication)
		if username.IsUnknown() || password.IsUnknown() {
			d.AddError("Unknown credentials", "Credentials must be known during apply.")
			return req, d
		}
		if username.IsNull() {
			d.AddError("Invalid proxy username", "Supply a username. Empty is valid.")
		}
		req.Authentication = &management.ProxyAuthentication{Username: username.ValueString()}
		if !password.IsNull() {
			req.Authentication.Password = hcti.Ptr(password.ValueString())
		} else {
			oldUsername := types.StringNull()
			if old != nil {
				oldUsername, _ = authentication(old.Authentication)
			}
			if oldUsername.IsNull() || !oldUsername.Equal(username) {
				d.AddError("Password required", "Supply password when creating authenticated proxies or changing username.")
			} else {
				req.Authentication.RetainPassword = hcti.Ptr(true)
			}
		}
	}
	return req, d
}
func (m *proxyModel) read(ctx context.Context, p *management.Proxy, importing bool) diag.Diagnostics {
	var d diag.Diagnostics
	m.ID = types.StringValue(p.ID)
	if m.Name.IsNull() || strings.TrimSpace(m.Name.ValueString()) != p.Name {
		m.Name = types.StringValue(p.Name)
	}
	if m.URL.IsNull() || strings.TrimSpace(m.URL.ValueString()) != p.URL {
		m.URL = types.StringValue(p.URL)
	}
	m.Disabled = types.BoolValue(!p.Enabled)
	m.EffectivePort = types.Int64Null()
	if p.Port != nil {
		m.EffectivePort = types.Int64Value(int64(*p.Port))
	}
	if importing || !m.Port.IsNull() {
		m.Port = m.EffectivePort
	} else if p.Port != nil {
		defaultPort := uint16(80)
		if strings.HasPrefix(strings.TrimSpace(p.URL), "https:") {
			defaultPort = 443
		}
		if *p.Port != defaultPort {
			m.Port = m.EffectivePort
		}
	}
	var oldHosts []string
	if !m.BypassHosts.IsUnknown() && !m.BypassHosts.IsNull() {
		d.Append(m.BypassHosts.ElementsAs(ctx, &oldHosts, false)...)
	}
	if importing || fmt.Sprint(normalizeHosts(oldHosts)) != fmt.Sprint(normalizeHosts(p.BypassHosts)) || m.BypassHosts.IsUnknown() {
		if len(p.BypassHosts) == 0 {
			m.BypassHosts = types.SetNull(types.StringType)
		} else {
			var ds diag.Diagnostics
			m.BypassHosts, ds = types.SetValueFrom(ctx, types.StringType, p.BypassHosts)
			d.Append(ds...)
		}
	}
	oldUsername, password := authentication(m.Authentication)
	if p.Username == nil {
		m.Authentication = types.ObjectNull(authenticationTypes)
	} else {
		if oldUsername.IsNull() || oldUsername.ValueString() != *p.Username || password.IsUnknown() {
			password = types.StringNull()
		}
		m.Authentication = types.ObjectValueMust(authenticationTypes, map[string]attr.Value{"username": types.StringValue(*p.Username), "password": password})
	}
	m.CreatedAt = types.StringValue(p.CreatedAt.UTC().Format(time.RFC3339Nano))
	m.UpdatedAt = types.StringValue(p.UpdatedAt.UTC().Format(time.RFC3339Nano))
	return d
}
