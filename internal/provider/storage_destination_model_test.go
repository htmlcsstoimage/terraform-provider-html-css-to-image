package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/testapi"
)

func storageObject(values map[string]attr.Value) types.Object {
	for k, t := range storageFlatTypes() {
		if _, ok := values[k]; !ok {
			values[k] = nullValue(t)
		}
	}
	return storageNested(types.ObjectValueMust(storageFlatTypes(), values))
}
func storageBase() storageModel {
	return storageModel{Name: types.StringValue("Test storage"), Disabled: types.BoolValue(false), HCTIStorageDisabled: types.BoolValue(false), Connection: storageObject(map[string]attr.Value{"provider": types.StringValue("google_cloud_storage"), "bucket": types.StringValue("bucket"), "access_key_id": types.StringValue("GOOG-test"), "secret_access_key": types.StringValue("secret")})}
}
func storageSet(m *storageModel, k string, v attr.Value) {
	a := storageFlat(m.Connection).Attributes()
	a[k] = v
	m.Connection = storageNested(types.ObjectValueMust(storageFlatTypes(), a))
}
func TestStorageRetentionAndSecretLimits(t *testing.T) {
	ctx := context.Background()
	old := storageBase()
	m := old
	body, d := m.request(ctx, &old)
	if d.HasError() {
		t.Fatal(d)
	}
	credentials := body.ConnectionInfo.(*management.GoogleCloudStorageConnection).StorageCredentials
	if credentials.SecretAccessKey != nil || credentials.RetainSecretAccessKey == nil || !*credentials.RetainSecretAccessKey {
		t.Fatal("unchanged secret was resent")
	}
	storageSet(&m, "secret_access_key", types.StringNull())
	body, d = m.request(ctx, &old)
	if d.HasError() {
		t.Fatal(d)
	}
	if body.ConnectionInfo.(*management.GoogleCloudStorageConnection).RetainSecretAccessKey == nil {
		t.Fatal("omitted secret not retained")
	}
	storageSet(&m, "access_key_id", types.StringValue("GOOG-changed"))
	if _, d = m.request(ctx, &old); !d.HasError() {
		t.Fatal("credential identity change accepted without secret")
	}
	m = old
	storageSet(&m, "provider", types.StringValue("other_s3_compatible"))
	storageSet(&m, "endpoint", types.StringValue("https://storage.example.com"))
	storageSet(&m, "secret_access_key", types.StringNull())
	if _, d = m.request(ctx, &old); !d.HasError() {
		t.Fatal("provider change accepted without secret")
	}
	for _, tc := range []struct {
		value string
		bad   bool
	}{{strings.Repeat("😀", 249), false}, {strings.Repeat("😀", 250), false}, {strings.Repeat("x", 990), false}, {strings.Repeat("x", 991), false}, {" ", true}, {string([]byte{0xff}), true}} {
		m = old
		storageSet(&m, "secret_access_key", types.StringValue(tc.value))
		if _, d = m.request(ctx, nil); d.HasError() != tc.bad {
			t.Fatalf("secret length boundary: bad=%v, diagnostics=%v", tc.bad, d)
		}
	}
	m = old
	storageSet(&m, "secret_access_key", types.StringUnknown())
	if m.known() {
		t.Fatal("unknown secret lost")
	}
	if _, d = m.request(ctx, &old); !d.HasError() {
		t.Fatal("unknown secret accepted at apply")
	}
}
func TestStorageImportPartialUpdateAndReadErrors(t *testing.T) {
	for _, rotateOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "combined edit", true: "secret only"}[rotateOnly], func(t *testing.T) {
			ctx := context.Background()
			api := testapi.New(t)
			c := management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))
			initial := storageBase()
			body, d := initial.request(ctx, nil)
			if d.HasError() {
				t.Fatal(d)
			}
			remote, err := c.CreateStorageDestination(ctx, body)
			if err != nil {
				t.Fatal(err)
			}
			r := &StorageDestinationResource{client: c}
			var schema resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schema)
			m := storageModel{ID: types.StringValue(remote.ID), Connection: types.ObjectNull(storageConnectionTypes())}
			state := tfsdk.State{Schema: schema.Schema}
			if d := state.Set(ctx, &m); d.HasError() {
				t.Fatal(d)
			}
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			if d := read.State.Get(ctx, &m); d.HasError() {
				t.Fatal(d)
			}
			if !storageString(m.Connection, "secret_access_key").IsNull() {
				t.Fatal("import invented secret")
			}
			// An imported destination can be renamed without knowing its secret.
			m.Name = types.StringValue("Renamed import")
			plan := tfsdk.Plan{Schema: schema.Schema}
			if d := plan.Set(ctx, &m); d.HasError() {
				t.Fatal(d)
			}
			updated := resource.UpdateResponse{State: read.State}
			r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: plan}, &updated)
			if updated.Diagnostics.HasError() {
				t.Fatal(updated.Diagnostics)
			}
			_, _, tests, retains := api.StorageSnapshot()
			if tests != 1 || retains != 1 {
				t.Fatal("import metadata update retested connection")
			}
			if d := updated.State.Get(ctx, &m); d.HasError() {
				t.Fatal(d)
			}
			m.Disabled = types.BoolValue(true)
			storageSet(&m, "secret_access_key", types.StringValue("changed-secret"))
			if !rotateOnly {
				m.Name = types.StringValue("Ignored name")
				storageSet(&m, "bucket", types.StringValue("ignored-bucket"))
			}
			if d := plan.Set(ctx, &m); d.HasError() {
				t.Fatal(d)
			}
			api.StorageMode(true, false)
			partial := resource.UpdateResponse{State: updated.State}
			r.Update(ctx, resource.UpdateRequest{State: updated.State, Plan: plan}, &partial)
			if !partial.Diagnostics.HasError() {
				t.Fatal("unconfirmed update reported success")
			}
			if d := partial.State.Get(ctx, &m); d.HasError() {
				t.Fatal(d)
			}
			if !m.Disabled.ValueBool() || m.Name.ValueString() != "Renamed import" || !storageString(m.Connection, "secret_access_key").IsNull() {
				t.Fatal("partial update did not save actual state")
			}
			for _, d := range partial.Diagnostics {
				if strings.Contains(d.Detail(), "changed-secret") {
					t.Fatal("secret leaked in diagnostic")
				}
			}
			for _, status := range []int{403, 429, 500} {
				api.Status(status)
				resp := resource.ReadResponse{State: partial.State}
				r.Read(ctx, resource.ReadRequest{State: partial.State}, &resp)
				if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(partial.State.Raw) {
					t.Fatalf("HTTP %d changed state", status)
				}
			}
			api.Status(0)
			api.StorageMalformed(`{}`)
			malformed := resource.ReadResponse{State: partial.State}
			r.Read(ctx, resource.ReadRequest{State: partial.State}, &malformed)
			if !malformed.Diagnostics.HasError() || !malformed.State.Raw.Equal(partial.State.Raw) {
				t.Fatal("malformed response changed state")
			}
			api.StorageMalformed("")
			api.Status(404)
			gone := resource.ReadResponse{State: partial.State}
			r.Read(ctx, resource.ReadRequest{State: partial.State}, &gone)
			if gone.Diagnostics.HasError() || !gone.State.Raw.IsNull() {
				t.Fatal("404 did not remove state")
			}
			api.Status(0)
			if err := c.DeleteStorageDestination(ctx, remote.ID); err != nil {
				t.Fatal(err)
			}
			deleted := resource.DeleteResponse{}
			r.Delete(ctx, resource.DeleteRequest{State: partial.State}, &deleted)
			if deleted.Diagnostics.HasError() {
				t.Fatal("repeat delete failed")
			}
		})
	}
}
func TestStorageDisabledFailedConnection(t *testing.T) {
	ctx := context.Background()
	api := testapi.New(t)
	api.StorageMode(false, true)
	c := management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))
	r := &StorageDestinationResource{client: c}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	m := storageBase()
	m.Disabled = types.BoolValue(true)
	plan := tfsdk.Plan{Schema: schema.Schema}
	if d := plan.Set(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	result := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &result)
	if result.Diagnostics.HasError() {
		t.Fatal(result.Diagnostics)
	}
	if d := result.State.Get(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	if m.LastTestSucceeded.IsNull() || m.LastTestSucceeded.ValueBool() || m.LastTestError.IsNull() {
		t.Fatal("failed test result not exposed")
	}
	// Read never retries the connection test.
	read := resource.ReadResponse{State: result.State}
	r.Read(ctx, resource.ReadRequest{State: result.State}, &read)
	if read.Diagnostics.HasError() {
		t.Fatal(read.Diagnostics)
	}
	_, _, tests, _ := api.StorageSnapshot()
	if tests != 1 {
		t.Fatal("read retested")
	}
}
func TestAWSExternalIDErrors(t *testing.T) {
	ctx := context.Background()
	api := testapi.New(t)
	r := &AWSStorageExternalIDDataSource{client: management.NewClient("test-id", "test-key", management.WithBaseURL(api.URL), management.WithUserAgentSuffix("HCTITerraform/test"))}
	var schema datasource.SchemaResponse
	r.Schema(ctx, datasource.SchemaRequest{}, &schema)
	for _, status := range []int{403, 404, 429, 500} {
		api.Status(status)
		resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
		r.Read(ctx, datasource.ReadRequest{}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("lookup swallowed API error")
		}
	}
}

func TestStorageProviderObjectSchema(t *testing.T) {
	for kind := range storageVariants {
		fields := storageVariantAttributes(kind)
		if !fields["bucket"].IsRequired() {
			t.Fatal("bucket is not required")
		}
		if _, ok := fields["provider"]; ok {
			t.Fatal("redundant discriminator exposed")
		}
		for _, name := range storageVariants[kind] {
			optional := name == "cloudflare_jurisdiction" || name == "force_path_style" || (kind == "other_s3_compatible" && name == "region")
			if fields[name].IsRequired() == optional {
				t.Fatalf("%s.%s has incorrect required status", kind, name)
			}
		}
		if kind != "aws_s3" && (!fields["access_key_id"].IsRequired() || !fields["secret_access_key"].IsOptional() || !fields["secret_access_key"].IsSensitive()) {
			t.Fatalf("%s credentials schema incorrect", kind)
		}
	}
	m := storageBase()
	a := map[string]attr.Value{}
	for kind, typ := range storageConnectionTypes() {
		a[kind] = nullValue(typ)
	}
	m.Connection = types.ObjectValueMust(storageConnectionTypes(), a)
	if !m.validate(nil).HasError() {
		t.Fatal("empty union accepted")
	}
	child := storageBase().Connection.Attributes()["google_cloud_storage"]
	a["google_cloud_storage"] = child
	aws := storageObject(map[string]attr.Value{"provider": types.StringValue("aws_s3"), "bucket": types.StringValue("bucket"), "region": types.StringValue("us-east-1"), "role_arn": types.StringValue("arn:aws:iam::123456789012:role/hcti")})
	a["aws_s3"] = aws.Attributes()["aws_s3"]
	m.Connection = types.ObjectValueMust(storageConnectionTypes(), a)
	if !m.validate(nil).HasError() {
		t.Fatal("multiple providers accepted")
	}
	a["aws_s3"] = nullValue(storageConnectionTypes()["aws_s3"])
	a["google_cloud_storage"] = types.ObjectUnknown(child.(types.Object).AttributeTypes(context.Background()))
	m.Connection = types.ObjectValueMust(storageConnectionTypes(), a)
	if m.known() {
		t.Fatal("unknown provider object lost")
	}
}
