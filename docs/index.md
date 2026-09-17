---
page_title: "HTML/CSS to Image Provider"
description: "Manage HTML/CSS to Image resources."
---

# HTML/CSS to Image Provider

Configure `api_id` and sensitive `api_key`, or use `HCTI_API_ID` and `HCTI_API_KEY`. Optional `base_url` defaults to `https://hcti.io`.

```hcl
terraform {
  required_providers {
    htmlcsstoimage = {
      source  = "htmlcsstoimage/html-css-to-image"
      version = "= 0.1.0-beta.1"
    }
  }
}

provider "htmlcsstoimage" {}
```

This provider is in beta. Select the prerelease version explicitly as shown above; configuration and behavior may change before the stable release.

Resources include proxies, API keys, OG configs, storage destinations, templates, and HTML/CSS, URL, and templated images. Images save rendering definitions; creation does not render image bytes.

## Authentication

Set `HCTI_API_ID` and `HCTI_API_KEY` in the environment running Terraform. The key must have the permissions required by the resources being managed. Do not commit credentials to source control.

## Arguments

| Argument | Description |
| --- | --- |
| `api_id` | Optional API ID; otherwise uses `HCTI_API_ID`. |
| `api_key` | Optional sensitive API key; otherwise uses `HCTI_API_KEY`. |
| `base_url` | Optional API base URL; defaults to `https://hcti.io`. |

The [AWS external-ID data source](data-sources/aws_storage_external_id.md) helps configure role trust before creating an AWS storage destination.
