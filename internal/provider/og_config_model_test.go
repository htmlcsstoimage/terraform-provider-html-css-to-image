package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/testapi"
)

func ogBaseModel() ogConfigModel {
	return ogConfigModel{ConfigType: types.StringValue("html_css"), Name: types.StringValue("Website cards"), BaseURL: types.StringValue("https://example.com"), Disabled: types.BoolValue(false), DefaultOptions: types.ObjectNull(ogOptionTypes()), TemplateValuesMapping: types.ListNull(ogMappingType), Headers: types.MapNull(types.StringType), AdditionalHeaderOrigins: types.SetNull(types.StringType)}
}
func ogObject(values map[string]attr.Value) types.Object {
	for k, t := range ogOptionTypes() {
		if _, ok := values[k]; !ok {
			values[k] = nullValue(t)
		}
	}
	return types.ObjectValueMust(ogOptionTypes(), values)
}
func TestOGOptionsNullAndExplicitRoundTrips(t *testing.T) {
	ctx := context.Background()
	cases := map[string]types.Object{
		"omitted": types.ObjectNull(ogOptionTypes()),
		"empty":   ogObject(map[string]attr.Value{}),
		"all fields": ogObject(map[string]attr.Value{
			"css": types.StringValue("body::after { content: '🌎'; }"), "device_scale": types.Float64Value(1.25), "max_wait_ms": types.Int64Value(500), "ms_delay": types.Int64Value(0), "render_when_ready": types.BoolValue(false), "selector": types.StringValue("#main"), "viewport_width": types.Int64Value(1200), "viewport_height": types.Int64Value(630), "disable_twemoji": types.BoolValue(true), "color_scheme": types.StringValue("dark"), "timezone": types.StringValue("America/New_York"), "block_consent_banners": types.BoolValue(false), "identify_as_hcti": types.BoolValue(false), "headers": types.MapValueMust(types.StringType, map[string]attr.Value{"X-Test": types.StringValue("🌎\tvalue")}), "additional_header_origins": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("https://assets.example.com")}), "include_headers_on_subrequests": types.BoolValue(true), "viewport_mobile": types.BoolValue(true), "viewport_landscape": types.BoolValue(false), "viewport_touch": types.BoolValue(false), "media_type": types.StringValue("screen"), "proxy_id": types.StringValue("proxy-test"), "storage_destination_id": types.StringValue("storage-test"), "transparent_background": types.BoolValue(true),
		}),
	}
	for name, options := range cases {
		t.Run(name, func(t *testing.T) {
			api := testapi.New(t)
			client := management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))
			m := ogBaseModel()
			m.DefaultOptions = options
			body, d := m.request(ctx)
			if d.HasError() {
				t.Fatal(d)
			}
			remote, err := client.CreateOGConfig(ctx, body)
			if err != nil {
				t.Fatal(err)
			}
			_, _, wire := api.OGSnapshot()
			if string(wire["description"]) != "null" {
				t.Fatal("description not explicit null")
			}
			if _, ok := wire["optimization_mode"]; ok {
				t.Fatal("injected optimization default")
			}
			if _, ok := wire["refresh_interval_s"]; ok {
				t.Fatal("injected refresh default")
			}
			if name == "omitted" {
				if string(wire["default_options"]) != "null" {
					t.Fatal("options not explicit null")
				}
			}
			if name == "empty" {
				var fields map[string]json.RawMessage
				json.Unmarshal(wire["default_options"], &fields)
				if len(fields) != len(ogOptionTypes()) {
					t.Fatal("options coverage mismatch")
				}
				for k := range ogOptionTypes() {
					if string(fields[k]) != "null" {
						t.Fatalf("%s was not null", k)
					}
				}
			}
			if d := m.read(ctx, remote, false); d.HasError() {
				t.Fatal(d)
			}
			if !m.DefaultOptions.Equal(options) {
				t.Fatal("option round trip changed nullable inputs")
			}
			if !m.OptimizationMode.IsNull() || !m.RefreshInterval.IsNull() || !m.ExtractValues.IsNull() {
				t.Fatal("defaults leaked into inputs")
			}
		})
	}
}
func TestOGHeaderValidation(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		bad              bool
	}{
		{"emoji boundary", "X-Test", strings.Repeat("😀", 2048), false},
		{"emoji overflow", "X-Test", strings.Repeat("😀", 2048) + "a", false},
		{"invalid UTF8", "X-Test", string([]byte{0xff}), true},
		{"newline", "X-Test", "secret\nvalue", true},
		{"NUL", "X-Test", "secret\x00value", true},
		{"tab", "X-Test", "secret\tvalue", false},
		{"nonASCII name", "X-😀", "value", true},
		{"invalid token", "X:Test", "value", true},
		{"name boundary", strings.Repeat("x", 512), "value", false},
		{"name overflow", strings.Repeat("x", 513), "value", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := ogBaseModel()
			m.DefaultOptions = ogObject(map[string]attr.Value{"headers": types.MapValueMust(types.StringType, map[string]attr.Value{tc.key: types.StringValue(tc.value)})})
			d := m.validate(context.Background())
			if d.HasError() != tc.bad {
				t.Fatalf("bad=%v diagnostics=%v", tc.bad, d)
			}
			for _, v := range d {
				if strings.Contains(v.Detail(), "secret") {
					t.Fatal("diagnostic disclosed secret")
				}
			}
		})
	}
	m := ogBaseModel()
	m.Name = types.StringValue(strings.Repeat("😀", 128))
	if m.validate(context.Background()).HasError() {
		t.Fatal("name length limit must be left to the API")
	}
}
func TestOGUnknowns(t *testing.T) {
	for _, name := range []string{"device_scale", "headers", "additional_header_origins"} {
		t.Run(name, func(t *testing.T) {
			m := ogBaseModel()
			var unknown attr.Value
			switch name {
			case "device_scale":
				unknown = types.Float64Unknown()
			case "headers":
				unknown = types.MapValueMust(types.StringType, map[string]attr.Value{"X-Test": types.StringUnknown()})
			default:
				unknown = types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()})
			}
			m.DefaultOptions = ogObject(map[string]attr.Value{name: unknown})
			if m.known() {
				t.Fatal("nested unknown was lost")
			}
			if _, d := m.request(context.Background()); !d.HasError() {
				t.Fatal("apply accepted unknown")
			}
			var schema resource.SchemaResponse
			(&OGConfigResource{}).Schema(context.Background(), resource.SchemaRequest{}, &schema)
			plan := tfsdk.Plan{Schema: schema.Schema}
			if d := plan.Set(context.Background(), &m); d.HasError() {
				t.Fatal(d)
			}
			resp := resource.ModifyPlanResponse{Plan: plan}
			(&OGConfigResource{}).ModifyPlan(context.Background(), resource.ModifyPlanRequest{Plan: plan}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatal("plan rejected unresolved dependency")
			}
		})
	}
}

func TestOGPolicyLimitsAreServerControlled(t *testing.T) {
	m := ogBaseModel()
	m.RefreshInterval = types.Int64Value(1)
	m.DefaultOptions = ogObject(map[string]attr.Value{
		"ms_delay":        types.Int64Value(10001),
		"max_wait_ms":     types.Int64Value(1),
		"device_scale":    types.Float64Value(4),
		"viewport_width":  types.Int64Value(6001),
		"viewport_height": types.Int64Value(6001),
	})
	if d := m.validate(context.Background()); d.HasError() {
		t.Fatal("server-controlled limits rejected locally", d)
	}
}
func TestOGImportPartialUpdateAndReadErrors(t *testing.T) {
	ctx := context.Background()
	api := testapi.New(t)
	c := management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))
	remote, err := c.CreateOGConfig(ctx, &management.TemplatedOGConfigRequest{OGConfigOptions: management.OGConfigOptions{Name: "Original", BaseURL: "https://example.com"}, TemplateID: "t-test", Headers: map[string]string{"Authorization": "secret"}})
	if err != nil {
		t.Fatal(err)
	}
	r := &OGConfigResource{client: c}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	if !schema.Schema.Attributes["headers"].IsSensitive() {
		t.Fatal("headers not sensitive")
	}
	m := ogBaseModel()
	m.ID = types.StringValue(remote.ID)
	m.ConfigType = types.StringNull()
	state := tfsdk.State{Schema: schema.Schema}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	read := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &read)
	if read.Diagnostics.HasError() {
		t.Fatal(read.Diagnostics)
	}
	if d := read.State.Get(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	if !m.TemplateVersion.IsNull() || m.ConfigType.ValueString() != "templated" || m.Headers.Elements()["Authorization"].(types.String).ValueString() != "secret" {
		t.Fatal("import failed")
	}
	// The API accepts disabling but ignores name, headers, and version changes.
	m.Name = types.StringValue("Edited")
	m.Disabled = types.BoolValue(true)
	m.TemplateVersion = types.Int64Value(123)
	m.Headers = types.MapNull(types.StringType)
	plan := tfsdk.Plan{Schema: schema.Schema}
	if d := plan.Set(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	api.DowngradeOG()
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: plan}, &update)
	if !update.Diagnostics.HasError() {
		t.Fatal("partial update reported success")
	}
	if d := update.State.Get(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	if !m.Disabled.ValueBool() || m.Name.ValueString() != "Original" || !m.TemplateVersion.IsNull() || len(m.Headers.Elements()) != 1 {
		t.Fatal("partial update did not save actual state")
	}
	for _, d := range update.Diagnostics {
		if strings.Contains(d.Detail(), "secret") {
			t.Fatal("diagnostic leaked header")
		}
	}
	for _, status := range []int{403, 429, 500} {
		api.Status(status)
		resp := resource.ReadResponse{State: update.State}
		r.Read(ctx, resource.ReadRequest{State: update.State}, &resp)
		if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(update.State.Raw) {
			t.Fatalf("HTTP %d did not preserve state", status)
		}
	}
	api.Status(0)
	api.OGMalformed(`{}`)
	bad := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &bad)
	if !bad.Diagnostics.HasError() || !bad.State.Raw.Equal(update.State.Raw) {
		t.Fatal("malformed read changed state")
	}
	api.OGMalformed("")
	api.Status(404)
	gone := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &gone)
	if gone.Diagnostics.HasError() || !gone.State.Raw.IsNull() {
		t.Fatal("404 did not remove state")
	}
	api.Status(0)
	if err := c.DeleteOGConfig(ctx, remote.ID); err != nil {
		t.Fatal(err)
	}
	deleted := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &deleted)
	if deleted.Diagnostics.HasError() {
		t.Fatal("repeated delete failed")
	}
}
