---
page_title: "htmlcsstoimage_api_key Resource"
description: "Manage an HTML/CSS to Image API key and its permissions."
---

# htmlcsstoimage_api_key

Create an API key with the permissions needed by an application or automation. Updates change its settings without rotating its secret. Destroy disables the key; it does not permanently delete its metadata.

See [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation) for installation and credentials. Use a separate management key to manage application keys: disabling the credentials used by the provider can prevent later operations.

## Example

```hcl
resource "htmlcsstoimage_api_key" "rendering" {
  name        = "Rendering service"
  description = "Create images and read templates"
  permissions = ["images:create", "templates:read"]
}

output "rendering_api_id" {
  value = htmlcsstoimage_api_key.rendering.api_id
}

output "rendering_api_key" {
  value     = htmlcsstoimage_api_key.rendering.api_key
  sensitive = true
}
```

`id` is the management ID used for import and CRUD. `api_id` is the authentication ID used with `api_key` to call the API. These IDs are distinct.

## Inputs

| Argument | Type | Description |
| --- | --- | --- |
| `name` | Optional string | Display name, up to 255 characters. Omitted, null, or blank uses `Key created YYYY-MM-DD HH:mm:ss`, based on the key's original creation time in UTC. |
| `description` | Optional string | Purpose of the key, up to 2000 characters. Omitted, null, or blank clears it. |
| `disabled` | Optional boolean | Prevent the key from authenticating. Defaults to `false`. |
| `all_future_permissions` | Optional boolean | Grant all current and future permissions. Defaults to `false`; the caller must have this authority. |
| `permissions` | Set of strings | Complete grant list. Required unless `all_future_permissions` is true. `[]` grants no product permissions. With future grants enabled, omit this field or supply `[]`. |

Permission order and duplicates do not affect the requested grant set. Known permission names are listed below; the API validates grants against the caller's authority.

Omitted names remain unset in configuration. `effective_name` exposes the actual server-selected name, including generated names. Removing a previously configured name resets it to the API-generated name on the next update. Updates also replace the description, disabled state, and permission settings.

## Outputs

| Output | Type | Description |
| --- | --- | --- |
| `id` | String | Management ID used for import and key management operations. |
| `api_id` | String | Authentication ID for the created key. |
| `api_key` | Sensitive string | Secret returned only at creation. Preserved during refresh and updates; unavailable after import. |
| `effective_name` | String | Actual display name returned by the API. |
| `effective_permissions` | Set of strings | Current grants reported by the API, including expansion of all-future access. |
| `created_at`, `updated_at` | String | Creation and last-update timestamps in RFC 3339 format. |

## Permissions

| Prefix | Supported operations |
| --- | --- |
| `api_keys` | `create_update`, `delete`, `read` |
| `images` | `create`, `delete`, `read`, `store` |
| `templates` | `create_update`, `delete`, `read` |
| `proxies` | `create_update`, `delete`, `read` |
| `storage_destinations` | `create_update`, `delete`, `read` |
| `og_configs` | `create_update`, `delete`, `read` |
| `usage` | `read` |

Combine the prefix and operation with a colon, for example `images:create`. Granting `usage:read` is supported even though these providers do not offer a usage lookup resource.

The managing key needs `api_keys:read`, `api_keys:create_update`, and `api_keys:delete` for the full lifecycle. It must also be authorized to grant the key's existing and requested permissions. These authority checks apply to disabling, re-enabling, and destroying keys as well as ordinary updates.

### All current and future permissions

```hcl
resource "htmlcsstoimage_api_key" "automation" {
  name                   = "Automation"
  all_future_permissions = true
}
```

This is an explicit opt-in to permissions added in the future. `effective_permissions` shows the current grants, while `permissions` stays omitted or empty. Newly added server permissions do not cause a provider update.

To leave this mode, set `all_future_permissions` to false and provide an explicit `permissions` list. Use `[]` if the key should have no product permissions.

## Secret lifecycle

The API returns the secret once, during creation. The provider saves it in state and preserves it through updates and refreshes. A read never retrieves or rotates a secret.

Terraform marks the secret sensitive, but it remains stored in Terraform state. Sensitive outputs must also be marked `sensitive = true`.

Changing the name, description, permissions, or disabled state does not rotate the secret. To obtain a new secret, explicitly replace the resource. Replacement creates a new key and disables the old one as part of removing it; update consumers to use the new credentials.

The provider does not retry key creation automatically. If the server creates a key but its response is lost, the secret cannot be recovered by reading or importing the key.

## Import

Import by management `id`, **not** `api_id`:

```sh
terraform import htmlcsstoimage_api_key.rendering MANAGEMENT_ID
```

Import restores readable settings and authentication ID, but `api_key` is null. No placeholder secret is generated. Later updates keep the secret unavailable in state; they do not rotate or recover it.

Match your configuration to the imported name, description, disabled state, and permission mode. In all-future mode, keep `permissions` omitted or empty and inspect `effective_permissions` for the current grants. Run `terraform plan` before applying changes.

## Disable, re-enable, and destroy

Set `disabled` to true to stop the key from authenticating while keeping it managed. Set it to false to re-enable it, subject to the caller's permissions.

Destroy disables the key and removes it from provider state. The API can still return its metadata afterward. A disabled key is not treated as missing on refresh or import; it can be imported again and re-enabled deliberately.

Only HTTP 404 removes a key during refresh. Authorization, rate-limit, malformed-response, and other errors preserve state and report an error. Preview and unchanged plans do not create or update keys.
