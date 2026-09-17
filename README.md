# HTML/CSS to Image Terraform provider

Manage HTML/CSS to Image resources with Terraform.

[Website](https://htmlcsstoimage.com) · [Documentation](https://docs.htmlcsstoimage.com/) · [Management API](https://docs.htmlcsstoimage.com/management-api/) · [API reference](https://htmlcsstoimage.com/api-docs/) · [Terraform Registry](https://registry.terraform.io/providers/htmlcsstoimage/html-css-to-image/latest/docs)

Resources include proxies, API keys, OG configs, storage destinations, HTML/CSS templates, and HTML/CSS, URL, and templated images. Templates create new versions on edits; image inputs trigger replacement. Image creation saves a definition without rendering bytes.

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0"
    }
  }
}

provider "htmlcsstoimage" {}

resource "htmlcsstoimage_image_html_css" "hello" {
  html = "<h1>Hello from Terraform</h1>"
}
```

Set `HCTI_API_ID` and `HCTI_API_KEY` in your environment, then run:

```sh
terraform init
terraform plan
terraform apply
```

The API key needs the permissions required by the resources you manage. The default API URL is `https://hcti.io`. See [provider configuration](docs/index.md) and the [Terraform Registry](https://registry.terraform.io/providers/htmlcsstoimage/html-css-to-image/latest/docs).

## Resource documentation

For a complete AWS setup, see [the end-to-end example](examples/aws-e2e/README.md): S3, IAM trust with the HCTI external ID, two API keys in Secrets Manager, a storage destination, and image creation/rendering.

- [Proxy: examples, inputs, outputs, credential updates, and import](docs/resources/proxy.md)
- [API key: permissions, one-time secrets, import, and disable behavior](docs/resources/api_key.md)
- [OG config: HTML/CSS and templates, render options, headers, and import](docs/resources/og_config.md)
- [Storage destination: providers, credentials, connection tests, and import](docs/resources/storage_destination.md)
- [Template: versioning, rendering inputs, and import](docs/resources/template.md)
- [HTML/CSS image](docs/resources/image_html_css.md), [URL image](docs/resources/image_url.md), and [templated image](docs/resources/image_templated.md)
- [AWS storage external ID lookup](docs/data-sources/aws_storage_external_id.md)
- [Existing template lookup](docs/data-sources/template.md) and [template version listing](docs/data-sources/template_versions.md)

## Proxy behavior

See the [resource reference](docs/resources/proxy.md). All settings update in place. Reads only fetch metadata and never render images or test proxy connectivity.

Passwords are sensitive but persist in Terraform state when supplied. Import reconstructs the username without inventing a password. Omit `password` to retain an existing password during an unrelated update; changing username requires a password, and an empty password is valid. Omit the entire `authentication` object to remove credentials. The provider sends API retention flags internally.

An omitted `port` is sent as null. `effective_port` exposes the server-selected value without filling the input with a default. Equivalent normalized names and bypass hosts preserve the configured representation to avoid inconsistent plans.

Only HTTP 404 removes a resource during refresh or makes a repeated delete successful. Other errors surface without interpreting the resource as missing. Writes are not retried automatically.

API failures include the HTTP status, API error code/message, and field-level validation paths and messages. Known sensitive request/state values are redacted from messages; raw response bodies and headers are not included.

Local validation checks input structure and basic invariants, such as nonnegative rendering values. The API controls rendering ranges, account limits, string lengths, collection sizes, supported timezones, and storage regions. Documented API limits are informational, not hardcoded provider restrictions.

## Contributing

See [development](docs/development.md), [live acceptance testing](docs/live-testing.md), and [releasing](docs/releasing.md).
