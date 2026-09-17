---
page_title: "htmlcsstoimage_image_templated Resource"
description: "Save an independently owned image definition without rendering it."
---

# htmlcsstoimage_image_templated

Save an independently owned image definition without rendering it.

Creation saves the definition; bytes are rendered later. Every rendering-input change replaces the image with a new ID. The provider always sends `dedupe_duration_s: 0` so separate resources own separate images. Refresh uses authenticated metadata reads and never fetches the rendering URL. Delete accepts the API’s asynchronous deletion response.

The caller needs `images:create`, `images:read`, and `images:delete`, plus permission to use referenced proxies/storage destinations/templates. See [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation) and the [complete examples](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/images/main.tf).

```hcl
resource "htmlcsstoimage_image_templated" "card" {
  template_id     = htmlcsstoimage_template.card.id
  template_values = jsonencode({ title = "Hello" })
}
```

`template_version` is optional. Omit it to use the latest version **at image creation**. A later template update does not replace an existing unpinned image. To replace the image whenever a managed template changes, set `template_version = htmlcsstoimage_template.card.version`. `resolved_template_version` reports the actual saved version without inserting a pin into the input.

`template_values` is a sensitive JSON object string; use `jsonencode`. Equality ignores object key ordering and whitespace while preserving large numbers. Sensitive values are still stored in Terraform state.

## Inputs

Unspecified nullable rendering inputs are sent as explicit JSON null. The provider does not insert rendering defaults. Explicit `false` and `0` remain explicit; removing an input restores API-default behavior.

| Input | Required | Description |
| --- | --- | --- |
| `template_values` | Yes | Nonempty JSON object string, preferably from `jsonencode`. Sensitive and stored in state. |
| `format` | No |  |
| `template_id` | Yes | Template ID including its t- prefix. |
| `template_version` | No | Positive int64 version. Omission selects latest at creation. |

## Additional computed attributes

All inputs listed above are also readable resource attributes, for example `htmlcsstoimage_image_templated.card.template_id`. Omitted nullable settings remain null rather than exposing effective rendering defaults. The following attributes are computed by the provider.

- `id`: unique image ID.
- `image_url`: rendering operation URL returned by creation.
- `public_url`: public GET rendering URL, or null for custom-storage-only output.
- `render_method` and `render_requires_auth`: how to invoke the operation. Custom-storage-only output uses **authenticated PUT** to `/v1/store/{id}`.
- `created_at`, `last_render_stored_at`, and `saved_to_storage_destination_at`: observed timestamps, not a completion guarantee.
- `og_config_id` and `og_config_content_version`: originating OG configuration metadata, if available.
- `storage_destination_hcti_storage_disabled`: whether HCTI storage was disabled when the image was created.
- `resolved_template_version`: version actually used to create the image.

## Import

```sh
terraform import htmlcsstoimage_image_templated.card IMAGE_ID
```

Import recovers the stored definition without rendering. The API does not store the requested `format`; it remains null after import. Adding an explicit format plans replacement. Keep HTML, URLs, headers, and template values in a protected state backend when they contain confidential content.

Import leaves `template_version` null and populates `resolved_template_version`. The rendering URL’s mode comes directly from the image’s saved `storage_destination_hcti_storage_disabled` metadata. No template or storage-destination reads are needed.
