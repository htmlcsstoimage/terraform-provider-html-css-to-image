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

func TestProxyLifecycle(t *testing.T) {
	api := testapi.New(t)
	config := func(name, auth, extra string) string {
		return fmt.Sprintf(`
 provider "htmlcsstoimage" {
 api_id = "test-id"
 api_key = "test-key"
 base_url = %q
 }
 resource "htmlcsstoimage_proxy" "test" {
 name = %q
 url = "https://proxy.example"
 port = 8080
 %s
 %s
 }
 `, api.URL, name, auth, extra)
	}
	auth := `authentication = { username = "", password = "" }`
	retained := `authentication = { username = "" }`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())},
		Steps: []resource.TestStep{
			{Config: config("initial proxy", auth, `bypass_hosts = ["example.com"]`), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("htmlcsstoimage_proxy.test", "id", "proxy-test"), resource.TestCheckResourceAttr("htmlcsstoimage_proxy.test", "authentication.password", ""))},
			{Config: config("initial proxy", auth, `bypass_hosts = ["example.com"]`), PlanOnly: true},
			{Config: config("renamed proxy", retained, `bypass_hosts = ["example.com"]`)},
			{ResourceName: "htmlcsstoimage_proxy.test", ImportState: true, ImportStateVerify: true},
			{Config: config("changed user", `authentication = {username = "new-user"}`, ""), ExpectError: regexp.MustCompile("Password required")},
			{Config: config("changed user", `authentication = {username = "new-user", password = "new-secret"}`, "")},
			{Config: config("no auth", "", "")},
			{Config: config("no auth", "", ""), PlanOnly: true},
		},
	})
	if writes, retentions := api.Counts(); writes != 5 || retentions != 1 {
		t.Fatalf("writes=%d retentions=%d; expected create, three updates, delete", writes, retentions)
	}
}
func TestProxyNormalizationAndDefaultPort(t *testing.T) {
	api := testapi.New(t)
	config := fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="test-id"
 api_key="test-key"
 base_url=%q
}
 resource "htmlcsstoimage_proxy" "test" {
 name="  spaced name  "
 url="https://proxy.example"
 bypass_hosts=["EXAMPLE.COM", "https://example.com/path"]
 }`, api.URL)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttr("htmlcsstoimage_proxy.test", "effective_port", "443")},
		{Config: config, PlanOnly: true},
	}})
	if writes, _ := api.Counts(); writes != 2 {
		t.Fatalf("unchanged plan wrote to API: %d writes", writes)
	}
}
