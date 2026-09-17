package provider

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var storageRoleARN = regexp.MustCompile(`^arn:aws:iam::[0-9]{12}:role/.+$`)
var storageAccountID = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

func (m storageModel) validate(old *storageModel) diag.Diagnostics {
	var d diag.Diagnostics
	selected, _ := storageSelection(m.Connection)
	invalid := func(k, rule string) {
		if selected != "" && strings.HasPrefix(k, "connection_info.") {
			k = "connection_info." + selected + "." + strings.TrimPrefix(k, "connection_info.")
		}
		d.AddError("Invalid storage destination", k+" "+rule)
	}
	if !m.known() {
		d.AddError("Unknown storage settings", "Storage settings must be known during apply.")
		return d
	}
	if strings.TrimSpace(m.Name.ValueString()) == "" {
		invalid("name", "must not be empty after trimming.")
	}
	if m.Connection.IsNull() {
		invalid("connection_info", "is required.")
		return d
	}
	if _, count := storageSelection(m.Connection); count != 1 {
		invalid("connection_info", "must contain exactly one provider object.")
		return d
	}
	a := storageFlat(m.Connection).Attributes()
	s := func(k string) string { return storageNormalized(k, a[k].(types.String).ValueString()) }
	kind := a["provider"].(types.String).ValueString()
	required := []string{"bucket"}
	allowed := map[string]bool{"provider": true, "bucket": true, "key_prefix": true}
	switch kind {
	case "aws_s3":
		required = append(required, "region", "role_arn")
	case "cloudflare_r2":
		required = append(required, "cloudflare_account_id")
		allowed["cloudflare_jurisdiction"] = true
	case "backblaze_b2", "digitalocean_spaces", "wasabi":
		required = append(required, "region")
	case "google_cloud_storage":
	case "other_s3_compatible":
		required = append(required, "endpoint")
		allowed["region"] = true
		allowed["force_path_style"] = true
	default:
		invalid("connection_info.provider", "must be a supported storage provider.")
		return d
	}
	if kind != "aws_s3" {
		required = append(required, "access_key_id")
		allowed["secret_access_key"] = true
	}
	for _, k := range required {
		allowed[k] = true
		if s(k) == "" {
			invalid("connection_info."+k, "is required for the selected provider.")
		}
	}
	for k, v := range a {
		if !allowed[k] && !v.IsNull() {
			invalid("connection_info."+k, "does not apply to the selected provider.")
		}
	}
	if kind == "aws_s3" {
		if !storageRoleARN.MatchString(s("role_arn")) {
			invalid("connection_info.role_arn", "must be an AWS IAM role ARN.")
		}
		if strings.ContainsAny(s("bucket")+s("key_prefix"), "*?") {
			invalid("connection_info", "AWS buckets and prefixes cannot contain IAM wildcards (* or ?).")
		}
	} else {
		secret := a["secret_access_key"].(types.String)
		if secret.IsNull() {
			if old == nil || !storageSameIdentity(m.Connection, old.Connection) {
				d.AddError("Secret access key required", "Supply secret_access_key when creating a destination or changing its provider or access key ID.")
			}
		}
		if !secret.IsNull() && (s("secret_access_key") == "" || !utf8.ValidString(secret.ValueString())) {
			invalid("connection_info.secret_access_key", "must be nonempty UTF-8 after trimming.")
		}
	}
	if kind == "cloudflare_r2" {
		if !storageAccountID.MatchString(s("cloudflare_account_id")) {
			invalid("connection_info.cloudflare_account_id", "must contain 32 hexadecimal characters.")
		}
		if v := s("cloudflare_jurisdiction"); v != "" && v != "eu" && v != "fedramp" {
			invalid("connection_info.cloudflare_jurisdiction", "must be eu, fedramp, or omitted.")
		}
	}
	if kind == "google_cloud_storage" && !strings.HasPrefix(s("access_key_id"), "GOOG") {
		invalid("connection_info.access_key_id", "must be a Google Cloud Storage HMAC access ID beginning with GOOG.")
	}
	if kind == "other_s3_compatible" && !validOrigin(s("endpoint"), true) {
		invalid("connection_info.endpoint", "must be a public HTTPS origin without credentials, path, query, or fragment. The API verifies network eligibility.")
	}
	return d
}
