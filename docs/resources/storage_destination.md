---
page_title: "htmlcsstoimage_storage_destination Resource"
description: "Manage a storage destination backed by an existing bucket or Space."
---

# htmlcsstoimage_storage_destination

Configure HCTI to write rendered files to an existing bucket or Space. The resource manages the HCTI configuration; it does not create or delete buckets, IAM roles, or stored files.

See [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation) and the [complete example](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/storage-destination/main.tf). The caller needs `storage_destinations:read`, `storage_destinations:create_update`, and `storage_destinations:delete` for the full lifecycle.

## Cloudflare R2 example

```hcl
variable "r2_secret_access_key" {
  type      = string
  sensitive = true
}

resource "htmlcsstoimage_storage_destination" "images" {
  name = "Rendered images"
  connection_info = {
    cloudflare_r2 = {
      bucket                = "rendered-images"
      cloudflare_account_id = "0123456789abcdef0123456789abcdef"
      access_key_id         = "YOUR_ACCESS_KEY_ID"
      secret_access_key     = var.r2_secret_access_key
      key_prefix            = "cards"
    }
  }
}
```

## AWS S3 example

AWS uses role assumption, not access-key credentials. Configure the role's trust policy using the organization's [AWS external ID](../data-sources/aws_storage_external_id.md) before creating the destination. The role and bucket must already exist and permit HCTI's writes.

```hcl
variable "storage_role_arn" {
  type = string
}

resource "htmlcsstoimage_storage_destination" "aws_images" {
  name = "AWS rendered images"
  connection_info = {
    aws_s3 = {
      bucket     = "rendered-images"
      region     = "us-east-1"
      role_arn   = var.storage_role_arn
      key_prefix = "cards"
    }
  }
}
```

If Terraform manages the role, reference its ARN and add a dependency on separately managed access policies so they are ready before the API tests the destination.

## Inputs

| Field | Type | Behavior |
| --- | --- | --- |
| `name` | string, required | Display name, 3–255 characters after trimming. |
| `disabled` | bool | Defaults to false. Prevent use for new renders. |
| `hcti_storage_disabled` | bool | Defaults to false. True stores output only at this destination, with no public HCTI CDN URL; rendering requires an authenticated `/store` request. The API rejects this for destinations referenced by enabled OG configurations. |
| `connection_info` | object, required | Provider-specific settings below. |

### Connection settings

Set exactly one provider-specific object inside `connection_info`: `aws_s3`, `cloudflare_r2`, `backblaze_b2`, `digitalocean_spaces`, `wasabi`, `google_cloud_storage`, or `other_s3_compatible`. There is no separate `provider` string. Each object exposes only its own fields, with required settings marked required in the schema.

Every provider object requires `bucket`. `bucket` is the existing bucket or Space name, up to 255 characters. Optional `key_prefix` is up to 1024 characters; omission uses the bucket root. The API trims surrounding whitespace and slashes from the prefix. Equivalent configured spelling is preserved in state, so `/cards/` does not produce a recurring diff after the API returns `cards`.

| Object in `connection_info` | Required fields | Optional fields |
| --- | --- | --- |
| `aws_s3` | `region`, `role_arn` | None beyond `key_prefix`. |
| `cloudflare_r2` | `cloudflare_account_id`, `access_key_id`, `secret_access_key` | `cloudflare_jurisdiction`: `eu` or `fedramp`; omission uses the default jurisdiction. |
| `backblaze_b2` | `region`, `access_key_id`, `secret_access_key` | None beyond `key_prefix`. |
| `digitalocean_spaces` | `region`, `access_key_id`, `secret_access_key` | None beyond `key_prefix`. |
| `wasabi` | `region`, `access_key_id`, `secret_access_key` | None beyond `key_prefix`. |
| `google_cloud_storage` | HMAC `access_key_id` beginning with `GOOG`, `secret_access_key` | None beyond `key_prefix`. |
| `other_s3_compatible` | `endpoint`, `access_key_id`, `secret_access_key` | `region` (API default `us-east-1`), `force_path_style` (API default true). |

Secrets are required on creation and credential-identity changes, with update/import exceptions below. Empty `connection_info` and multiple provider objects are rejected. Only fields for the selected provider are exposed. All variant changes update the same destination ID.

Additional limits:

- Regions for AWS, Backblaze, DigitalOcean, and Wasabi must match the supported API catalog. For custom S3-compatible services, the signing region is up to 64 characters.
- `role_arn` must be an AWS IAM role ARN, up to 2048 characters. AWS bucket names and prefixes cannot contain `*` or `?`.
- `cloudflare_account_id` is exactly 32 hexadecimal characters.
- `access_key_id` is up to 512 characters.
- `secret_access_key` must be nonempty after trimming, at most 990 UTF-16 units and 996 UTF-8 bytes. Empty strings do not mean retention.
- `endpoint` is a public HTTPS origin of up to 512 characters, without credentials, a path, query, or fragment. The API validates DNS and address eligibility; the provider does not perform network probes.

Omitted optional connection fields are sent as null. Defaults for custom S3-compatible region and path style are left to the API.

## Credentials and import

`connection_info.<provider>.secret_access_key` is sensitive and persists in Terraform state when supplied. The API never returns the secret.

- Creating an access-key destination requires a secret, including when disabled.
- Changing the provider or access key ID requires a supplied secret.
- Omitting the secret on an unrelated update retains the existing credential internally. No public retention flag is needed.
- When the configured secret matches the previous known value, the provider also sends retention rather than resupplying the secret. This avoids unnecessary API connection tests for metadata-only edits.
- Import reconstructs readable settings without inventing a secret. Imported credentials can be retained on unrelated updates.
- Switching to AWS removes the access-key fields. An out-of-band provider/access-key-ID change clears any locally remembered secret.

```sh
terraform import htmlcsstoimage_storage_destination.images STORAGE_DESTINATION_ID
```

After import, configure the readable connection settings. You can omit an existing secret until it needs changing. Omitting a previously configured secret also removes that value from subsequent Terraform state while retaining it remotely; historical state snapshots may still contain it.

## Connection tests and partial updates

Create and relevant updates can cause the API to write a test object to the bucket. Enabled destinations must pass. A disabled destination may be saved with a failed connection test; Terraform exposes the actual saved status. Refresh, import, and metadata-only changes do not initiate provider-side connection tests.

Following a plan downgrade, the API may accept disabling while ignoring other edits. The provider saves actual readable state and reports unapplied fields. Secret rotations cannot be read back, so it also checks for the connection-test result that a rotation requires; if a disable-only response does not confirm the rotation, it preserves the previous known secret state and reports the uncertainty.

Deletion removes the HCTI destination configuration. It does not delete the bucket, role, or stored images. Only HTTP 404 is treated as absent; authorization, throttling, malformed responses, and server errors remain errors. Writes are not automatically retried.

## Outputs

| Field | Meaning |
| --- | --- |
| `id` | Management ID, also the import ID and the value used by image/template/OG storage references. |
| `created_at`, `updated_at` | API timestamps. |
| `last_tested_at` | Latest connection-test timestamp, or null. |
| `last_test_succeeded` | Latest write-test result, or null. It does not guarantee future availability or read access. |
| `last_test_error` | Latest API test error, or null. Sensitive because it contains external-service diagnostic text. |
