package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/testapi"
)

func TestReadErrorsPreserveState(t *testing.T) {
	api := testapi.New(t)
	ctx := context.Background()
	client := management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))
	proxy, err := client.CreateProxy(ctx, &management.ProxyRequest{Name: "test proxy", URL: "https://proxy.example"})
	if err != nil {
		t.Fatal(err)
	}
	r := &ProxyResource{client: client}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	model := proxyModel{Port: types.Int64Null(), BypassHosts: types.SetNull(types.StringType), Authentication: types.ObjectNull(authenticationTypes)}
	if d := model.read(ctx, proxy, false); d.HasError() {
		t.Fatal(d)
	}
	state := tfsdk.State{Schema: schema.Schema}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	for _, status := range []int{403, 429, 500} {
		api.Status(status)
		resp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &resp)
		if !resp.Diagnostics.HasError() || resp.State.Raw.IsNull() {
			t.Fatalf("HTTP %d removed state or didn't fail", status)
		}
	}
	api.Status(404)
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatal("404 did not remove state")
	}
}
