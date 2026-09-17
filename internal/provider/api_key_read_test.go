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

func TestAPIKeyImportUpdateAndReadErrors(t *testing.T) {
	ctx := context.Background()
	api := testapi.New(t)
	c := management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))
	key, err := c.CreateAPIKey(ctx, &management.APIKeyRequest{Permissions: []management.Permission{}})
	if err != nil {
		t.Fatal(err)
	}
	r := &APIKeyResource{client: c}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	model := apiKeyModel{ID: types.StringValue(key.ID), Permissions: types.SetNull(types.StringType), EffectivePermissions: types.SetNull(types.StringType)}
	state := tfsdk.State{Schema: schema.Schema}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	read := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &read)
	if read.Diagnostics.HasError() {
		t.Fatal(read.Diagnostics)
	}
	if d := read.State.Get(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	if !model.APIKey.IsNull() || model.APIID.ValueString() != "key-auth-id" {
		t.Fatal("import fabricated secret or lost authentication ID")
	}
	model.Name = types.StringValue("renamed imported key")
	model.Disabled = types.BoolValue(true)
	plan := tfsdk.Plan{Schema: schema.Schema}
	if d := plan.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	updated := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: plan}, &updated)
	if updated.Diagnostics.HasError() {
		t.Fatal(updated.Diagnostics)
	}
	if d := updated.State.Get(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	if !model.APIKey.IsNull() {
		t.Fatal("updating an imported key fabricated a secret")
	}
	// A disabled key is still readable and must remain managed.
	read = resource.ReadResponse{State: updated.State}
	r.Read(ctx, resource.ReadRequest{State: updated.State}, &read)
	if read.Diagnostics.HasError() || read.State.Raw.IsNull() {
		t.Fatal("disabled key treated as deleted")
	}
	for _, status := range []int{403, 429, 500} {
		api.Status(status)
		resp := resource.ReadResponse{State: updated.State}
		r.Read(ctx, resource.ReadRequest{State: updated.State}, &resp)
		if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(updated.State.Raw) {
			t.Fatalf("HTTP %d did not preserve state", status)
		}
	}
	api.Status(0)
	api.KeyBody(`{}`)
	malformed := resource.ReadResponse{State: updated.State}
	r.Read(ctx, resource.ReadRequest{State: updated.State}, &malformed)
	if !malformed.Diagnostics.HasError() || !malformed.State.Raw.Equal(updated.State.Raw) {
		t.Fatal("malformed response did not preserve state")
	}
	api.KeyBody("")
	api.Status(404)
	gone := resource.ReadResponse{State: updated.State}
	r.Read(ctx, resource.ReadRequest{State: updated.State}, &gone)
	if gone.Diagnostics.HasError() || !gone.State.Raw.IsNull() {
		t.Fatal("404 did not remove state")
	}
}
