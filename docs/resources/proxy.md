---
page_title: "htmlcsstoimage_proxy Resource"
description: "Configure a proxy for HTML/CSS to Image rendering."
---

# htmlcsstoimage_proxy

Register an existing HTTP or HTTPS proxy with HTML/CSS to Image. Use the resource's `id` as `proxy_id` when creating images or configuring templates.

This resource manages the proxy's HCTI configuration. It does not provision or operate the proxy server. Updating its settings keeps the same ID; destroying the resource removes the configuration from HCTI.

For provider installation and API credentials, see [provider setup](https://github.com/htmlcsstoimage/terraform-provider-html-css-to-image#installation). The API key needs `proxies:read`, `proxies:create_update`, and `proxies:delete` to manage the full lifecycle.

## Basic example

```hcl
resource "htmlcsstoimage_proxy" "rendering" {
  name = "Rendering proxy"
  url  = "https://proxy.example.com"
}

output "proxy_id" {
  value = htmlcsstoimage_proxy.rendering.id
}
```

Omitting `port` lets the API choose the URL scheme's default: 80 for HTTP or 443 for HTTPS. The input remains unset; `effective_port` reports the port returned by the API. Omitting `authentication` configures a proxy without credentials.

## Proxy with authentication

```hcl
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

  bypass_hosts = ["static.example.com", "192.0.2.10"]
}
```

Supply the password through `TF_VAR_proxy_password` or your normal Terraform variable mechanism. Put credentials and port in their separate fields rather than including them in `url`.

Terraform marks the password sensitive, which hides it in ordinary plan output. Supplied passwords are still stored in Terraform state.

## Arguments

| Argument | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | String | Yes | Display name, 3–500 characters after leading and trailing whitespace is removed. |
| `url` | String | Yes | Absolute HTTP or HTTPS proxy URL, up to 512 characters. No embedded credentials, port, query, fragment, or path other than `/`. |
| `port` | Number | No | Integer from 1 to 65535. Omitted or null lets the API select the scheme's default port. |
| `disabled` | Boolean | No | Disable the proxy without deleting its configuration. Defaults to `false`. |
| `bypass_hosts` | Set of strings | No | Up to 100 hostnames, IP addresses, or absolute URLs that should connect directly. Omitted, null, or empty clears the list. |
| `authentication` | Object | No | Credentials described below. Omitted or null removes authentication. |

### Authentication

| Argument | Type | Required | Description |
| --- | --- | --- | --- |
| `username` | String | Yes, within `authentication` | Up to 512 characters. An empty string is valid. Whitespace is preserved. |
| `password` | Sensitive string | On creation or username changes | Up to 484 UTF-8 bytes. An empty string sets an empty password. Whitespace is preserved. Omitted or null on an existing proxy retains its password when the username is unchanged. |

Password retention is handled by the provider. There is no `retain_password` argument in the Terraform resource.

## Outputs

The resource also exposes its configured arguments and these read-only attributes:

| Attribute | Type | Description |
| --- | --- | --- |
| `id` | String | HCTI proxy ID. Use this for `proxy_id` in image or template settings. |
| `effective_port` | Number | Port returned by the API; may be null if the response does not specify one. |
| `created_at` | String | Creation timestamp in RFC 3339 format. |
| `updated_at` | String | Last update timestamp in RFC 3339 format. |

## Updating credentials

For an existing authenticated proxy, this configuration retains the stored password while changing its name:

```hcl
resource "htmlcsstoimage_proxy" "rendering" {
  name = "Renamed rendering proxy"
  url  = "https://proxy.example.com"
  port = 8080

  authentication = {
    username = "rendering"
  }

  bypass_hosts = ["static.example.com", "192.0.2.10"]
}
```

Keep the rest of your desired configuration in place: updates replace settings, so omitted ordinary options are cleared or reset.

| Intended change | Configuration |
| --- | --- |
| Replace the password | Supply the new `authentication.password`. |
| Set an empty password | Set `authentication.password = ""`. |
| Keep the stored password | Keep the same username and omit `password`, or set it to null. |
| Change the username | Supply the new username and a password together. |
| Remove authentication | Remove the entire `authentication` object, or set it to null. |

A retained password must already exist. The same rules apply when the proxy is disabled. If the username changes outside Terraform, refresh exposes the new username and a subsequent change back requires a password.

## Bypass hosts

Bypass entries are matched by hostname. For example, `STATIC.EXAMPLE.COM` and `https://static.example.com/assets` both normalize to `static.example.com`. URL paths are not bypass rules. Hosts are lowercased and deduplicated.

The provider preserves equivalent configured values when refreshing, so server normalization alone does not cause another update. Omitting `bypass_hosts` or setting it to `[]` clears the complete list.

## Import

Import an existing HCTI proxy by its management ID:

```sh
terraform import htmlcsstoimage_proxy.rendering PROXY_ID
```

Declare the resource in configuration before importing. Then inspect the imported settings:

```sh
terraform state show htmlcsstoimage_proxy.rendering
terraform plan
```

Match the configuration to the imported name, URL, port, disabled state, bypass hosts, and username. The API never returns the password, so import leaves it unset. Keep `authentication` with the existing username and omit `password` to retain it during later updates. Removing the authentication object would remove credentials on the next apply.

Import records the port returned by the API. Keep that value in configuration for an unchanged plan, or intentionally omit it to reset to the scheme default on an update.

## Refresh and deletion

Refresh reads metadata. It does not render an image or test connectivity to the proxy. An unchanged plan performs no writes.

If HCTI reports the proxy missing with HTTP 404, refresh removes it from state. Authentication, authorization, rate-limit, and other errors are reported without treating the proxy as deleted.

Destroy deletes the HCTI proxy configuration. The external proxy server remains running. Deleting a configuration that is already missing succeeds. Review image and template references before removing a proxy they use.
