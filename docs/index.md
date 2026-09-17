---
page_title: "HTML/CSS to Image Provider"
description: "Manage HTML/CSS to Image resources."
---

# HTML/CSS to Image Provider

Manage [HTML/CSS to Image](https://htmlcsstoimage.com) with Terraform. Define images and reusable templates alongside the API keys, proxies, storage destinations, and Open Graph configurations your application uses.

- [Website](https://htmlcsstoimage.com) — learn about HTML/CSS and URL rendering.
- [Product documentation](https://docs.htmlcsstoimage.com/) — rendering options, templates, and integration guides.
- [Management API](https://docs.htmlcsstoimage.com/management-api/) — API keys, proxies, storage destinations, and OG configurations.
- [Interactive API reference](https://htmlcsstoimage.com/api-docs/) — request and response schemas.
- [Examples](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples) and [issue tracker](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/issues).

## Example Usage

Set `HCTI_API_ID` and `HCTI_API_KEY` in the environment running Terraform, then add:

```hcl
terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0"
    }
  }
}

provider "htmlcsstoimage" {}

resource "htmlcsstoimage_image_html_css" "card" {
  html         = "<h1>Hello from Terraform</h1>"
  css          = "h1 { font-family: Inter; padding: 48px; color: #334155; }"
  google_fonts = ["Inter"]
}

output "image_url" {
  value = htmlcsstoimage_image_html_css.card.image_url
}
```

Run `terraform init`, `terraform plan`, and `terraform apply`.

## Resources

For step-by-step examples, see [creating images](guides/create-images.md), [using reusable templates](guides/template-images.md), and [managing application API keys](guides/application-api-keys.md).

| Resource | Use it to |
| --- | --- |
| [HTML/CSS image](resources/image_html_css.md) | Create an image definition from HTML and CSS. |
| [URL image](resources/image_url.md) | Create an image definition from a webpage URL. |
| [Templated image](resources/image_templated.md) | Create an image definition using a template and values. |
| [Template](resources/template.md) | Manage reusable HTML/CSS templates and their versions. |
| [API key](resources/api_key.md) | Manage application credentials and permissions. |
| [Proxy](resources/proxy.md) | Configure proxies for rendering requests. |
| [Storage destination](resources/storage_destination.md) | Deliver rendered images to your own object storage. |
| [OG configuration](resources/og_config.md) | Configure Open Graph image generation for your website. |

The [AWS storage lookup](data-sources/aws_storage_external_id.md) returns your organization's external ID and HCTI's writer role ARN for an IAM trust policy. See the [complete AWS example](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/aws-e2e) for S3, IAM, Secrets Manager, and image delivery.

To reference a template created outside Terraform, use the [template lookup](data-sources/template.md). The [template versions lookup](data-sources/template_versions.md) lists available versions and their metadata without managing them.

## Authentication

Use an API ID and key for the organization whose resources you want to manage. See [authentication and API keys](https://docs.htmlcsstoimage.com/getting-started/using-the-api/api-keys/) for setup.

The key needs the permissions listed on each resource's documentation page. For the image example above, it needs `images:create`, `images:read`, and `images:delete`. Keep the provider's credentials active throughout the resource lifecycle, including destroy. Do not commit credentials to source control.

## Argument Reference

| Argument | Description |
| --- | --- |
| `api_id` | Optional API ID; otherwise uses `HCTI_API_ID`. |
| `api_key` | Optional sensitive API key; otherwise uses `HCTI_API_KEY`. |
| `base_url` | Optional API base URL; defaults to `https://hcti.io`. |

## Image lifecycle and secrets

Image creation saves a rendering definition; it does not render image bytes. Image input changes replace the definition, while refresh reads metadata. Custom-storage-only images require an authenticated `PUT` to the returned rendering URL. The image resource pages describe the URLs and authentication requirements.

API keys, proxy passwords, storage credentials, and sensitive image inputs can be stored in Terraform state. Sensitive marking hides values from normal CLI output; use an access-controlled state backend.
