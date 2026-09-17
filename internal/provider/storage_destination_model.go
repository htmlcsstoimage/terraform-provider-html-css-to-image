package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type storageModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Disabled            types.Bool   `tfsdk:"disabled"`
	HCTIStorageDisabled types.Bool   `tfsdk:"hcti_storage_disabled"`
	Connection          types.Object `tfsdk:"connection_info"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	LastTestedAt        types.String `tfsdk:"last_tested_at"`
	LastTestSucceeded   types.Bool   `tfsdk:"last_test_succeeded"`
	LastTestError       types.String `tfsdk:"last_test_error"`
}

func (m storageModel) known() bool {
	return !m.Name.IsUnknown() && !m.Disabled.IsUnknown() && !m.HCTIStorageDisabled.IsUnknown() && knownValue(m.Connection)
}
func storageString(c types.Object, k string) types.String {
	c = storageFlat(c)
	if c.IsNull() || c.IsUnknown() {
		return types.StringNull()
	}
	return c.Attributes()[k].(types.String)
}
func storageNormalized(k, s string) string {
	s = strings.TrimSpace(s)
	switch k {
	case "region", "cloudflare_account_id", "cloudflare_jurisdiction":
		return strings.ToLower(s)
	case "key_prefix":
		return strings.Trim(s, "/")
	case "endpoint":
		return strings.TrimRight(s, "/")
	}
	return s
}
func storageSameIdentity(a, b types.Object) bool {
	if a.IsNull() || a.IsUnknown() || b.IsNull() || b.IsUnknown() {
		return false
	}
	return storageString(a, "provider").Equal(storageString(b, "provider")) && storageNormalized("access_key_id", storageString(a, "access_key_id").ValueString()) == storageNormalized("access_key_id", storageString(b, "access_key_id").ValueString())
}
func (m storageModel) request(ctx context.Context, old *storageModel) (*management.StorageDestinationRequest, diag.Diagnostics) {
	d := m.validate(old)
	r := &management.StorageDestinationRequest{Name: m.Name.ValueString(), Disabled: m.Disabled.ValueBool(), HCTIStorageDisabled: m.HCTIStorageDisabled.ValueBool()}
	if d.HasError() {
		return r, d
	}
	a := storageFlat(m.Connection).Attributes()
	s := func(k string) types.String { return a[k].(types.String) }
	bucket := management.StorageBucket{Bucket: s("bucket").ValueString(), KeyPrefix: s("key_prefix").ValueStringPointer()}
	credentials := management.StorageCredentials{AccessKeyID: s("access_key_id").ValueString(), SecretAccessKey: s("secret_access_key").ValueStringPointer()}
	if old != nil && storageSameIdentity(m.Connection, old.Connection) {
		prior := storageString(old.Connection, "secret_access_key")
		if s("secret_access_key").IsNull() || (!prior.IsNull() && storageNormalized("secret_access_key", prior.ValueString()) == storageNormalized("secret_access_key", s("secret_access_key").ValueString())) {
			credentials.SecretAccessKey = nil
			credentials.RetainSecretAccessKey = hcti.Ptr(true)
		}
	}
	switch s("provider").ValueString() {
	case "aws_s3":
		r.ConnectionInfo = &management.AWSS3Connection{StorageBucket: bucket, Region: s("region").ValueString(), RoleARN: s("role_arn").ValueString()}
	case "cloudflare_r2":
		r.ConnectionInfo = &management.CloudflareR2Connection{StorageBucket: bucket, StorageCredentials: credentials, CloudflareAccountID: s("cloudflare_account_id").ValueString(), CloudflareJurisdiction: s("cloudflare_jurisdiction").ValueStringPointer()}
	case "backblaze_b2":
		r.ConnectionInfo = &management.BackblazeB2Connection{StorageBucket: bucket, StorageCredentials: credentials, Region: s("region").ValueString()}
	case "digitalocean_spaces":
		r.ConnectionInfo = &management.DigitalOceanSpacesConnection{StorageBucket: bucket, StorageCredentials: credentials, Region: s("region").ValueString()}
	case "wasabi":
		r.ConnectionInfo = &management.WasabiConnection{StorageBucket: bucket, StorageCredentials: credentials, Region: s("region").ValueString()}
	case "google_cloud_storage":
		r.ConnectionInfo = &management.GoogleCloudStorageConnection{StorageBucket: bucket, StorageCredentials: credentials}
	case "other_s3_compatible":
		r.ConnectionInfo = &management.OtherS3CompatibleConnection{StorageBucket: bucket, StorageCredentials: credentials, Endpoint: s("endpoint").ValueString(), Region: s("region").ValueStringPointer(), ForcePathStyle: a["force_path_style"].(types.Bool).ValueBoolPointer()}
	}
	return r, d
}
func storageEquivalent(k string, a, b attr.Value, provider string) bool {
	if a.Equal(b) {
		return true
	}
	if a.IsUnknown() || b.IsUnknown() {
		return false
	}
	if k == "force_path_style" {
		return a.IsNull() && provider == "other_s3_compatible" && !b.IsNull() && b.(types.Bool).ValueBool()
	}
	x, y := a.(types.String), b.(types.String)
	if k == "region" && provider == "other_s3_compatible" && x.IsNull() && storageNormalized(k, y.ValueString()) == "us-east-1" {
		return true
	}
	return storageNormalized(k, x.ValueString()) == storageNormalized(k, y.ValueString())
}
func (m *storageModel) read(r *management.StorageDestination, importing bool) {
	if importing || storageNormalized("name", m.Name.ValueString()) != r.Name {
		m.Name = types.StringValue(r.Name)
	}
	c := r.ConnectionInfo
	a := map[string]attr.Value{"provider": types.StringValue(string(c.Provider)), "bucket": types.StringValue(c.Bucket), "key_prefix": types.StringPointerValue(c.KeyPrefix), "region": types.StringPointerValue(c.Region), "role_arn": types.StringPointerValue(c.RoleARN), "access_key_id": types.StringPointerValue(c.AccessKeyID), "secret_access_key": types.StringNull(), "cloudflare_account_id": types.StringPointerValue(c.CloudflareAccountID), "cloudflare_jurisdiction": types.StringPointerValue(c.CloudflareJurisdiction), "endpoint": types.StringPointerValue(c.Endpoint), "force_path_style": types.BoolPointerValue(c.ForcePathStyle)}
	remote := types.ObjectValueMust(storageFlatTypes(), a)
	if !importing && !m.Connection.IsNull() && !m.Connection.IsUnknown() {
		for k, v := range storageFlat(m.Connection).Attributes() {
			if k != "secret_access_key" && storageEquivalent(k, v, a[k], string(c.Provider)) {
				a[k] = v
			}
		}
		if c.Provider != management.AWSS3 && storageSameIdentity(m.Connection, remote) {
			a["secret_access_key"] = storageString(m.Connection, "secret_access_key")
		}
	}
	m.Connection = storageNested(types.ObjectValueMust(storageFlatTypes(), a))
	m.ID = types.StringValue(r.ID)
	m.Disabled = types.BoolValue(!r.Enabled)
	m.HCTIStorageDisabled = types.BoolValue(r.HCTIStorageDisabled)
	m.CreatedAt = types.StringValue(r.CreatedAt.UTC().Format(time.RFC3339Nano))
	m.UpdatedAt = types.StringValue(r.UpdatedAt.UTC().Format(time.RFC3339Nano))
	m.LastTestedAt = types.StringNull()
	if r.LastTestedAt != nil {
		m.LastTestedAt = types.StringValue(r.LastTestedAt.UTC().Format(time.RFC3339Nano))
	}
	m.LastTestSucceeded = types.BoolPointerValue(r.LastTestSucceeded)
	m.LastTestError = types.StringPointerValue(r.LastTestError)
}
func storageUnapplied(want, got storageModel) []string {
	var fields []string
	if storageNormalized("name", want.Name.ValueString()) != storageNormalized("name", got.Name.ValueString()) {
		fields = append(fields, "name")
	}
	if !want.Disabled.Equal(got.Disabled) {
		fields = append(fields, "disabled")
	}
	if !want.HCTIStorageDisabled.Equal(got.HCTIStorageDisabled) {
		fields = append(fields, "hcti_storage_disabled")
	}
	kind := storageString(want.Connection, "provider").ValueString()
	actualKind := storageString(got.Connection, "provider").ValueString()
	if kind != actualKind {
		return append(fields, "connection_info")
	}
	actual := storageFlat(got.Connection).Attributes()
	for k, v := range storageFlat(want.Connection).Attributes() {
		if k != "secret_access_key" && !storageEquivalent(k, v, actual[k], storageString(got.Connection, "provider").ValueString()) {
			fields = append(fields, "connection_info."+kind+"."+k)
		}
	}
	return fields
}
