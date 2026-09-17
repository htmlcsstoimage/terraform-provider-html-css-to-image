#!/usr/bin/env bash
# Run explicitly after terraform apply. This renders the managed image into S3.
set -euo pipefail
set +x
export AWS_PAGER=""
outputs="$(terraform output -json)"
region="$(jq -er '.aws_region.value' <<< "$outputs")"
secret_arn="$(jq -er '.automation_secret_arn.value' <<< "$outputs")"
image_url="$(jq -er '.image_url.value' <<< "$outputs")"
method="$(jq -er '.render_method.value' <<< "$outputs")"
if [[ "$method" != PUT ]] || ! jq -e '.render_requires_auth.value == true' <<< "$outputs" >/dev/null; then
  echo "Expected an authenticated PUT rendering operation." >&2
  exit 1
fi
# The all-permissions key has images:store. The create-only key deliberately does not.
credentials="$(aws secretsmanager get-secret-value --region "$region" --secret-id "$secret_arn" --query SecretString --output text)"
authorization="$(jq -er '(.api_id + ":" + .api_key) | @base64' <<< "$credentials")"
# Pass auth through stdin rather than exposing the key in command-line arguments.
printf 'header = "Authorization: Basic %s"\n' "$authorization" |
  curl --config - --fail --silent --show-error --request PUT --output /dev/null "$image_url"
unset credentials authorization
bucket="$(jq -er '.bucket_name.value' <<< "$outputs")"
prefix="$(jq -er '.key_prefix.value' <<< "$outputs")"
echo "Render request succeeded. Objects under s3://$bucket/$prefix/:"
aws s3api list-objects-v2 --region "$region" --bucket "$bucket" --prefix "$prefix/" --query 'Contents[].{Key:Key,Size:Size}' --output table
