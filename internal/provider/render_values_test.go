package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

func TestTemplatedImageImportUsesSavedStorageMode(t *testing.T) {
	for _, storageOnly := range []bool{false, true} {
		t.Run(fmt.Sprint(storageOnly), func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected lookup during metadata merge: %s", r.URL.Path)
				w.WriteHeader(http.StatusForbidden)
			}))
			defer s.Close()
			r := &RenderResource{kind: "templated", client: management.NewClient("id", "key", management.WithBaseURL(s.URL))}
			id, version := "t-deleted", int64(9007199254740993)
			remote := &management.SavedRender{ID: "image-id", ImageType: "templated", StorageDestinationHCTIStorageDisabled: &storageOnly, RenderDefinition: management.RenderDefinition{TemplateID: &id, TemplateVersion: &version, TemplateValues: json.RawMessage(`{"title":"test"}`)}}
			a := map[string]attr.Value{}
			if d := r.merge(context.Background(), a, remote, true); d.HasError() {
				t.Fatal(d)
			}
			route, method := "image", "GET"
			if storageOnly {
				route, method = "store", "PUT"
			}
			if !a["image_url"].Equal(types.StringValue(s.URL+"/v1/"+route+"/image-id")) || !a["render_method"].Equal(types.StringValue(method)) || !a["render_requires_auth"].Equal(types.BoolValue(storageOnly)) || !a["storage_destination_hcti_storage_disabled"].Equal(types.BoolValue(storageOnly)) {
				t.Fatalf("incorrect saved storage mode: %v", a)
			}
			if a["public_url"].IsNull() != storageOnly {
				t.Fatal("incorrect public URL visibility")
			}
		})
	}
}

func TestTemplatedImageReadbackRecoveryPreservesPin(t *testing.T) {
	r := &RenderResource{kind: "templated", client: management.NewClient("id", "key")}
	id, version := "t-test", int64(9007199254740993)
	remote := &management.SavedRender{ID: "image-id", ImageType: "templated", RenderDefinition: management.RenderDefinition{TemplateID: &id, TemplateVersion: &version, TemplateValues: json.RawMessage(`{"title":"test"}`)}}
	for _, pin := range []types.Int64{types.Int64Value(version), types.Int64Null(), types.Int64Unknown()} {
		a := map[string]attr.Value{"template_version": pin}
		if d := r.merge(context.Background(), a, remote, true); d.HasError() {
			t.Fatal(d)
		}
		want := pin
		if pin.IsUnknown() {
			want = types.Int64Null()
		}
		if !a["template_version"].Equal(want) || !a["resolved_template_version"].Equal(types.Int64Value(version)) {
			t.Fatalf("pin lost or inferred during recovery: %v", a)
		}
	}
}

func TestRenderJSONEquality(t *testing.T) {
	for _, v := range []struct {
		a, b  string
		equal bool
	}{
		{`{"n":9007199254740993,"emoji":"😀"}`, `{"emoji":"\ud83d\ude00","n":9007199254740993.0}`, true},
		{`{"n":9007199254740993}`, `{"n":9007199254740992}`, false},
		{`{"n":1e3,"x":[null,false,{"y":"a"}]}`, `{"x":[null,false,{"y":"a"}],"n":1000}`, true},
		{`{"x":null}`, `{"y":null}`, false},
		{`{"x":1} {"x":2}`, `{"x":1}`, false},
	} {
		if got := renderEquivalent("template_values", types.StringValue(v.a), types.StringValue(v.b)); got != v.equal {
			t.Errorf("equality %s / %s = %v", v.a, v.b, got)
		}
	}
}
func TestRenderMappingRoundTrips(t *testing.T) {
	ctx := context.Background()
	for _, kind := range []string{"template", "html_css", "url", "templated"} {
		t.Run(kind, func(t *testing.T) {
			a := map[string]attr.Value{}
			for k, s := range renderAttributes(kind) {
				a[k] = nullValue(s.GetType())
			}
			// Every optional scalar survives null and explicit values without injecting defaults.
			for k, s := range renderAttributes(kind) {
				switch s.GetType() {
				case types.StringType:
					a[k] = types.StringValue("value 😀")
				case types.BoolType:
					a[k] = types.BoolValue(false)
				case types.Int64Type:
					a[k] = types.Int64Value(9007199254740993)
				case types.Float64Type:
					a[k] = types.Float64Value(1.5)
				}
			}
			if _, ok := a["template_values"]; ok {
				a["template_values"] = types.StringValue(`{"n":9007199254740993}`)
			}
			if _, ok := a["google_fonts"]; ok {
				a["google_fonts"] = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("Roboto"), types.StringValue("Open Sans")})
			}
			if _, ok := a["headers"]; ok {
				a["headers"] = types.MapValueMust(types.StringType, map[string]attr.Value{"Authorization": types.StringValue("secret 😀")})
			}
			request, d := renderRequest(ctx, a)
			if d.HasError() {
				t.Fatal(d)
			}
			got, d := renderReadValues(ctx, kind, *request)
			if d.HasError() {
				t.Fatal(d)
			}
			for k, v := range a {
				if !renderEquivalent(k, v, got[k]) {
					t.Errorf("%s failed round trip", k)
				}
			}
			for k, s := range renderAttributes(kind) {
				a[k] = nullValue(s.GetType())
			}
			request, d = renderRequest(ctx, a)
			if d.HasError() {
				t.Fatal(d)
			}
			got, d = renderReadValues(ctx, kind, *request)
			if d.HasError() {
				t.Fatal(d)
			}
			for k, v := range a {
				if !v.Equal(got[k]) {
					t.Errorf("%s null did not round trip", k)
				}
			}
		})
	}
}
func TestRenderReadErrorsPreserveState(t *testing.T) {
	ctx := context.Background()
	status := 200
	body := `{"id":"image-1","image_type":"html_css","html":"hello","created_at":"2026-09-16T00:00:00Z"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status); fmt.Fprint(w, body) }))
	defer server.Close()
	r := &RenderResource{kind: "html_css", client: management.NewClient("id", "key", management.WithBaseURL(server.URL))}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	a := map[string]attr.Value{"id": types.StringValue("image-1"), "html": types.StringValue("hello"), "created_at": types.StringValue("2026-09-16T00:00:00Z")}
	if d := setRenderState(ctx, &state, a); d.HasError() {
		t.Fatal(d)
	}
	for _, code := range []int{403, 429, 500, 200} {
		status = code
		if code == 200 {
			body = `{}`
		}
		resp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &resp)
		if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(state.Raw) {
			t.Fatalf("status %d did not preserve state", code)
		}
	}
	status = 404
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatal("404 did not remove state")
	}
}
func TestPDFNormalizationAndMargins(t *testing.T) {
	ctx := context.Background()
	var v management.SavedRender
	if err := json.Unmarshal([]byte(`{"pdf_options":{"page_width":"8.5in","print_background":false,"scale":0.10000000149011612,"margins":["1pt","2mm","3cm","4px"]}}`), &v); err != nil {
		t.Fatal(err)
	}
	p, d := renderPDFRead(ctx, v.PDFOptions)
	if d.HasError() {
		t.Fatal(d)
	}
	a := p.Attributes()
	a["page_width"] = types.StringValue("8.50in")
	a["print_background"] = types.BoolNull()
	a["scale"] = types.Float64Value(0.1)
	want := types.ObjectValueMust(p.AttributeTypes(ctx), a)
	if !renderEquivalent("pdf_options", want, p) {
		t.Fatal("PDF normalized values differed")
	}
	req, d := renderRequest(ctx, map[string]attr.Value{"pdf_options": want})
	if d.HasError() {
		t.Fatal(d)
	}
	if req.PDFOptions.Margins[0] != "1pt" || req.PDFOptions.Margins[3] != "4px" || req.PDFOptions.PrintBackground != nil {
		t.Fatal("PDF wire mapping failed")
	}
}

func TestRenderValidation(t *testing.T) {
	ctx := context.Background()
	base := func(kind string) map[string]attr.Value {
		a := map[string]attr.Value{}
		for k, s := range renderAttributes(kind) {
			a[k] = nullValue(s.GetType())
		}
		return a
	}
	for _, tc := range []struct {
		name, kind, key string
		v               attr.Value
	}{
		{"viewport pair", "html_css", "viewport_width", types.Int64Value(100)},
		{"invalid scale", "html_css", "device_scale", types.Float64Value(-1)},
		{"invalid delay", "html_css", "ms_delay", types.Int64Value(-1)},
		{"invalid source", "url", "url", types.StringValue("file:///tmp/file")},
		{"invalid pin", "templated", "template_version", types.Int64Value(0)},
		{"empty values", "templated", "template_values", types.StringValue("{}")},
		{"array values", "templated", "template_values", types.StringValue("[]")},
		{"no placeholder", "template", "html", types.StringValue("hello")},
		{"CSS handlebars", "template", "css", types.StringValue("{{color}}")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := base(tc.kind)
			a[tc.key] = tc.v
			if d := validateRender(ctx, tc.kind, a); !d.HasError() {
				t.Fatal("invalid input accepted")
			}
		})
	}
	a := base("html_css")
	a["viewport_width"] = types.Int64Value(100)
	a["viewport_height"] = types.Int64Unknown()
	if d := validateRender(ctx, "html_css", a); d.HasError() {
		t.Fatal("unknown paired dimension was rejected", d)
	}
	a = base("templated")
	a["template_id"] = types.StringUnknown()
	a["template_version"] = types.Int64Unknown()
	a["template_values"] = types.StringUnknown()
	if d := validateRender(ctx, "templated", a); d.HasError() {
		t.Fatal(d)
	}
}

func TestRenderPolicyLimitsAreServerControlled(t *testing.T) {
	ctx := context.Background()
	for _, kind := range []string{"html_css", "url", "templated", "template"} {
		t.Run(kind, func(t *testing.T) {
			a := map[string]attr.Value{}
			for k, s := range renderAttributes(kind) {
				a[k] = nullValue(s.GetType())
			}
			for k, v := range map[string]attr.Value{
				"ms_delay":         types.Int64Value(10001),
				"max_wait_ms":      types.Int64Value(1),
				"viewport_width":   types.Int64Value(6001),
				"viewport_height":  types.Int64Value(6001),
				"device_scale":     types.Float64Value(4),
				"jumbo_max_width":  types.Int64Value(80001),
				"jumbo_max_height": types.Int64Value(80001),
				"timezone":         types.StringValue("Future/Zone"),
				"format":           types.StringValue("future-format"),
				"name":             types.StringValue(strings.Repeat("x", 1025)),
			} {
				if _, ok := a[k]; ok {
					a[k] = v
				}
			}
			if d := validateRender(ctx, kind, a); d.HasError() {
				t.Fatal("server-controlled values rejected locally", d)
			}
		})
	}
}

func TestSavedIdentitySurvivesReadbackFailure(t *testing.T) {
	ctx := context.Background()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			fmt.Fprint(w, `{"id":"image-created","url":"https://hcti.io/v1/image/image-created"}`)
		} else {
			w.WriteHeader(500)
		}
	}))
	defer s.Close()
	r := &RenderResource{kind: "html_css", client: management.NewClient("id", "key", management.WithBaseURL(s.URL))}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	a := map[string]attr.Value{"html": types.StringValue("hello")}
	if d := setRenderState(ctx, &state, a); d.HasError() {
		t.Fatal(d)
	}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: schema.Schema, Raw: state.Raw}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("read failure was hidden")
	}
	got, d := renderState(ctx, resp.State)
	if d.HasError() || got["id"].(types.String).ValueString() != "image-created" {
		t.Fatal("created ID was lost", d)
	}
}
