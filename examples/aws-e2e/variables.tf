variable "name" {
  type        = string
  default     = "hcti-e2e"
  description = "Short prefix for AWS and HCTI resources."
  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{2,24}$", var.name))
    error_message = "Use 3–25 lowercase letters, digits, or hyphens, starting with a letter."
  }
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "hcti_base_url" {
  type        = string
  default     = "https://hcti.io"
  description = "Deployed HCTI API origin."
}

variable "hcti_writer_role_arn" {
  type        = string
  description = "HCTI storage writer role ARN from the dashboard-generated S3 trust policy. Not the customer role created here."
  validation {
    condition     = can(regex("^arn:aws:iam::[0-9]{12}:role/.+$", var.hcti_writer_role_arn))
    error_message = "Supply the HCTI AWS writer role ARN, not an account root or wildcard principal."
  }
}

variable "force_destroy_bucket" {
  type        = bool
  default     = false
  description = "Explicitly allow destroying the example bucket and all its objects. Apply this change before destroy."
}

variable "iam_propagation_wait" {
  type        = string
  default     = "30s"
  description = "Wait after role/policy changes before HCTI tests the connection. AWS propagation can take longer."
}
