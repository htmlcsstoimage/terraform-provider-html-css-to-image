package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type ogConfigModel struct {
	ID                        types.String `tfsdk:"id"`
	DomainID                  types.String `tfsdk:"domain_id"`
	ConfigType                types.String `tfsdk:"config_type"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	BaseURL                   types.String `tfsdk:"base_url"`
	Disabled                  types.Bool   `tfsdk:"disabled"`
	OptimizationMode          types.String `tfsdk:"optimization_mode"`
	EffectiveOptimizationMode types.String `tfsdk:"effective_optimization_mode"`
	RefreshInterval           types.Int64  `tfsdk:"refresh_interval_s"`
	EffectiveRefreshInterval  types.Int64  `tfsdk:"effective_refresh_interval_s"`
	ExtractValues             types.Bool   `tfsdk:"extract_values"`
	DefaultOptions            types.Object `tfsdk:"default_options"`
	TemplateID                types.String `tfsdk:"template_id"`
	TemplateVersion           types.Int64  `tfsdk:"template_version"`
	TemplateValuesMapping     types.List   `tfsdk:"template_values_mapping"`
	Headers                   types.Map    `tfsdk:"headers"`
	AdditionalHeaderOrigins   types.Set    `tfsdk:"additional_header_origins"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}

func (m ogConfigModel) inputs() map[string]attr.Value {
	return map[string]attr.Value{"config_type": m.ConfigType, "name": m.Name, "description": m.Description, "base_url": m.BaseURL, "disabled": m.Disabled, "optimization_mode": m.OptimizationMode, "refresh_interval_s": m.RefreshInterval, "extract_values": m.ExtractValues, "default_options": m.DefaultOptions, "template_id": m.TemplateID, "template_version": m.TemplateVersion, "template_values_mapping": m.TemplateValuesMapping, "headers": m.Headers, "additional_header_origins": m.AdditionalHeaderOrigins}
}

func (m ogConfigModel) known() bool {
	for _, v := range m.inputs() {
		if !knownValue(v) {
			return false
		}
	}
	return true
}

func (m ogConfigModel) request(ctx context.Context) (management.OGConfigRequest, diag.Diagnostics) {
	d := m.validate(ctx)
	if d.HasError() {
		return nil, d
	}
	common := management.OGConfigOptions{Name: m.Name.ValueString(), Description: m.Description.ValueStringPointer(), BaseURL: m.BaseURL.ValueString(), Disabled: m.Disabled.ValueBool()}
	if !m.OptimizationMode.IsNull() {
		common.OptimizationMode = hcti.Ptr(management.OptimizationMode(m.OptimizationMode.ValueString()))
	}
	if !m.RefreshInterval.IsNull() {
		common.RefreshIntervalSeconds = hcti.Ptr(uint32(m.RefreshInterval.ValueInt64()))
	}
	if m.ConfigType.ValueString() == "html_css" {
		options, ds := ogOptionsRequest(ctx, m.DefaultOptions)
		d.Append(ds...)
		return &management.HTMLCSSOGConfigRequest{OGConfigOptions: common, DefaultOptions: options, ExtractValues: m.ExtractValues.ValueBool()}, d
	}
	r := &management.TemplatedOGConfigRequest{OGConfigOptions: common, TemplateID: m.TemplateID.ValueString(), TemplateVersion: m.TemplateVersion.ValueInt64Pointer()}
	if !m.Headers.IsNull() {
		d.Append(m.Headers.ElementsAs(ctx, &r.Headers, false)...)
	}
	if !m.AdditionalHeaderOrigins.IsNull() {
		d.Append(m.AdditionalHeaderOrigins.ElementsAs(ctx, &r.AdditionalHeaderOrigins, false)...)
	}
	if !m.TemplateValuesMapping.IsNull() {
		r.TemplateValuesMapping = make([]management.OGTemplateValueMapping, 0, len(m.TemplateValuesMapping.Elements()))
		for _, v := range m.TemplateValuesMapping.Elements() {
			a := v.(types.Object).Attributes()
			item := management.OGTemplateValueMapping{TemplateKey: a["template_key"].(types.String).ValueString(), MetaKey: a["meta_key"].(types.String).ValueStringPointer()}
			if !a["fallback"].IsNull() {
				item.Fallback = hcti.Ptr(management.OGMetadataFallback(a["fallback"].(types.String).ValueString()))
			}
			r.TemplateValuesMapping = append(r.TemplateValuesMapping, item)
		}
	}
	return r, d
}

func (m *ogConfigModel) read(ctx context.Context, r *management.OGConfig, importing bool) diag.Diagnostics {
	var d diag.Diagnostics
	// Retain the configured spelling only when it represents the same server value.
	keep := func(name string, old, remote attr.Value) attr.Value {
		if !importing && ogEquivalent(name, old, remote) {
			return old
		}
		return remote
	}
	m.Name = keep("name", m.Name, types.StringPointerValue(r.Name)).(types.String)
	m.Description = keep("description", m.Description, types.StringPointerValue(r.Description)).(types.String)
	m.BaseURL = keep("base_url", m.BaseURL, types.StringValue(r.BaseURL)).(types.String)
	m.ConfigType = types.StringValue(string(r.ConfigType))
	m.Disabled = types.BoolValue(!r.Enabled)
	m.EffectiveOptimizationMode = types.StringValue(string(r.OptimizationMode))
	m.EffectiveRefreshInterval = types.Int64Value(int64(r.RefreshIntervalSeconds))
	m.OptimizationMode = keep("optimization_mode", m.OptimizationMode, m.EffectiveOptimizationMode).(types.String)
	m.RefreshInterval = keep("refresh_interval_s", m.RefreshInterval, m.EffectiveRefreshInterval).(types.Int64)
	if r.ConfigType == management.OGConfigHTMLCSS {
		options, ds := ogOptionsRead(ctx, r.DefaultOptions)
		d.Append(ds...)
		m.DefaultOptions = keep("default_options", m.DefaultOptions, options).(types.Object)
		m.ExtractValues = keep("extract_values", m.ExtractValues, types.BoolValue(r.ExtractValues)).(types.Bool)
		m.TemplateID = types.StringNull()
		m.TemplateVersion = types.Int64Null()
		m.TemplateValuesMapping = types.ListNull(ogMappingType)
		m.Headers = types.MapNull(types.StringType)
		m.AdditionalHeaderOrigins = types.SetNull(types.StringType)
	} else {
		m.DefaultOptions = types.ObjectNull(ogOptionTypes())
		m.ExtractValues = types.BoolNull()
		m.TemplateID = keep("template_id", m.TemplateID, types.StringPointerValue(r.TemplateID)).(types.String)
		m.TemplateVersion = types.Int64PointerValue(r.TemplateVersion)
		values := []attr.Value{}
		for _, item := range r.TemplateValuesMapping {
			fallback := types.StringNull()
			if item.Fallback != nil {
				fallback = types.StringValue(string(*item.Fallback))
			}
			values = append(values, types.ObjectValueMust(ogMappingTypes, map[string]attr.Value{"template_key": types.StringValue(item.TemplateKey), "meta_key": types.StringPointerValue(item.MetaKey), "fallback": fallback}))
		}
		mappings := types.ListNull(ogMappingType)
		if len(values) > 0 {
			mappings = types.ListValueMust(ogMappingType, values)
		}
		m.TemplateValuesMapping = keep("template_values_mapping", m.TemplateValuesMapping, mappings).(types.List)
		headers, ds := types.MapValueFrom(ctx, types.StringType, r.Headers)
		d.Append(ds...)
		m.Headers = keep("headers", m.Headers, headers).(types.Map)
		origins, ds := types.SetValueFrom(ctx, types.StringType, r.AdditionalHeaderOrigins)
		d.Append(ds...)
		m.AdditionalHeaderOrigins = keep("additional_header_origins", m.AdditionalHeaderOrigins, origins).(types.Set)
	}
	m.ID = types.StringValue(r.ID)
	m.DomainID = types.StringValue(r.DomainID)
	m.CreatedAt = types.StringValue(r.CreatedAt.UTC().Format(time.RFC3339Nano))
	m.UpdatedAt = types.StringValue(r.UpdatedAt.UTC().Format(time.RFC3339Nano))
	return d
}

// Verify writes against readback: an HTTP success may only have disabled the
// configuration after a plan downgrade. Never include values in diagnostics.
func ogUnapplied(want, got ogConfigModel) []string {
	var fields []string
	actual := got.inputs()
	for name, v := range want.inputs() {
		if !ogEquivalent(name, v, actual[name]) {
			fields = append(fields, name)
		}
	}
	return fields
}

func ogTrim(s types.String) string { return strings.TrimSpace(s.ValueString()) }
