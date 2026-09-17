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

variable "proxy_password" {
  type      = string
  sensitive = true
}

resource "htmlcsstoimage_proxy" "rendering" {
  name = "Rendering proxy"
  url  = "https://proxy.example.com"
  port = 8080

  authentication = {
    username = "rendering"
    password = var.proxy_password
  }

  bypass_hosts = ["static.example.com"]
}

output "proxy_id" {
  value = htmlcsstoimage_proxy.rendering.id
}
