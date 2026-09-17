# Development

Requires Go 1.25.11 or newer and Terraform (tests use 1.14.3). The management SDK is pinned to the published `go-client` v0.3.1 module; no sibling SDK checkout is required.

```sh
make build
make test
make vet
```

Tests run against a local in-memory API with dummy credentials. Lifecycle tests invoke Terraform and check import, secret transitions, normalized inputs, unchanged plans, and deletion. They do not contact HCTI.

To use the built provider locally, add this to a separate Terraform CLI configuration file, replacing the path:

```hcl
provider_installation {
  dev_overrides {
    "htmlcsstoimage/html-css-to-image" = "/absolute/path/to/terraform-provider-html-css-to-image/bin"
  }
  direct {}
}
```

Set `TF_CLI_CONFIG_FILE` to that file and run `terraform plan`/`apply` in [examples/proxy](../examples/proxy). With this development override, skip `terraform init` for single-provider examples; Terraform uses your built binary directly. Set `HCTI_API_ID`, `HCTI_API_KEY`, and `TF_VAR_proxy_password` before applying. Live use requires the management API to be deployed and the caller to have proxy read/write/delete permissions.


See [live acceptance testing](live-testing.md) and [releasing](guides/releasing.md).
