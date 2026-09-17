#!/usr/bin/env bash
# Install the development HCTI provider alongside registry AWS/time providers.
set -euo pipefail
cd "$(dirname "$0")"
example_dir="$PWD"
make -C ../.. build
version="$(cat ../../VERSION)"
platform="$(terraform version -json | jq -er '.platform')"
mirror="$example_dir/.local-providers"
package_dir="$mirror/registry.terraform.io/htmlcsstoimage/html-css-to-image/$version/$platform"
mkdir -p "$package_dir"
cp ../../bin/terraform-provider-html-css-to-image "$package_dir/terraform-provider-html-css-to-image_v${version}"
cat > .local.tfrc <<EOF
provider_installation {
  filesystem_mirror {
    path    = "$mirror"
    include = ["htmlcsstoimage/html-css-to-image"]
  }
  direct {
    exclude = ["htmlcsstoimage/html-css-to-image"]
  }
}
EOF
export TF_CLI_CONFIG_FILE="$example_dir/.local.tfrc"
terraform init -backend=false
terraform validate
printf '\nFor subsequent commands in this shell, run:\nexport TF_CLI_CONFIG_FILE=%q\n' "$TF_CLI_CONFIG_FILE"
