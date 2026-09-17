terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0-beta.1"
    }
  }
}

provider "htmlcsstoimage" {}

variable "template_id" {
  type        = string
  description = "Existing HTML/CSS template ID, including its t- prefix."
}

variable "source_authorization" {
  type        = string
  sensitive   = true
  description = "Authorization header for the source website."
}

resource "htmlcsstoimage_og_config" "pages" {
  name        = "Website screenshots"
  config_type = "html_css"
  base_url    = "https://www.example.com"

  default_options = {
    selector        = "#social-card"
    viewport_width  = 1200
    viewport_height = 630
    headers = {
      Authorization = var.source_authorization
    }
  }
}

resource "htmlcsstoimage_og_config" "articles" {
  name        = "Article cards"
  config_type = "templated"
  base_url    = "https://blog.example.com"
  template_id = var.template_id
  # Omit template_version to follow the latest template at serving time.

  template_values_mapping = [
    { template_key = "headline", meta_key = "og:title" },
    { template_key = "summary", fallback = "descriptions" },
  ]
}

output "pages_domain_id" {
  value = htmlcsstoimage_og_config.pages.domain_id
}

output "articles_management_id" {
  value = htmlcsstoimage_og_config.articles.id
}
