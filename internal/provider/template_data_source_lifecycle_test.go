package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTemplateDataSourceRefresh(t *testing.T) {
	var latest atomic.Int64
	latest.Store(100)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("data source performed write")
			w.WriteHeader(500)
			return
		}
		fmt.Fprintf(w, `{"data":[%s,%s],"pagination":{}}`, templateResponse(latest.Load()), templateResponse(42))
	}))
	defer s.Close()
	config := fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="id"
 api_key="key"
 base_url=%q
}
data "htmlcsstoimage_template" "latest" { id="t-test" }
data "htmlcsstoimage_template" "pinned" {
 id="t-test"
 version=42
}
data "htmlcsstoimage_template_versions" "all" { id="t-test" }
`, s.URL)
	check := func(v string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.htmlcsstoimage_template.latest", "version", v),
			resource.TestCheckResourceAttr("data.htmlcsstoimage_template.pinned", "version", "42"),
			resource.TestCheckResourceAttr("data.htmlcsstoimage_template_versions.all", "versions.#", "2"),
			resource.TestCheckResourceAttr("data.htmlcsstoimage_template_versions.all", "versions.0.version", v),
		)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{
		{Config: config, Check: check("100")},
		{PreConfig: func() { latest.Store(200) }, Config: config, Check: check("200")},
	}})
}
