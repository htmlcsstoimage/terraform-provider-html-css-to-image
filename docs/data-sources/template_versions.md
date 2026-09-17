---
page_title: "htmlcsstoimage_template_versions Data Source"
---

# htmlcsstoimage_template_versions

List versions of an existing template, newest first, up to `limit` (default `1000`). Requires `templates:read`. The provider follows API pagination automatically and preserves exact integer version identifiers.

## Example Usage

```hcl
data "htmlcsstoimage_template_versions" "card" {
  id    = "t-YOUR_TEMPLATE_ID"
  limit = 10
}

output "available_versions" {
  value = data.htmlcsstoimage_template_versions.card.versions[*].version
}
```

## Argument Reference

| Argument | Required | Description |
| --- | --- | --- |
| `id` | Yes | Existing template ID, including its `t-` prefix. |
| `limit` | No | Maximum versions to return. Defaults to `1000`; must be positive. |

## Attribute Reference

`limit` reports the configured or default limit. Pagination stops at this limit or when the API has no more results; each request asks for at most 100 entries and no more than the remaining limit.

`versions` is a list of metadata objects, each containing:

| Attribute | Description |
| --- | --- |
| `version` | Exact integer version identifier. |
| `name` | Display name for this version. |
| `description` | Description for this version. |
| `template_type` | Type reported by the API, such as `html_css` or `blocks`. |
| `created_at` | Creation timestamp returned by the API. |
| `updated_at` | Update timestamp returned by the API. |

Template contents and rendering settings are excluded from this list. Use [template](template.md) to read a particular HTML/CSS version's settings or track the latest version. An empty successful API result produces an empty list; HTTP 404 and other API failures remain errors. Listing versions never creates, updates, renders, or deletes templates.
