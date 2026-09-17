package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (*StorageDestinationResource) Metadata(_ context.Context, _ resource.MetadataRequest, r *resource.MetadataResponse) {
	r.TypeName = "htmlcsstoimage_storage_destination"
}
func (*StorageDestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, r *resource.SchemaResponse) {
	stable := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	r.Schema = schema.Schema{MarkdownDescription: "Manage an existing bucket's storage configuration. Updates preserve the destination ID. The API can test connections on create or relevant updates; reads only fetch metadata. Secrets persist in provider state and cannot be recovered by import.", Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "Destination management ID, used for import."},
		"name":                  schema.StringAttribute{Required: true, MarkdownDescription: "Display name, 3–255 characters after trimming."},
		"disabled":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Disable this destination for new renders. Defaults to false."},
		"hcti_storage_disabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Store output only at this destination. Defaults to false. True requires authenticated /store rendering and supplies no public HCTI CDN URL."},
		"connection_info":       schema.SingleNestedAttribute{Required: true, MarkdownDescription: "Set exactly one provider object: aws_s3, cloudflare_r2, backblaze_b2, digitalocean_spaces, wasabi, google_cloud_storage, or other_s3_compatible. Switching objects updates in place.", Attributes: storageConnectionAttributes()},
		"created_at":            schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "Creation timestamp."},
		"updated_at":            schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
		"last_tested_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Timestamp of the latest API connection test, or null. Refresh does not run tests."},
		"last_test_succeeded":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Latest API connection test result, or null."},
		"last_test_error":       schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Latest API connection test error, or null. Marked sensitive because it is external service diagnostic text."},
	}}
}
func storageFieldCatalog() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"provider":                schema.StringAttribute{Required: true, MarkdownDescription: "Storage provider: aws_s3, cloudflare_r2, backblaze_b2, digitalocean_spaces, wasabi, google_cloud_storage, or other_s3_compatible."},
		"bucket":                  schema.StringAttribute{Required: true, MarkdownDescription: "Existing bucket or Space name, up to 255 characters. The provider never creates or deletes the bucket."},
		"key_prefix":              schema.StringAttribute{Optional: true, MarkdownDescription: "Object key prefix, up to 1024 characters. The API trims surrounding whitespace and slashes; omission uses the bucket root."},
		"region":                  schema.StringAttribute{Optional: true, MarkdownDescription: "Required for AWS, Backblaze, DigitalOcean, and Wasabi. Optional signing region for other_s3_compatible, whose default is us-east-1."},
		"role_arn":                schema.StringAttribute{Optional: true, MarkdownDescription: "AWS only. IAM role ARN, up to 2048 characters. Configure the organization external ID in its trust policy before saving."},
		"access_key_id":           schema.StringAttribute{Optional: true, MarkdownDescription: "Required except for AWS. Up to 512 characters. Google Cloud Storage requires an HMAC access ID beginning with GOOG."},
		"secret_access_key":       schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Sensitive access key secret. Required on create or a change of provider/access key ID. Omit on unrelated updates to retain the existing secret. Empty is invalid. Up to 990 UTF-16 units and 996 UTF-8 bytes after trimming."},
		"cloudflare_account_id":   schema.StringAttribute{Optional: true, MarkdownDescription: "R2 only. Required 32-character hexadecimal account ID."},
		"cloudflare_jurisdiction": schema.StringAttribute{Optional: true, MarkdownDescription: "R2 only. eu or fedramp; omit for default jurisdiction."},
		"endpoint":                schema.StringAttribute{Optional: true, MarkdownDescription: "Required for other_s3_compatible. Public HTTPS origin, up to 512 characters, without credentials, path, query, or fragment."},
		"force_path_style":        schema.BoolAttribute{Optional: true, MarkdownDescription: "Other S3-compatible only. Omitted or null uses true. No value is injected into the request."},
	}
}
func storageConnectionTypes() map[string]attr.Type {
	r := map[string]attr.Type{}
	for k, a := range storageConnectionAttributes() {
		r[k] = a.GetType()
	}
	return r
}

// Each object exposes only its provider's fields and their actual requirements.
var storageVariants = map[string][]string{
	"aws_s3":               {"region", "role_arn"},
	"cloudflare_r2":        {"cloudflare_account_id", "cloudflare_jurisdiction"},
	"backblaze_b2":         {"region"},
	"digitalocean_spaces":  {"region"},
	"wasabi":               {"region"},
	"google_cloud_storage": {},
	"other_s3_compatible":  {"endpoint", "region", "force_path_style"},
}

func storageVariantAttributes(kind string) map[string]schema.Attribute {
	catalog := storageFieldCatalog()
	fields := map[string]schema.Attribute{"bucket": catalog["bucket"], "key_prefix": catalog["key_prefix"]}
	names := append([]string{}, storageVariants[kind]...)
	if kind != "aws_s3" {
		names = append(names, "access_key_id", "secret_access_key")
	}
	for _, name := range names {
		a := catalog[name]
		optional := name == "secret_access_key" || name == "cloudflare_jurisdiction" || name == "force_path_style" || (kind == "other_s3_compatible" && name == "region")
		if !optional {
			v := a.(schema.StringAttribute)
			v.Optional = false
			v.Required = true
			a = v
		}
		if name == "region" {
			v := a.(schema.StringAttribute)
			v.MarkdownDescription = "Bucket region supported by " + kind + "."
			if kind == "other_s3_compatible" {
				v.MarkdownDescription = "Optional signing region, up to 64 characters. Omitted or null uses the API default us-east-1."
			}
			a = v
		}
		fields[name] = a
	}
	return fields
}
func storageConnectionAttributes() map[string]schema.Attribute {
	fields := map[string]schema.Attribute{}
	for kind := range storageVariants {
		fields[kind] = schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "Connection settings for " + kind + ". Set exactly one provider object inside connection_info.", Attributes: storageVariantAttributes(kind)}
	}
	return fields
}
