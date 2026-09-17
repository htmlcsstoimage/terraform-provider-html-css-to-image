locals {
  key_prefix = "renders"
}

# Least privilege for an application that only creates image definitions.
resource "htmlcsstoimage_api_key" "create" {
  name        = "${var.name} image creation"
  permissions = ["images:create"]
}

# Create an automation key with all existing permissions and any added in the future.
resource "htmlcsstoimage_api_key" "automation" {
  name                   = "${var.name} automation"
  all_future_permissions = true
}

# Create a separate secret container for application credentials, with seven-day deletion recovery.
resource "aws_secretsmanager_secret" "create" {
  name_prefix             = "${var.name}-create-"
  description             = "HCTI credentials restricted to images:create"
  recovery_window_in_days = 7
}

# Save the create-only key's authentication ID and secret as a JSON credential pair.
resource "aws_secretsmanager_secret_version" "create" {
  secret_id = aws_secretsmanager_secret.create.id
  secret_string = jsonencode({
    api_id  = htmlcsstoimage_api_key.create.api_id
    api_key = htmlcsstoimage_api_key.create.api_key
  })
}

# Keep automation credentials separate so access can be granted independently of the application key.
resource "aws_secretsmanager_secret" "automation" {
  name_prefix             = "${var.name}-automation-"
  description             = "HCTI credentials with all current and future permissions"
  recovery_window_in_days = 7
}

# Save the automation credential pair; render.sh reads this secret to authenticate storage rendering.
resource "aws_secretsmanager_secret_version" "automation" {
  secret_id = aws_secretsmanager_secret.automation.id
  secret_string = jsonencode({
    api_id  = htmlcsstoimage_api_key.automation.api_id
    api_key = htmlcsstoimage_api_key.automation.api_key
  })
}

# Create a uniquely named bucket for rendered images; deleting a nonempty bucket requires explicit opt-in.
resource "aws_s3_bucket" "images" {
  bucket_prefix = "${var.name}-"
  force_destroy = var.force_destroy_bucket
}

# Block public access through both bucket policies and object ACLs.
resource "aws_s3_bucket_public_access_block" "images" {
  bucket                  = aws_s3_bucket.images.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Give the bucket owner ownership of uploaded objects and disable object ACLs.
resource "aws_s3_bucket_ownership_controls" "images" {
  bucket = aws_s3_bucket.images.id
  rule { object_ownership = "BucketOwnerEnforced" }
}

# Encrypt new objects with S3-managed keys, without a separate cross-account KMS policy.
resource "aws_s3_bucket_server_side_encryption_configuration" "images" {
  bucket = aws_s3_bucket.images.id
  rule {
    apply_server_side_encryption_by_default { sse_algorithm = "AES256" }
  }
}

# Retrieve the organization-specific external ID required by the role's trust policy.
data "htmlcsstoimage_aws_storage_external_id" "organization" {}

# Allow HCTI's writer role to assume the customer role only with this organization's external ID.
data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "AWS"
      identifiers = [var.hcti_writer_role_arn]
    }
    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = [data.htmlcsstoimage_aws_storage_external_id.organization.external_id]
    }
  }
}

# Create the customer-side role HCTI assumes when testing the destination and accessing image objects.
resource "aws_iam_role" "hcti_storage" {
  name_prefix        = "${var.name}-storage-"
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

# Scope reads/writes to the image prefix and deletion to connection-test cleanup objects.
data "aws_iam_policy_document" "objects" {
  statement {
    sid       = "WriteAndReadRenderedImages"
    actions   = ["s3:PutObject", "s3:GetObject"]
    resources = ["${aws_s3_bucket.images.arn}/${local.key_prefix}/*"]
  }
  statement {
    sid       = "CleanUpConnectionTests"
    actions   = ["s3:DeleteObject"]
    resources = ["${aws_s3_bucket.images.arn}/${local.key_prefix}/.hcti/connection-tests/*"]
  }
}

# Attach the scoped S3 permissions to the customer role before configuring HCTI storage.
resource "aws_iam_role_policy" "objects" {
  name   = "hcti-rendered-images"
  role   = aws_iam_role.hcti_storage.id
  policy = data.aws_iam_policy_document.objects.json
}

# Role existence alone does not imply its policy is attached or propagated.
resource "time_sleep" "iam_ready" {
  create_duration = var.iam_propagation_wait
  triggers = {
    role_arn = aws_iam_role.hcti_storage.arn
    trust    = data.aws_iam_policy_document.trust.json
    policy   = aws_iam_role_policy.objects.policy
  }
}

# Register the bucket and role with HCTI after AWS setup; output is stored only in this S3 destination.
resource "htmlcsstoimage_storage_destination" "images" {
  name                  = "${var.name} S3 images"
  hcti_storage_disabled = true
  connection_info = {
    aws_s3 = {
      bucket     = aws_s3_bucket.images.id
      region     = var.aws_region
      role_arn   = aws_iam_role.hcti_storage.arn
      key_prefix = local.key_prefix
    }
  }
  depends_on = [
    time_sleep.iam_ready,
    aws_s3_bucket_public_access_block.images,
    aws_s3_bucket_ownership_controls.images,
    aws_s3_bucket_server_side_encryption_configuration.images,
  ]
}

# Save an image definition using the destination; run render.sh afterward to render its bytes into S3.
resource "htmlcsstoimage_image_html_css" "card" {
  html                   = "<main><h1>Hello from AWS + HCTI</h1><p>Terraform → HCTI → S3</p></main>"
  css                    = "main { padding: 48px; background: #eff6ff; color: #172554; font-family: sans-serif; }"
  storage_destination_id = htmlcsstoimage_storage_destination.images.id
  metadata               = { example = "aws-e2e" }
}
