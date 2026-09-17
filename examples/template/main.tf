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

resource "htmlcsstoimage_template" "card" {
  name         = "Social card"
  description  = "Shared social image design"
  html         = "<article><h1>{{title}}</h1><p>{{subtitle}}</p></article>"
  css          = "article { width: 1200px; height: 630px; font-family: Inter; background: #f1f5f9; padding: 48px; }"
  google_fonts = ["Inter"]
}

output "template_id" {
  value = htmlcsstoimage_template.card.id
}
output "template_version" {
  value = htmlcsstoimage_template.card.version
}
