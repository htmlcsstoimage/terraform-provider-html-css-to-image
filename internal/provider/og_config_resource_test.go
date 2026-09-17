package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/testapi"
)

func ogTestConfig(api *testapi.Server, settings string) string {
	return fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="test-id"
 api_key="test-key"
 base_url=%q
}
resource "htmlcsstoimage_og_config" "test" {
 name="  Website cards  "
 base_url="https://example.com/"
 %s
}`, api.URL, settings)
}
func TestOGConfigLifecycle(t *testing.T) {
	api := testapi.New(t)
	name := "htmlcsstoimage_og_config.test"
	html := ogTestConfig(api, `config_type="html_css"`)
	styled := ogTestConfig(api, `config_type="html_css"
 extract_values=true
 default_options={
 css="body { color: red }"
 viewport_width=1200
 viewport_height=630
 device_scale=1
 ms_delay=0
 headers={ Authorization="Bearer test-secret-😀" }
 additional_header_origins=["https://assets.example.com"]
 }`)
	templated := ogTestConfig(api, `config_type="templated"
 template_id="t-test"
 template_values_mapping=[
 {template_key="headline",meta_key="og:title"},
 {template_key="summary",fallback="descriptions"}
 ]
 headers={ Authorization="Bearer template-secret" }
 additional_header_origins=["https://metadata.example.com"]`)
	pinned := ogTestConfig(api, `config_type="templated"
 template_id="t-test"
 template_version=9007199254740993
 headers={}
 template_values_mapping=[]
 additional_header_origins=[]
 optimization_mode="set_viewport"
 refresh_interval_s=1800`)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{
		{Config: html, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "id", "og-management"), resource.TestCheckResourceAttr(name, "domain_id", "domain-html_css"), resource.TestCheckNoResourceAttr(name, "default_options"), resource.TestCheckNoResourceAttr(name, "optimization_mode"), resource.TestCheckResourceAttr(name, "effective_optimization_mode", "post_process"))},
		{Config: html, PlanOnly: true},
		{Config: styled, Check: resource.TestCheckResourceAttr(name, "default_options.headers.Authorization", "Bearer test-secret-😀")},
		{Config: styled, PlanOnly: true},
		{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"name", "base_url", "optimization_mode", "refresh_interval_s", "default_options.include_headers_on_subrequests"}},
		{Config: templated, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "id", "og-management"), resource.TestCheckResourceAttr(name, "domain_id", "domain-templated"), resource.TestCheckNoResourceAttr(name, "default_options"), resource.TestCheckNoResourceAttr(name, "template_version"), resource.TestCheckResourceAttr(name, "template_values_mapping.0.meta_key", "og:title"), resource.TestCheckResourceAttr(name, "headers.Authorization", "Bearer template-secret"))},
		{Config: templated, PlanOnly: true},
		{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"name", "base_url", "optimization_mode", "refresh_interval_s"}},
		{Config: pinned, Check: resource.TestCheckResourceAttr(name, "template_version", "9007199254740993")},
		{Config: pinned, PlanOnly: true},
		{Config: html, Check: resource.ComposeTestCheckFunc(resource.TestCheckNoResourceAttr(name, "template_id"), resource.TestCheckNoResourceAttr(name, "headers"))},
		{Config: html, PlanOnly: true},
	}})
	c, w, _ := api.OGSnapshot()
	if c != nil || w != 6 {
		t.Fatalf("unexpected lifecycle writes: %d, deleted=%v", w, c == nil)
	}
}
func TestOGValidation(t *testing.T) {
	api := testapi.New(t)
	for _, settings := range []string{
		`config_type="unknown"`,
		`config_type="html_css"
 template_id="t-test"`,
		`config_type="templated"`,
		`config_type="templated"
 template_id="t-test"
 extract_values=false`,
		`config_type="html_css"
 default_options={viewport_width=1200}`,
		`config_type="html_css"
 default_options={ms_delay=4294967296}`,
		`config_type="html_css"
 default_options={include_headers_on_subrequests=true}`,
		`config_type="templated"
 template_id="t-test"
 template_values_mapping=[{template_key="title",meta_key="og:title",fallback="titles"}]`,
	} {
		resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{{Config: ogTestConfig(api, settings), ExpectError: regexp.MustCompile("Invalid OG configuration")}}})
	}
	_, w, _ := api.OGSnapshot()
	if w != 0 {
		t.Fatal("validation wrote to API")
	}
}
