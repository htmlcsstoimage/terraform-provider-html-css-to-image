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

func TestAPIKeyLifecycle(t *testing.T) {
	api := testapi.New(t)
	config := func(settings string) string {
		return fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="test-id"
 api_key="test-key"
 base_url=%q
 }
 resource "htmlcsstoimage_api_key" "test" {
 %s
 }`, api.URL, settings)
	}
	name := "htmlcsstoimage_api_key.test"
	first := config(`permissions=["images:create", "usage:read"]`)
	future := config(`all_future_permissions=true`)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{
		{Config: first, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "id", "key-management"), resource.TestCheckResourceAttr(name, "api_id", "key-auth-id"), resource.TestCheckResourceAttr(name, "api_key", "create-only-secret"), resource.TestCheckResourceAttr(name, "effective_name", "Key created 2026-09-16 12:00:00"))},
		{Config: first, PlanOnly: true},
		{Config: config(`name="  Renamed key  "
 description="  purpose  "
 permissions=[]
 disabled=true`), Check: resource.TestCheckResourceAttr(name, "api_key", "create-only-secret")},
		{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"api_key", "name", "description"}},
		{Config: future, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "api_key", "create-only-secret"), resource.TestCheckResourceAttr(name, "effective_permissions.#", "2"))},
		{PreConfig: api.ExpandKeyGrants, Config: future, PlanOnly: true},
		{Config: config(`all_future_permissions=true
 permissions=["images:create"]`), ExpectError: regexp.MustCompile("Conflicting permissions")},
		{Config: config(`name="missing grants"`), ExpectError: regexp.MustCompile("Permissions required")},
	}})
	k, writes := api.KeySnapshot()
	if k == nil || k.Enabled || writes != 4 {
		t.Fatalf("destroy must disable and retain key metadata; writes=%d key=%v", writes, k != nil)
	}
}
func TestAPIKeyBlankNormalization(t *testing.T) {
	api := testapi.New(t)
	config := fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="test-id"
 api_key="test-key"
 base_url=%q
 }
 resource "htmlcsstoimage_api_key" "test" {
 name=" "
 description=" "
 permissions=[]
 }`, api.URL)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{{Config: config}, {Config: config, PlanOnly: true}}})
}
