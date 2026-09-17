---
page_title: "htmlcsstoimage_template Data Source"
---

# htmlcsstoimage_template

Read an existing HTML/CSS template without managing its lifecycle. Requires `templates:read`.

## Example Usage

```hcl
data "htmlcsstoimage_template" "card" {
  id = "t-YOUR_TEMPLATE_ID"
}

resource "htmlcsstoimage_image_templated" "card" {
  template_id      = data.htmlcsstoimage_template.card.id
  template_version = data.htmlcsstoimage_template.card.version
  template_values  = jsonencode({ title = "Hello from Terraform" })
}
```

Omitting `version` resolves the latest version on every refresh. In this example a new template version changes the image's `template_version`, so Terraform replaces the image. To pin the lookup, set `version` to a specific version identifier.

## Argument Reference

| Argument | Required | Description |
| --- | --- | --- |
| `id` | Yes | Existing template ID, including its `t-` prefix. |
| `version` | No | Exact version to read. Omit for latest; the resulting attribute contains the resolved version. |

## Attribute Reference

The data source exposes `id`, `version`, `template_type`, `created_at`, and `updated_at`, plus all saved [template settings](../resources/template.md#inputs): `html`, `css`, `name`, `description`, `google_fonts`, and rendering options such as `device_scale`, `viewport_width`, `viewport_height`, `proxy_id`, and `storage_destination_id`. These settings are computed outputs, not configurable inputs. Unset settings remain null; the lookup does not invent effective defaults.

Only HTML/CSS templates are supported by this lookup. Use [template_versions](template_versions.md) for metadata about versions of other template types. Missing templates or versions, permission errors, and API failures are reported as errors. The lookup never renders or deletes anything. Template contents are stored in Terraform state.
