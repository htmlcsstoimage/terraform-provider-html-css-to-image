terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0"
    }
  }
}

# Uses HCTI_API_ID and HCTI_API_KEY.
provider "htmlcsstoimage" {}

resource "htmlcsstoimage_image_html_css" "card" {
  html         = "<h1>Hello from Terraform</h1>"
  css          = "h1 { font-family: Inter; padding: 48px; background: #f1f5f9; }"
  google_fonts = ["Inter"]
  format       = "webp"
  metadata     = { application = "website" }
}

resource "htmlcsstoimage_image_url" "page" {
  url         = "https://example.com"
  full_screen = true
}

resource "htmlcsstoimage_template" "card" {
  html = "<h1>{{title}}</h1>"
  css  = "h1 { padding: 48px; }"
}

resource "htmlcsstoimage_image_templated" "card" {
  template_id = htmlcsstoimage_template.card.id
  # Optional: including this replaces the image when the template version changes.
  template_version = htmlcsstoimage_template.card.version
  template_values  = jsonencode({ title = "Hello from a template" })
}

output "image_url" {
  value = htmlcsstoimage_image_html_css.card.image_url
}
