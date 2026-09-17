package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/testapi"
)

func storageConfig(api *testapi.Server, name, connection, extra string) string {
	match := regexp.MustCompile(`provider="([^"]+)"`).FindStringSubmatch(connection)
	kind := match[1]
	connection = regexp.MustCompile(`provider="[^"]+"`).ReplaceAllString(connection, "")
	return fmt.Sprintf(`provider "htmlcsstoimage" {
 api_id="test-id"
 api_key="test-key"
 base_url=%q
}
data "htmlcsstoimage_aws_storage_external_id" "org" {}
resource "htmlcsstoimage_storage_destination" "test" {
 name=%q
 %s
 connection_info={
 %s={
 bucket="  existing-bucket  "
 key_prefix=" /images/cards/ "
 %s
 }
 }
}`, api.URL, name, extra, kind, connection)
}
func TestStorageLifecycle(t *testing.T) {
	api := testapi.New(t)
	name := "htmlcsstoimage_storage_destination.test"
	factories := map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}
	aws := `provider="aws_s3"
 region="us-east-1"
 role_arn="arn:aws:iam::123456789012:role/hcti"`
	r2 := `provider="cloudflare_r2"
 cloudflare_account_id="0123456789abcdef0123456789abcdef"
 access_key_id="key-id"
 secret_access_key=" secret-😀 "`
	steps := []resource.TestStep{
		{Config: storageConfig(api, "  Original  ", aws, ""), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(name, "id", "storage-test"), resource.TestCheckResourceAttr("data.htmlcsstoimage_aws_storage_external_id.org", "external_id", "org-external-test"), resource.TestCheckResourceAttr("data.htmlcsstoimage_aws_storage_external_id.org", "writer_role_arn", "arn:aws:iam::123456789012:role/hcti-writer"))},
		{Config: storageConfig(api, "  Original  ", aws, ""), PlanOnly: true},
		{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"name", "connection_info.aws_s3.bucket", "connection_info.aws_s3.key_prefix"}},
		{Config: storageConfig(api, "R2 destination", r2, ""), Check: resource.TestCheckResourceAttr(name, "connection_info.cloudflare_r2.secret_access_key", " secret-😀 ")},
		{Config: storageConfig(api, "R2 renamed", r2, ""), Check: func(_ *terraform.State) error {
			_, _, tests, retains := api.StorageSnapshot()
			if tests != 2 || retains != 1 {
				return fmt.Errorf("metadata edit retested: tests=%d retentions=%d", tests, retains)
			}
			return nil
		}},
		{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"connection_info.cloudflare_r2.secret_access_key", "connection_info.cloudflare_r2.bucket", "connection_info.cloudflare_r2.key_prefix"}},
	}
	for _, v := range []struct{ provider, settings string }{
		{"backblaze_b2", `region="us-west-004"`},
		{"digitalocean_spaces", `region="nyc3"`},
		{"wasabi", `region="us-east-1"`},
		{"google_cloud_storage", ``},
		{"other_s3_compatible", `endpoint="https://storage.example.com/"`},
	} {
		connection := fmt.Sprintf("provider=%q\naccess_key_id=\"GOOG-key-id\"\nsecret_access_key=\"new-secret\"\n%s", v.provider, v.settings)
		config := storageConfig(api, v.provider, connection, "")
		steps = append(steps, resource.TestStep{Config: config, Check: resource.TestCheckResourceAttr(name, "id", "storage-test")}, resource.TestStep{Config: config, PlanOnly: true}, resource.TestStep{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"connection_info." + v.provider + ".secret_access_key", "connection_info." + v.provider + ".bucket", "connection_info." + v.provider + ".key_prefix", "connection_info." + v.provider + ".endpoint"}})
	}
	steps = append(steps, resource.TestStep{Config: storageConfig(api, "Back to AWS", aws, "hcti_storage_disabled=true"), Check: resource.TestCheckNoResourceAttr(name, "connection_info.cloudflare_r2")})
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: factories, Steps: steps})
	v, w, tests, _ := api.StorageSnapshot()
	if v != nil || w != 10 || tests != 8 {
		t.Fatalf("writes=%d tests=%d deleted=%v", w, tests, v == nil)
	}
}
func TestStorageValidation(t *testing.T) {
	api := testapi.New(t)
	for _, connection := range []string{
		`provider="unknown"`,
		`provider="aws_s3"
 region="invalid"
 role_arn="bad"`,
		`provider="google_cloud_storage"
 access_key_id="GOOG-test"`,
		`provider="google_cloud_storage"
 access_key_id="GOOG-test"
 secret_access_key=""`,
	} {
		resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"htmlcsstoimage": providerserver.NewProtocol6WithError(New("test")())}, Steps: []resource.TestStep{{Config: storageConfig(api, "Test storage", connection, ""), ExpectError: regexp.MustCompile("Invalid storage destination|Secret access key required|Unsupported attribute|Invalid Configuration|Missing Configuration")}}})
	}
	_, w, _, _ := api.StorageSnapshot()
	if w != 0 {
		t.Fatal("invalid plan wrote to API")
	}
}
