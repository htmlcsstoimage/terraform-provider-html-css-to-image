terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0"
    }
  }
}

provider "htmlcsstoimage" {}

variable "aws_bucket" {
  type        = string
  description = "Existing S3 bucket."
}

variable "aws_role_arn" {
  type        = string
  description = "Existing role with HCTI trust and bucket write permissions."
}

data "htmlcsstoimage_aws_storage_external_id" "organization" {}

resource "htmlcsstoimage_storage_destination" "images" {
  name = "Rendered images"
  connection_info = {
    aws_s3 = {
      bucket     = var.aws_bucket
      region     = "us-east-1"
      role_arn   = var.aws_role_arn
      key_prefix = "cards"
    }
  }
}

output "storage_destination_id" {
  value = htmlcsstoimage_storage_destination.images.id
}

output "aws_external_id" {
  value = data.htmlcsstoimage_aws_storage_external_id.organization.external_id
}
