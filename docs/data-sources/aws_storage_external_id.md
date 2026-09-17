---
page_title: "htmlcsstoimage_aws_storage_external_id Data Source"
description: "Read the organization's external ID for AWS storage role trust policies."
---

# htmlcsstoimage_aws_storage_external_id

Read the organization's external ID and HCTI writer role ARN before creating an AWS storage destination. No destination ID or input attributes are needed. This does not create a destination or test a bucket connection.

The caller requires `storage_destinations:create_update` for this lookup.

```hcl
data "htmlcsstoimage_aws_storage_external_id" "organization" {}

output "hcti_aws_external_id" {
  value = data.htmlcsstoimage_aws_storage_external_id.organization.external_id
}
```

Use `external_id` in the IAM role trust policy's `sts:ExternalId` condition, and `writer_role_arn` as its `Principal.AWS`, allowing `sts:AssumeRole`. An external-ID condition alone is not a complete trust policy. Reference the resulting role ARN from the [storage destination](../resources/storage_destination.md) so Terraform creates the role before HCTI tests the connection.

| Output | Meaning |
| --- | --- |
| `external_id` | Organization-specific external ID for the AWS role trust policy. |
| `writer_role_arn` | HCTI writer role ARN to use as the trust policy's AWS principal. |

Lookup failures remain errors, including authorization failures and HTTP 404.

The returned `writer_role_arn` identifies HCTI's caller role. The storage destination's `role_arn` identifies the role you create in your own AWS account.
