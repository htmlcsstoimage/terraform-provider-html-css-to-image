---
page_title: "Create images from HTML or a URL"
---

# Create images from HTML or a URL

Use Terraform to manage image definitions alongside your application infrastructure. This example creates an HTML/CSS card and a webpage screenshot.

## Configure the provider

Follow the [provider setup](../index.md) to configure the provider and set `HCTI_API_ID` and `HCTI_API_KEY`. The key needs `images:create`, `images:read`, and `images:delete` for this example.

## Define the images

Add these resources to your Terraform configuration:

```hcl
resource "htmlcsstoimage_image_html_css" "card" {
  html         = "<h1>Hello from Terraform</h1>"
  css          = "h1 { font-family: Inter; padding: 48px; background: #f1f5f9; }"
  google_fonts = ["Inter"]
  format       = "webp"
}

resource "htmlcsstoimage_image_url" "homepage" {
  url         = "https://example.com"
  full_screen = true
}

output "card_url" {
  value = htmlcsstoimage_image_html_css.card.image_url
}

output "homepage_url" {
  value = htmlcsstoimage_image_url.homepage.image_url
}
```

## Apply and render

```sh
terraform init
terraform plan
terraform apply
terraform output -raw card_url
```

Open the returned URL to render the card. `terraform apply` creates the image definition; it does not render bytes. The URL screenshot works the same way. A public rendering request can consume your image allowance.

This example uses HCTI storage. Custom-storage-only output instead requires an authenticated `PUT` to the returned rendering URL; see the [storage destination reference](../resources/storage_destination.md).

## Update an image

Change the HTML or CSS and run `terraform plan`. Rendering input changes replace the image definition with a new ID and URL. A refresh reads metadata without rendering the image again.

When finished with this example, `terraform destroy` deletes the managed image definitions. Deleting an image does not remove objects already written to your own storage.

See the [HTML/CSS image reference](../resources/image_html_css.md), [URL image reference](../resources/image_url.md), and [rendering parameter documentation](https://docs.htmlcsstoimage.com/parameters/) for more options.
