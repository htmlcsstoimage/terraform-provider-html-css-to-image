# AWS + HCTI end-to-end example

This Terraform stack creates:

- A private, encrypted S3 bucket with public access blocked and bucket-owner-enforced ownership.
- An HCTI API key granting only `images:create`, stored as JSON in its own AWS Secrets Manager secret.
- An HCTI API key granting all current and future permissions, stored in a separate secret.
- The organization's AWS external-ID lookup, an IAM role trusting HCTI's writer role with that external ID, and a bucket-prefix-scoped permissions policy.
- An HCTI S3 storage destination with HCTI storage disabled.
- An HTML/CSS image using the destination. An explicit render script writes the image to S3 afterward.

Both secrets contain `api_id` and `api_key`. Only secret ARNs are exported from the stack. Separate secrets let applications receive access to the create-only credentials without access to the automation credentials. This example does not create an application runtime or grant a workload permission to read either secret.

## Prerequisites

- Terraform 1.5+, AWS CLI, `jq`, and `curl`.
- AWS credentials via your normal AWS profile/SSO/environment configuration, authorized to manage S3, IAM roles/policies, and Secrets Manager. The render script also needs Secrets Manager `GetSecretValue` and S3 `ListBucket` access.
- An HCTI organization with storage destinations enabled.
- Existing bootstrap credentials in `HCTI_API_ID` and `HCTI_API_KEY`, authorized for the resource lifecycle and for granting all current/future permissions. Keep these bootstrap credentials active through destroy. The provider does not switch to a key created by this stack.
- The HCTI storage writer role ARN from the dashboard-generated S3 trust policy's `Principal.AWS`. This is the HCTI role, not the customer role created here. The AWS external-ID lookup also returns this principal as `writer_role_arn`.

AWS resources incur normal charges. Terraform state contains both API key secrets even though outputs are redacted and the keys are also saved in Secrets Manager. Use a private test directory or an encrypted, access-controlled state backend.

## Configure and apply

From this directory:

```sh
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars, especially hcti_writer_role_arn.
# Set AWS credentials and HCTI_API_ID / HCTI_API_KEY in your shell.
terraform init
terraform plan
terraform apply
terraform plan -detailed-exitcode
```

Terraform installs the HCTI and AWS/time providers from the Terraform Registry.

For staging, set `hcti_base_url` to the deployed API origin and use that environment's writer role ARN and bootstrap credentials. AWS region/profile selection must remain consistent between Terraform and the render script.

The role depends on the organization external ID. The storage destination waits for its role permissions and bucket settings, including a configurable 30-second IAM propagation delay, before HCTI tests the connection. The delay is a practical allowance, not a guarantee. If AWS propagation causes an error, inspect the destination and Terraform state before retrying; writes are not blindly replayed by the provider.

The role grants `PutObject` and `GetObject` only under `renders/`; `DeleteObject` is limited to connection-test cleanup objects. It has no Secrets Manager permissions. SSE-S3 encryption avoids needing a separate KMS key policy for the cross-account writer.

## Render and verify

`terraform apply` saves an image definition. It does not render bytes. Expect an authenticated `PUT` operation URL, `render_requires_auth = true`, and no public HCTI URL.

```sh
terraform output
bash render.sh
terraform apply -refresh-only
terraform plan -detailed-exitcode
```

The render script retrieves the automation key from Secrets Manager, invokes the managed image's `/store` URL, and lists objects under the S3 prefix. It does not print the credentials. The create-only key cannot perform this operation: authenticated storage rendering requires `images:store`, which is intentionally absent from that key.

Verify that a rendered image object appears, the metadata records its custom-storage save, and the subsequent plan exits 0 (no changes). Rendering can change computed timestamps, so refresh before comparing plans. The bucket remains private; use your AWS credentials to download an object for visual inspection.

For broader acceptance, change the HTML and apply: the image should receive a new ID while the bucket, role, destination, and keys remain stable. Change a key description and confirm its stored secret remains unchanged. Import the image in a separate test state to verify its saved custom-storage mode, then remove that temporary import from state without destroying the shared image.

## Cleanup

HCTI image deletion does not delete rendered S3 objects. By default, Terraform refuses to delete a nonempty bucket. Empty the test bucket yourself, or explicitly set `force_destroy_bucket = true` and apply that setting before reviewing and running destroy:

```sh
terraform plan -destroy
terraform destroy
```

Destroy disables the two HCTI API keys and schedules both Secrets Manager secrets for deletion after seven days. Secrets scheduled for deletion cannot be used normally. It deletes the image definition, storage destination, customer IAM role/policy, and empty (or explicitly force-destroyable) bucket. It does not affect the HCTI writer role or your bootstrap key.

References: [HCTI S3 setup](https://docs.htmlcsstoimage.com/guides/advanced/storage-destinations/s3/), [AWS secret versions](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret_version), [S3 bucket lifecycle](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket).
