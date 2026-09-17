# Live acceptance testing

The automated test suite uses a simulated API. These steps exercise the deployed
management API with real resources. Use a disposable test organization with the
relevant read/write/delete permissions. State files contain credentials and other
sensitive values; keep the test directory private and outside version control.

## Build and configure Terraform

From the provider repository:

```sh
GOWORK=off make build
```

Create a separate Terraform CLI configuration file, for example
`/private/tmp/hcti-live.tfrc`, containing:

```hcl
provider_installation {
  dev_overrides {
    "htmlcsstoimage/html-css-to-image" = "/absolute/path/to/terraform-provider-html-css-to-image/bin"
  }
  direct {}
}
```

Adjust the binary directory for your checkout. Set `TF_CLI_CONFIG_FILE` to the
absolute path of that file, and set `HCTI_API_ID` and `HCTI_API_KEY` securely in your
shell. Credentials must belong to the test organization.

Copy `examples/images/main.tf` to a new private test directory and run commands
there. It covers HTML/CSS images, URL images, templates, and templated images.
For a nonproduction API, set `base_url` in its provider block to the deployed API
origin. The default is `https://hcti.io`.

Skip `terraform init` for this development-override setup; Terraform uses the
built binary directly.

## First lifecycle check

```sh
terraform plan
terraform apply
terraform plan -detailed-exitcode
terraform apply -refresh-only
terraform plan -detailed-exitcode
```

Both plans after creation should exit 0 with no changes. Exit 2 means Terraform
has planned changes; exit 1 means an error. Check API logs: refresh should fetch
metadata, with no requests to render images or test storage/proxy connectivity.

Edit the HTML image's content and the template's content, then plan and apply.
The HTML image should be replaced. The template should keep its ID and receive a
new version. The example's pinned templated image should be replaced when that
version changes. Run another plan and expect no changes.

## Import and saved storage mode

Create a second private directory with the same provider configuration and an
`htmlcsstoimage_image_templated` resource matching the test image's template ID and
values. Omit `format` and `template_version` initially, since import cannot recover
format or infer pinning intent. Import the existing test image:

```sh
terraform import htmlcsstoimage_image_templated.card IMAGE_ID
terraform state show htmlcsstoimage_image_templated.card
terraform plan
```

Check that `resolved_template_version` is populated and `template_version` stays
null. Repeat for an image created from a template using a destination with HCTI
storage disabled: the saved flag should be true, `render_method` should be `PUT`,
`render_requires_auth` should be true, and `public_url` should be null. Import needs
image metadata access, with no template or storage-destination reads.

For ordinary images, expect a public GET operation with authentication not required.
Changing a destination's current mode must not change an existing image's saved
mode. Test this only with the disposable destination.

The import directory temporarily tracks an image owned by the original test
directory. After inspection, remove it from the import directory's state with
`terraform state rm htmlcsstoimage_image_templated.card`; this does not delete the
remote image. Do not apply or destroy from both directories.

## Complete the acceptance pass

- Create two identical images as separate resources; their IDs must differ.
- Exercise each remaining resource example: API key, proxy, OG config, and storage
  destination. Check create, unrelated updates, refresh, import, and unchanged plans.
- Check proxy password and storage secret retention, replacement, and removal.
- Verify both OG variants and each storage provider you intend to support at launch
  using valid test credentials and destinations.
- Test unpinned templated images: a new template version alone must not replace the
  existing image. Explicit replacement should use the latest version.
- Review `terraform plan -destroy`, then run `terraform destroy` from the original
  test directory. Image deletion may be asynchronous; confirm eventual absence
  through metadata. API-key deletion disables the key rather than erasing metadata.

After Terraform acceptance, repeat representative preview/update/refresh/import/
destroy checks using the Pulumi YAML examples and the built local Pulumi plugin.
Include secret outputs and a second preview with no changes. Terraform owns the
resource lifecycle implementation; Pulumi needs this additional bridge check.
