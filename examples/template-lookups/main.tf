terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0"
    }
  }
}

provider "htmlcsstoimage" {}

variable "template_id" {
  type        = string
  description = "Existing HTML/CSS template ID, including its t- prefix."
}

data "htmlcsstoimage_template" "card" {
  id = var.template_id
}

data "htmlcsstoimage_template_versions" "card" {
  id = var.template_id
}

output "latest_version" {
  value = data.htmlcsstoimage_template.card.version
}

output "available_versions" {
  value = data.htmlcsstoimage_template_versions.card.versions[*].version
}
