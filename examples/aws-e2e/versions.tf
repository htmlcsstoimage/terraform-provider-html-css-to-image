terraform {
  required_version = ">= 1.5.0"
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0-beta.2"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.13"
    }
  }
}

provider "aws" {
  region = var.aws_region
  default_tags {
    tags = { Application = var.name, ManagedBy = "Terraform", Purpose = "HCTI end-to-end example" }
  }
}

# Use existing HCTI_API_ID/HCTI_API_KEY bootstrap credentials throughout the lifecycle.
provider "htmlcsstoimage" {
  base_url = var.hcti_base_url
}
