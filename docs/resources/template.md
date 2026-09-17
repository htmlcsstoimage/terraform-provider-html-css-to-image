---
page_title: "htmlcsstoimage_template Resource"
description: "Manage an HTML/CSS template and its versions."
---

# htmlcsstoimage_template

Manage an HTML/CSS template and its versions.

Changes, including name and description, create a new version under the same ID. Refresh reads the latest version and detects out-of-band edits. Deleting this resource deletes **all versions** of the template. Do not manage individual versions as separate resources. Cached images can remain accessible after deletion.

The caller needs `templates:read`, `templates:create_update`, and `templates:delete`. See [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation) and the [complete example](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/template/main.tf).

```hcl
resource "htmlcsstoimage_template" "card" {
  name         = "Social card"
  html         = "<h1>{{title}}</h1>"
  css          = "h1 { font-family: Inter; padding: 48px; }"
  google_fonts = ["Inter"]
}
```

## Inputs

Unspecified nullable rendering inputs are sent as explicit JSON null. The provider does not insert rendering defaults. Explicit `false` and `0` remain explicit; removing an input restores API-default behavior.

| Input | Required | Description |
| --- | --- | --- |
| `html` | Yes | HTML to render for the template. Use Handlebars placeholders for values that will be substituted when an image is rendered. Must be non-empty, contain at least one Handlebars placeholder, and compile as valid Handlebars. |
| `name` | No | The name of the template, used to identify it in your account. Maximum: 64 characters. |
| `description` | No | An optional description of the template for your reference. Maximum: 1024 characters. |
| `css` | No | CSS used to style images rendered from the template. Handlebars expressions are not supported in CSS; put dynamic CSS in the HTML instead. |
| `device_scale` | No | Adjusts the pixel ratio used for the screenshot. Minimum: 0.1. Maximum: 3. HTML and template renders default to 2; URL renders default to 1. |
| `google_fonts` | No | Set of font family names, such as `["Inter", "Roboto"]`. Order is ignored; the provider sends pipe-delimited text. |
| `max_wait_ms` | No | Sets a limit on how long to wait before taking the screenshot when the page continues loading irrelevant content. Minimum: 500. Maximum: 10000 and subject to the account plan limit. |
| `ms_delay` | No | Adds extra time in milliseconds before taking the screenshot so JavaScript can execute. Minimum: 0. Maximum: 10000. |
| `render_when_ready` | No | Waits until the page signals that the screenshot is ready. The image fails if the readiness signal is never sent. |
| `selector` | No | A CSS selector for an element in the HTML. We’ll crop the image to this specific element. |
| `viewport_height` | No | Sets the height of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together. |
| `viewport_width` | No | Sets the width of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together. |
| `disable_twemoji` | No | Disables the Twemoji fallback and renders emoji using native fonts instead. |
| `color_scheme` | No |  |
| `timezone` | No | Sets the IANA timezone used by Chrome while rendering. Must be a recognized IANA timezone identifier. |
| `viewport_mobile` | No | Specifies whether the page uses mobile viewport behavior, including its viewport meta tag. |
| `viewport_landscape` | No | Specifies whether the emulated viewport is in landscape orientation. |
| `viewport_touch` | No | Specifies whether the emulated viewport supports touch events. |
| `media_type` | No |  |
| `proxy_id` | No | Specifies which configured organization proxy to use when rendering. |
| `storage_destination_id` | No | Specifies which configured organization storage destination receives the rendered image. |
| `jumbo_max_height` | No | Maximum height of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_width. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels. |
| `jumbo_max_width` | No | Maximum width of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_height. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels. |
| `transparent_background` | No | Specifies whether the image is rendered with a transparent background. |

## Additional computed attributes

All inputs listed above are also readable resource attributes, for example `htmlcsstoimage_template.card.html`. Omitted nullable settings remain null rather than exposing effective rendering defaults. The following attributes are computed by the provider.

`id` is the stable template ID. `version` is the latest saved int64 version. `template_type` is `html_css`; `created_at` and `updated_at` are API timestamps.

## Import

```sh
terraform import htmlcsstoimage_template.card t-TEMPLATE_ID
```

Import reads the latest HTML/CSS version. Block templates are rejected because this resource cannot edit their design.
