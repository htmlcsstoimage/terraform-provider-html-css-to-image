package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

func readTemplateData(t *testing.T, url string, versions bool, pin *int64, limit ...int64) datasource.ReadResponse {
	t.Helper()
	ctx := context.Background()
	d := &templateDataSource{versions: versions, clients: &dataSourceClients{management: management.NewClient("id", "key", management.WithBaseURL(url)), public: hcti.NewClient("id", "key", hcti.WithBaseURL(url))}}
	var s datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &s)
	attrs := map[string]attr.Type{}
	values := map[string]attr.Value{}
	for k, v := range s.Schema.Attributes {
		attrs[k] = v.GetType()
		values[k] = nullValue(v.GetType())
	}
	values["id"] = types.StringValue("t-test")
	if len(limit) > 0 {
		values["limit"] = types.Int64Value(limit[0])
	}
	if pin != nil {
		values["version"] = types.Int64Value(*pin)
	}
	obj := types.ObjectValueMust(attrs, values)
	raw, err := obj.ToTerraformValue(ctx)
	if err != nil {
		t.Fatal(err)
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: raw}}, &resp)
	return resp
}
func templateResponse(version int64) string {
	return fmt.Sprintf(`{"id":"t-test","version":%d,"template_type":"html_css","html":"<h1>{{title}}</h1>","css":null,"google_fonts":"Inter|Roboto","ms_delay":0,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}`, version)
}
func TestTemplateDataLatestAndPinned(t *testing.T) {
	latest := int64(9007199254740993)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/template/t-test" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
			w.WriteHeader(500)
			return
		}
		if r.URL.Query().Get("max_version") != "" {
			fmt.Fprintf(w, `{"data":[%s],"pagination":{}}`, templateResponse(42))
			return
		}
		fmt.Fprintf(w, `{"data":[%s],"pagination":{"next_page_start":100}}`, templateResponse(latest))
	}))
	defer s.Close()
	for _, pin := range []*int64{nil, hcti.Ptr(int64(42)), nil} {
		resp := readTemplateData(t, s.URL, false, pin)
		if resp.Diagnostics.HasError() {
			t.Fatal(resp.Diagnostics)
		}
		var state types.Object
		resp.State.Get(context.Background(), &state)
		expected := latest
		if pin != nil {
			expected = *pin
		}
		if state.Attributes()["version"].(types.Int64).ValueInt64() != expected {
			t.Fatal("version lost", state)
		}
		if !state.Attributes()["css"].IsNull() || state.Attributes()["ms_delay"].(types.Int64).ValueInt64() != 0 {
			t.Fatal("null/zero changed")
		}
		latest++
	}
}
func TestTemplateVersionsPagination(t *testing.T) {
	count := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if r.Method != "GET" || r.URL.Query().Get("count") != "100" {
			t.Error("unexpected request", r.URL)
		}
		if count == 1 {
			fmt.Fprintf(w, `{"data":[%s],"pagination":{"next_page_start":9007199254740992}}`, templateResponse(9007199254740993))
			return
		}
		if r.URL.Query().Get("max_version") != "9007199254740992" {
			t.Error("cursor precision lost", r.URL)
		}
		fmt.Fprintf(w, `{"data":[%s],"pagination":{}}`, templateResponse(42))
	}))
	defer s.Close()
	resp := readTemplateData(t, s.URL, true, nil)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var state types.Object
	resp.State.Get(context.Background(), &state)
	list := state.Attributes()["versions"].(types.List).Elements()
	if count != 2 || len(list) != 2 || list[0].(types.Object).Attributes()["version"].(types.Int64).ValueInt64() != 9007199254740993 {
		t.Fatal("pagination or precision lost", state)
	}
}
func TestTemplateLookupFailures(t *testing.T) {
	for _, versions := range []bool{false, true} {
		for _, code := range []int{400, 403, 404, 429, 500} {
			t.Run(fmt.Sprintf("%t/%d", versions, code), func(t *testing.T) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(code)
					fmt.Fprint(w, `{"error":"Denied","message":"Lookup failed"}`)
				}))
				defer s.Close()
				if resp := readTemplateData(t, s.URL, versions, nil); !resp.Diagnostics.HasError() {
					t.Fatal("error swallowed")
				}
			})
		}
	}
}
func TestTemplateVersionsStuckCursor(t *testing.T) {
	count := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		fmt.Fprint(w, `{"data":[],"pagination":{"next_page_start":100}}`)
	}))
	defer s.Close()
	if resp := readTemplateData(t, s.URL, true, nil); !resp.Diagnostics.HasError() || count != 2 {
		t.Fatal("nonadvancing pagination accepted", count, resp.Diagnostics)
	}
}

func TestTemplateVersionsLimit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		limit  []int64
		want   int
		counts []int
	}{
		{"default", nil, 1000, []int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100}},
		{"small", []int64{10}, 10, []int{10}},
		{"partial final page", []int64{125}, 125, []int{100, 25}},
		{"above default", []int64{1001}, 1001, []int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			next := int64(10000)
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n, err := strconv.Atoi(r.URL.Query().Get("count"))
				if err != nil || calls >= len(tc.counts) {
					t.Error("unexpected request", r.URL)
					w.WriteHeader(500)
					return
				}
				if n != tc.counts[calls] {
					t.Errorf("count=%d want=%d", n, tc.counts[calls])
				}
				calls++
				parts := []string{}
				for i := 0; i < n; i++ {
					parts = append(parts, templateResponse(next))
					next--
				}
				fmt.Fprintf(w, `{"data":[%s],"pagination":{"next_page_start":%d}}`, strings.Join(parts, ","), next)
			}))
			defer s.Close()
			resp := readTemplateData(t, s.URL, true, nil, tc.limit...)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
			var state types.Object
			resp.State.Get(context.Background(), &state)
			if len(state.Attributes()["versions"].(types.List).Elements()) != tc.want || calls != len(tc.counts) || state.Attributes()["limit"].(types.Int64).ValueInt64() != int64(tc.want) {
				t.Fatal("limit not respected", calls)
			}
		})
	}
}
func TestTemplateVersionsInvalidLimit(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("invalid limit made API request") }))
	defer s.Close()
	for _, limit := range []int64{0, -1} {
		if resp := readTemplateData(t, s.URL, true, nil, limit); !resp.Diagnostics.HasError() {
			t.Fatal("accepted invalid limit")
		}
	}
}
