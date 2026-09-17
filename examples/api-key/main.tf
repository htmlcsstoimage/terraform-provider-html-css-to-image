terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0-beta.2"
    }
  }
}

# Uses HCTI_API_ID and HCTI_API_KEY to manage the new key.
provider "htmlcsstoimage" {}

resource "htmlcsstoimage_api_key" "rendering" {
  name        = "Rendering service"
  description = "Create images and read templates"
  permissions = ["images:create", "templates:read"]
}

output "rendering_api_id" {
  value = htmlcsstoimage_api_key.rendering.api_id
}

output "rendering_api_key" {
  value     = htmlcsstoimage_api_key.rendering.api_key
  sensitive = true
}
