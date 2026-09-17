---
page_title: "Create images with a reusable template"
---

# Create images with a reusable template

Manage a shared design as a template, then create images by supplying values for its Handlebars placeholders.

Follow the [provider setup](../index.md). The provider's API key needs `templates:read`, `templates:create_update`, `templates:delete`, `images:create`, `images:read`, and `images:delete` for this example.

## Define the template and image

```hcl
resource "htmlcsstoimage_template" "social_card" {
  name         = "Social card"
  html         = "<article><h1>{{title}}</h1><p>{{subtitle}}</p></article>"
  css          = "article { width: 1200px; height: 630px; padding: 48px; box-sizing: border-box; font-family: Inter; background: #f1f5f9; }"
  google_fonts = ["Inter"]
}

resource "htmlcsstoimage_image_templated" "announcement" {
  template_id      = htmlcsstoimage_template.social_card.id
  template_version = htmlcsstoimage_template.social_card.version
  template_values = jsonencode({
    title    = "Introducing our new release"
    subtitle = "Built with HTML/CSS to Image"
  })
}

output "announcement_url" {
  value = htmlcsstoimage_image_templated.announcement.image_url
}
```

Run `terraform init`, `terraform plan`, and `terraform apply`. Open the URL returned by `terraform output -raw announcement_url` to render the image.

## Choose how template updates affect images

The example explicitly references the managed template's version. Editing the template creates a new version, and Terraform replaces the image to use that version. Editing `template_values` also replaces the image.

You can omit `template_version` to select the latest template version when the image is created. Later template updates will not replace that existing image. The computed `resolved_template_version` attribute records the version used.

Use `jsonencode` to preserve nested JSON values and escape strings correctly. Template values are marked sensitive and are still stored in Terraform state.

See the [template reference](../resources/template.md) and [templated image reference](../resources/image_templated.md) for all inputs, outputs, and import behavior.

## Use an existing template

For a template created in the dashboard or through the API, use a data source. Terraform reads the template without taking ownership of it. The lookup needs `templates:read`; managing the image also needs the image permissions listed above.

```hcl
data "htmlcsstoimage_template" "existing" {
  id = "t-YOUR_TEMPLATE_ID"
}

resource "htmlcsstoimage_image_templated" "existing_card" {
  template_id      = data.htmlcsstoimage_template.existing.id
  template_version = data.htmlcsstoimage_template.existing.version
  template_values  = jsonencode({ title = "Hello from an existing template" })
}
```

The lookup resolves the latest version on each refresh. Because the image references that resolved version, a new template version causes Terraform to plan image replacement. Set `version` on the data source to a specific version identifier to pin it instead. See the [template data source](../data-sources/template.md).

## List recent template versions

Use `template_versions` to discover available versions and their metadata:

```hcl
data "htmlcsstoimage_template_versions" "recent" {
  id    = "t-YOUR_TEMPLATE_ID"
  limit = 10
}

output "recent_template_versions" {
  value = data.htmlcsstoimage_template_versions.recent.versions[*].version
}
```

This returns the newest 10 versions, or fewer if the template has less history. Omit `limit` to return up to **1000** versions. A supplied limit must be positive; increase it to retrieve more history.

Pagination is automatic. Each API request asks for at most 100 entries and no more than the remaining limit. Fetching stops when the limit is reached or the API has no more results. This limit only applies to the version listing; a pinned `template` lookup searches until it finds the requested version or exhausts the history.

Each entry includes its version identifier, name, description, template type, and timestamps. See the [template versions data source](../data-sources/template_versions.md) for the complete output reference.
