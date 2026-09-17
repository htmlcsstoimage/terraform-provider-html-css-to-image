---
page_title: "htmlcsstoimage_og_config Resource"
description: "Manage Open Graph image configurations using page HTML/CSS or templates."
---

# htmlcsstoimage_og_config

Manage an Open Graph image configuration for a website. Choose `html_css` to render pages using configured or extracted rendering options, or `templated` to fill a template from page metadata. The resource manages the configuration, not the individual images served through it.

See [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation) and the [complete example](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/og-config/main.tf). The API key requires `og_configs:read`, `og_configs:create_update`, and `og_configs:delete` for the full lifecycle.

## HTML/CSS configuration

```hcl
resource "htmlcsstoimage_og_config" "pages" {
  name        = "Website screenshots"
  config_type = "html_css"
  base_url    = "https://www.example.com"

  default_options = {
    selector        = "#social-card"
    viewport_width  = 1200
    viewport_height = 630
    css             = ".cookie-banner { display: none; }"
  }
}
```

With `extract_values = true`, image options extracted from page metadata override the configured defaults. Omission uses false. The source page comes from `base_url` and the serving path; do not put a page URL or inline HTML inside `default_options`.

## Template configuration

```hcl
resource "htmlcsstoimage_og_config" "articles" {
  name        = "Article cards"
  config_type = "templated"
  base_url    = "https://blog.example.com"
  template_id = "t-YOUR_TEMPLATE_ID"

  template_values_mapping = [
    { template_key = "headline", meta_key = "og:title" },
    { template_key = "summary", fallback = "descriptions" },
  ]
}
```

`template_version` is optional. Omission follows the latest version when an image is served. Set a positive version to pin it. Mapping order is preserved; each template field must be unique and cannot overlap another field's parent or child path. Each mapping selects exactly one `meta_key` or `fallback`. Fallbacks select page metadata (`titles` or `descriptions`), not literal replacement text.

## Inputs

| Field | Type | Behavior |
| --- | --- | --- |
| `name` | string, required | Display name; 1–255 characters after trimming. |
| `config_type` | string, required | `html_css` or `templated`. Updates in place. |
| `base_url` | string, required | HTTPS website origin, up to 2048 characters; no path, credentials, query, or fragment. A trailing slash is accepted. |
| `description` | string | Up to 1023 characters. Omission or blank clears it. |
| `disabled` | bool | Defaults to false. Disable serving through this configuration. |
| `optimization_mode` | string | `no_optimization`, `post_process`, or `set_viewport`. Omission leaves the default to the API (currently `post_process`). |
| `refresh_interval_s` | integer | 1800–31536000 seconds, subject to the account plan minimum. Omission leaves the default to the API (currently 86400). |
| `extract_values` | bool | HTML/CSS only. Extract page rendering options; omission uses false. |
| `default_options` | object | HTML/CSS only. Supported render settings below. |
| `template_id` | string | Required for `templated`; template ID including `t-`, up to 40 characters. |
| `template_version` | integer | Templated only. Positive int64, or omit to follow the latest version. |
| `template_values_mapping` | list(object) | Templated only; up to 32 ordered mappings. Each has `template_key` (up to 128 characters) and exactly one of `meta_key` (up to 136 characters) or `fallback`. |
| `headers` | map(string), sensitive | Templated only. Headers used to extract source page metadata. Omission clears them. |
| `additional_header_origins` | set(string) | Templated only. Additional exact origins allowed to receive those headers; requires at least one header. Omission clears the set. |

Text length limits follow the API's UTF-16 character counting. Header values and origins have separate UTF-8 byte limits.

### HTML/CSS rendering options

All fields below are optional within `default_options`. Unspecified rendering fields are sent as **JSON null**, including numeric and boolean fields. The provider does not insert rendering defaults. An omitted `default_options` object is also sent as null.

| Field | Type | Behavior |
| --- | --- | --- |
| `css` | string | CSS injected into the source page. Blank clears it. |
| `device_scale` | number | Pixel ratio, 0.1–3. |
| `max_wait_ms` | integer | Maximum loading wait, 500–10000 ms, subject to plan limits. |
| `ms_delay` | integer | Additional wait, 0–10000 ms. |
| `render_when_ready` | bool | Wait for the page's readiness signal. |
| `selector` | string | Crop to the selected element. |
| `viewport_width`, `viewport_height` | integer | 1–6000; supply both together. |
| `disable_twemoji` | bool | Use native emoji fonts. |
| `color_scheme` | string | `light` or `dark`. |
| `timezone` | string | IANA timezone; the API validates supported values. |
| `block_consent_banners` | bool | Attempt to block consent banners. |
| `identify_as_hcti` | bool | Identify top-level navigation using `X-HCTI-SCREENSHOT`. |
| `headers` | map(string), sensitive | Custom navigation headers. |
| `additional_header_origins` | set(string) | Additional exact origins allowed to receive headers; requires headers. |
| `include_headers_on_subrequests` | bool | Send headers on matching-origin subrequests; true requires headers. |
| `viewport_mobile`, `viewport_landscape`, `viewport_touch` | bool | Browser viewport/emulation settings. |
| `media_type` | string | `print` or `screen`. |
| `proxy_id` | string | Enabled rendering proxy ID. |
| `storage_destination_id` | string | Enabled storage destination ID that permits HCTI storage. |
| `transparent_background` | bool | Render with a transparent background. |

The template variant gets rendering settings, proxy, and storage from its template. Supplying `default_options` or `extract_values` with `templated` is rejected, as are template-only fields with `html_css`.

### Headers

Up to 20 headers are allowed, with unique case-insensitive ASCII token names of at most 512 characters and UTF-8 values of at most 8192 bytes. Values may contain tabs, but no other control characters. Up to 20 distinct additional HTTP(S) origins are allowed, each at most 512 UTF-8 bytes, without a path, credentials, query, fragment, or wildcard. The API also enforces restricted header names, origin restrictions, and reference eligibility.

For HTML/CSS, set headers inside `default_options`. For templates, use top-level `headers`:

```hcl
variable "source_authorization" {
  type      = string
  sensitive = true
}

resource "htmlcsstoimage_og_config" "private_articles" {
  name        = "Private article cards"
  config_type = "templated"
  base_url    = "https://blog.example.com"
  template_id = "t-YOUR_TEMPLATE_ID"
  headers = {
    Authorization = var.source_authorization
  }
}
```

Header maps are sensitive, including values recovered during import. They still persist in Terraform state. Updates send the complete desired header configuration; there are no header-retention flags. Removing the map clears it.

## Outputs

| Field | Meaning |
| --- | --- |
| `id` | Management ID used for updates, deletion, and import. |
| `domain_id` | Serving identifier used in OG image URLs; it can change when configuration settings change. |
| `effective_optimization_mode` | Optimization mode returned by the API. |
| `effective_refresh_interval_s` | Refresh interval returned by the API. |
| `created_at`, `updated_at` | API timestamps. |

The effective outputs expose server-selected values without populating omitted inputs with provider defaults.

## Updates, refresh, and deletion

All changes update in place, including changing `config_type`, source origin, or template. Remove fields belonging to the old variant when switching types. Refresh and import only fetch metadata; they never fetch the public serving URL, render an image, or test a proxy/storage connection.

If a plan downgrade permits disabling but ignores other edits, apply reports the unapplied field names and saves the actual server state. It does not report those edits as successful. The API enforces plan limits and validates referenced templates, proxies, and storage destinations.

Destroy soft-deletes/disables the configuration. It does not delete referenced templates, proxies, storage destinations, or rendered images, and does not promise immediate eviction of cached image bytes. Only HTTP 404 is treated as already absent; authorization, throttling, and server errors remain errors.

## Import

Use the management `id`, **not** `domain_id`:

```sh
terraform import htmlcsstoimage_og_config.pages OG_CONFIG_MANAGEMENT_ID
```

Import reconstructs the active variant, readable rendering settings, mappings, and headers. A null template version remains null. Review your configuration against imported settings before applying; omitted settings are cleared or return to API defaults on an update.
