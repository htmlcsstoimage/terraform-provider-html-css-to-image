---
page_title: "Manage an application API key"
---

# Manage an application API key

Give an application its own API key with the permissions it needs. Use separate credentials for Terraform so application key changes do not interrupt infrastructure management.

## Configure management credentials

Follow the [provider setup](../index.md) using an existing key authorized to read and manage API keys and grant the requested permissions. See [API key management](https://docs.htmlcsstoimage.com/management-api/api-keys/) for the authorization rules.

## Create a restricted key

```hcl
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
```

Run `terraform init`, `terraform plan`, and `terraform apply`. Your application authenticates with `api_id` and `api_key`. The resource's `id` is a separate management ID used for import and API key administration.

The example key can create images and read templates. It cannot delete images or perform authenticated custom-storage rendering; the latter needs `images:store`.

## Deliver credentials to your application

Reference `htmlcsstoimage_api_key.rendering.api_id` and `.api_key` in your secret manager resource. The [AWS example](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image/tree/main/examples/aws-e2e) demonstrates storing API credentials in Secrets Manager.

Sensitive outputs are hidden from ordinary CLI output, but the credentials remain in Terraform state. Use a protected state backend. Import cannot recover the API key secret; it restores readable metadata and the API ID.

Keep the credentials used by the Terraform provider active throughout updates and destroy. See the [API key reference](../resources/api_key.md) for permission changes, disabling, and import behavior.
