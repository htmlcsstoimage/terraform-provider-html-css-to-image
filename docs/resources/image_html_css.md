---
page_title: "htmlcsstoimage_image_html_css Resource"
description: "Save an independently owned image definition without rendering it."
---

# htmlcsstoimage_image_html_css

Save an independently owned image definition without rendering it.

Creation saves the definition; bytes are rendered later. Every rendering-input change replaces the image with a new ID. The provider always sends `dedupe_duration_s: 0` so separate resources own separate images. Refresh uses authenticated metadata reads and never fetches the rendering URL. Delete accepts the API’s asynchronous deletion response.

The caller needs `images:create`, `images:read`, and `images:delete`, plus permission to use referenced proxies/storage destinations/templates. See [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation) and the [complete examples](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/images/main.tf).

```hcl
resource "htmlcsstoimage_image_html_css" "card" {
  html         = "<h1>Hello</h1>"
  css          = "h1 { font-family: Inter; padding: 48px; }"
  google_fonts = ["Inter"]
  format       = "webp"
}
```

## Inputs

Unspecified nullable rendering inputs are sent as explicit JSON null. The provider does not insert rendering defaults. Explicit `false` and `0` remain explicit; removing an input restores API-default behavior.

| Input | Required | Description |
| --- | --- | --- |
| `html` | Yes | HTML to render and take a screenshot of. HTML fragments are rendered in a wrapper document unless a complete HTML document is supplied. Required for HTML image requests. |
| `css` | No | CSS used to style the rendered HTML. |
| `device_scale` | No | Adjusts the pixel ratio used for the screenshot. Minimum: 0.1. Maximum: 3. HTML and template renders default to 2; URL renders default to 1. |
| `google_fonts` | No | Set of font family names, such as `["Inter", "Roboto"]`. Order is ignored; the provider sends pipe-delimited text. |
| `max_wait_ms` | No | Sets a limit on how long to wait before taking the screenshot when the page continues loading irrelevant content. Minimum: 500. Maximum: 10000 and subject to the account plan limit. |
| `metadata` | No | Custom key-value metadata stored with the image. |
| `ms_delay` | No | Adds extra time in milliseconds before taking the screenshot so JavaScript can execute. Minimum: 0. Maximum: 10000. |
| `render_when_ready` | No | Waits until the page signals that the screenshot is ready. The image fails if the readiness signal is never sent. |
| `selector` | No | A CSS selector for an element in the HTML. We’ll crop the image to this specific element. |
| `viewport_height` | No | Sets the height of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together. |
| `viewport_width` | No | Sets the width of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together. |
| `pdf_options` | No | Optional object documented below. |
| `disable_twemoji` | No | Disables the Twemoji fallback and renders emoji using native fonts instead. |
| `max_render_once` | No | Ensure the image is only ever rendered and saved one time. This is an advanced option not applicable to most requests. |
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
| `format` | No |  |

## PDF options

`pdf_options` accepts `page_width`, `page_height`, `scale` (0.1–2), `print_background`, and `margins`. Dimensions and margins use `px`, `in`, `cm`, `mm`, or `pt`. When setting margins, provide all four named sides; the provider sends them in top/right/bottom/left order. `print_background` is nonnullable in the API, so omission leaves it out of the request; other absent PDF fields are sent as null.

```hcl
pdf_options = {
  page_width  = "8.5in"
  page_height = "11in"
  margins = {
    top    = "12pt"
    right  = "12pt"
    bottom = "12pt"
    left   = "12pt"
  }
}
```

## Additional computed attributes

All inputs listed above are also readable resource attributes, for example `htmlcsstoimage_image_html_css.card.html`. Omitted nullable settings remain null rather than exposing effective rendering defaults. The following attributes are computed by the provider.

- `id`: unique image ID.
- `image_url`: rendering operation URL returned by creation.
- `public_url`: public GET rendering URL, or null for custom-storage-only output.
- `render_method` and `render_requires_auth`: how to invoke the operation. Custom-storage-only output uses **authenticated PUT** to `/v1/store/{id}`.
- `created_at`, `last_render_stored_at`, and `saved_to_storage_destination_at`: observed timestamps, not a completion guarantee.
- `og_config_id` and `og_config_content_version`: originating OG configuration metadata, if available.
- `storage_destination_hcti_storage_disabled`: whether output is stored only at the custom destination.

## Import

```sh
terraform import htmlcsstoimage_image_html_css.card IMAGE_ID
```

Import recovers the stored definition without rendering. The API does not store the requested `format`; it remains null after import. Adding an explicit format plans replacement. Keep HTML, URLs, headers, and template values in a protected state backend when they contain confidential content.
