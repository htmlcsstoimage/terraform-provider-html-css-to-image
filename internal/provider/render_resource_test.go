package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type renderFixture struct {
	*httptest.Server
	mu                       sync.Mutex
	images                   map[string]map[string]any
	templates                []map[string]any
	creates, deletes, writes int
}

func newRenderFixture(t *testing.T) *renderFixture {
	t.Helper()
	s := &renderFixture{images: map[string]map[string]any{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		id, key, ok := r.BasicAuth()
		if !ok || id != "test-id" || key != "test-key" {
			t.Error("missing API auth")
			w.WriteHeader(401)
			return
		}
		if !strings.Contains(r.UserAgent(), "HCTITerraform/") {
			t.Error("missing provider user agent")
		}
		write := func(v any) { w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(v) }
		decode := func() map[string]any {
			var v map[string]any
			d := json.NewDecoder(r.Body)
			d.UseNumber()
			if e := d.Decode(&v); e != nil {
				t.Error(e)
			}
			return v
		}
		p := r.URL.Path
		switch {
		case strings.HasPrefix(p, "/v1/images/") && r.Method == "GET":
			v, ok := s.images[strings.TrimPrefix(p, "/v1/images/")]
			if !ok {
				w.WriteHeader(404)
				return
			}
			write(v)
		case (p == "/v1/image" || strings.HasPrefix(p, "/v1/image/t-")) && r.Method == "POST":
			v := decode()
			if v["dedupe_duration_s"] != json.Number("0") {
				t.Error("dedupe must be zero")
			}
			s.creates++
			id := fmt.Sprintf("image-%d", s.creates)
			kind := "html_css"
			if _, ok := v["url"]; ok {
				kind = "url"
			}
			if strings.HasPrefix(p, "/v1/image/t-") {
				kind = "templated"
				parts := strings.Split(strings.TrimPrefix(p, "/v1/image/"), "/")
				v["template_id"] = parts[0]
				version := "9007199254740993"
				if len(parts) > 1 {
					version = parts[1]
				}
				v["template_version"] = json.Number(version)
				if _, ok := v["device_scale"]; ok {
					t.Error("unexpected template body fields")
				}
			} else {
				if value, ok := v["device_scale"]; !ok || value != nil {
					t.Error("absent render input must be explicit null")
				}
			}
			format := v["format"]
			delete(v, "format")
			delete(v, "dedupe_duration_s")
			v["image_type"] = kind
			v["id"] = id
			v["created_at"] = "2026-09-16T00:00:00Z"
			v["storage_destination_hcti_storage_disabled"] = v["storage_destination_id"] == "storage-only"
			s.images[id] = v
			route := "image"
			if v["storage_destination_hcti_storage_disabled"] == true {
				route = "store"
			}
			suffix := ""
			if f, ok := format.(string); ok {
				suffix = "." + f
			}
			write(map[string]any{"id": id, "url": s.URL + "/v1/" + route + "/" + id + suffix})
		case strings.HasPrefix(p, "/v1/image/") && r.Method == "DELETE":
			delete(s.images, strings.TrimPrefix(p, "/v1/image/"))
			s.deletes++
			w.WriteHeader(202)
		case (p == "/v1/template" || p == "/v1/template/t-test") && r.Method == "POST":
			v := decode()
			s.writes++
			version := int64(9007199254740990) + int64(s.writes)
			v["id"] = "t-test"
			v["version"] = version
			v["template_type"] = "html_css"
			v["created_at"] = "2026-09-16T00:00:00Z"
			v["updated_at"] = "2026-09-16T00:00:00Z"
			s.templates = append([]map[string]any{v}, s.templates...)
			write(map[string]any{"template_id": "t-test", "template_version": version})
		case p == "/v1/template/t-test" && r.Method == "GET":
			write(map[string]any{"data": s.templates, "pagination": map[string]any{"next_page_start": nil}})
		case p == "/v1/template/t-test" && r.Method == "DELETE":
			s.templates = nil
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected request (rendering must never run): %s %s", r.Method, p)
			w.WriteHeader(500)
		}
	}))
	t.Cleanup(s.Close)
	return s
}
func renderConfig(s *renderFixture, kind, body string) string {
	return fmt.Sprintf(`provider "htmlcsstoimage" {
api_id="test-id"
api_key="test-key"
base_url=%q
}
resource "htmlcsstoimage_%s" "test" {
%s
}`, s.URL, kind, body)
}
func renderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}
}
func TestImageLifecycles(t *testing.T) {
	for _, kind := range []string{"html_css", "url", "templated"} {
		t.Run(kind, func(t *testing.T) {
			s := newRenderFixture(t)
			body := `html="<h1>Hello 😀</h1>"`
			if kind == "url" {
				body = "url=\"https://example.com\"\nheaders={Authorization=\"secret 😀\"}\n"
			}
			if kind == "templated" {
				body = "template_id=\"t-test\"\ntemplate_values=jsonencode({name=\"😀\",large=9007199254740993})\n"
				s.templates = []map[string]any{{"id": "t-test", "version": int64(9007199254740993), "template_type": "html_css", "html": "{{name}}", "created_at": "2026-09-16T00:00:00Z"}}
			}
			name := "htmlcsstoimage_image_" + kind + ".test"
			config := renderConfig(s, "image_"+kind, body)
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{
				{Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "render_method", "GET"), resource.TestCheckResourceAttr(name, "id", "image-1"))},
				{Config: config, PlanOnly: true},
				{ResourceName: name, ImportState: true, ImportStateVerify: true},
				{Config: renderConfig(s, "image_"+kind, body+"\nformat=\"webp\""), Check: resource.TestCheckResourceAttr(name, "id", "image-2")},
			}})
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.creates != 2 || s.deletes != 2 {
				t.Fatalf("creates=%d deletes=%d", s.creates, s.deletes)
			}
		})
	}
}
func TestTemplateLifecycle(t *testing.T) {
	s := newRenderFixture(t)
	name := "htmlcsstoimage_template.test"
	base := "html=\"<h1>{{name}}</h1>\"\ngoogle_fonts=[\"Roboto\",\"Open Sans\"]\n"
	config := renderConfig(s, "template", base)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttr(name, "version", "9007199254740991")},
		{Config: config, PlanOnly: true},
		{ResourceName: name, ImportState: true, ImportStateVerify: true},
		{Config: renderConfig(s, "template", base+"ms_delay=0\n"), Check: resource.TestCheckResourceAttr(name, "version", "9007199254740992")},
		{Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "version", "9007199254740993"), func(*terraform.State) error {
			s.mu.Lock()
			defer s.mu.Unlock()
			if v, ok := s.templates[0]["ms_delay"]; !ok || v != nil {
				return fmt.Errorf("removal did not send explicit null")
			}
			return nil
		})},
	}})
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writes != 3 {
		t.Fatal("unexpected writes: " + strconv.Itoa(s.writes))
	}
}
func TestStorageOnlyImage(t *testing.T) {
	s := newRenderFixture(t)
	name := "htmlcsstoimage_image_html_css.test"
	config := renderConfig(s, "image_html_css", "html=\"hello\"\nstorage_destination_id=\"storage-only\"")
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{{Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "render_method", "PUT"), resource.TestCheckResourceAttr(name, "render_requires_auth", "true"), resource.TestCheckNoResourceAttr(name, "public_url"))}, {ResourceName: name, ImportState: true, ImportStateVerify: true}}})
}

func TestTemplateOutOfBandVersion(t *testing.T) {
	s := newRenderFixture(t)
	name := "htmlcsstoimage_template.test"
	config := renderConfig(s, "template", "html=\"<h1>{{name}}</h1>\"")
	bump := func(change bool) func() {
		return func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			v := map[string]any{}
			for k, x := range s.templates[0] {
				v[k] = x
			}
			v["version"] = int64(9007199254741993)
			if change {
				v["html"] = "<h2>{{name}}</h2>"
			}
			s.templates = append([]map[string]any{v}, s.templates...)
		}
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{
		{Config: config},
		{PreConfig: bump(false), Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "version", "9007199254741993"), func(*terraform.State) error {
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.writes != 1 {
				return fmt.Errorf("identical remote content caused a write")
			}
			return nil
		})},
		{PreConfig: bump(true), Config: config, Check: resource.TestCheckResourceAttr(name, "html", "<h1>{{name}}</h1>")},
	}})
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writes != 2 {
		t.Fatalf("writes=%d", s.writes)
	}
}

func TestTemplatedImageOptionalPinAndSemanticJSON(t *testing.T) {
	s := newRenderFixture(t)
	name := "htmlcsstoimage_image_templated.test"
	body := "template_id=\"t-test\"\ntemplate_values=\"{\\\"n\\\":9007199254740993,\\\"x\\\":1}\"\n"
	equivalent := "template_id=\"t-test\"\ntemplate_values=\"{ \\\"x\\\": 1.0, \\\"n\\\": 9007199254740993 }\"\n"
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{
		{Config: renderConfig(s, "image_templated", body), Check: resource.ComposeTestCheckFunc(resource.TestCheckNoResourceAttr(name, "template_version"), resource.TestCheckResourceAttr(name, "resolved_template_version", "9007199254740993"))},
		{Config: renderConfig(s, "image_templated", equivalent), PlanOnly: true},
		{Config: renderConfig(s, "image_templated", body+"template_version=9007199254740995\n"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "id", "image-2"), resource.TestCheckResourceAttr(name, "resolved_template_version", "9007199254740995"))},
		{Config: renderConfig(s, "image_templated", body), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "id", "image-3"), resource.TestCheckNoResourceAttr(name, "template_version"))},
	}})
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.creates != 3 {
		t.Fatalf("creates=%d", s.creates)
	}
}

func TestImagePDFLifecycle(t *testing.T) {
	s := newRenderFixture(t)
	name := "htmlcsstoimage_image_html_css.test"
	body := `html="<h1>Hello</h1>"
format="pdf"
pdf_options={
page_width="8.5in"
page_height="11in"
scale=0.1
margins={top="1pt",right="2mm",bottom="3cm",left="4px"}
}
metadata={purpose="invoice"}
`
	config := renderConfig(s, "image_html_css", body)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{
		{Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "pdf_options.margins.top", "1pt"), resource.TestCheckResourceAttr(name, "metadata.purpose", "invoice"))},
		{Config: config, PlanOnly: true},
	}})
}

func TestTemplateReferencesReplaceOnlyPinnedImages(t *testing.T) {
	s := newRenderFixture(t)
	config := func(html string) string {
		return fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="test-id"
 api_key="test-key"
 base_url=%q
}
resource "htmlcsstoimage_template" "card" { html=%q }
resource "htmlcsstoimage_image_templated" "pinned" {
 template_id=htmlcsstoimage_template.card.id
 template_version=htmlcsstoimage_template.card.version
 template_values=jsonencode({name="hello"})
}
resource "htmlcsstoimage_image_templated" "latest" {
 template_id=htmlcsstoimage_template.card.id
 template_values=jsonencode({name="hello"})
}`, s.URL, html)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: renderFactories(), Steps: []resource.TestStep{
		{Config: config("<h1>{{name}}</h1>")},
		{Config: config("<h2>{{name}}</h2>"), Check: func(*terraform.State) error {
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.creates != 3 {
				return fmt.Errorf("expected only the pinned image replaced, creates=%d", s.creates)
			}
			return nil
		}},
	}})
}
