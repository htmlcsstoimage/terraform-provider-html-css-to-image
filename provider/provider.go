// Package provider exposes the Terraform provider for embedding in the Pulumi bridge.
package provider

import (
	framework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/htmlcsstoimage/terraform-provider-html-css-to-image/internal/provider"
)

// New constructs the Terraform provider. userAgent identifies the embedding host
// alongside the Go SDK identifier; an empty value uses HCTITerraform/version.
func New(version, userAgent string) framework.Provider {
	if userAgent == "" {
		return provider.New(version)()
	}
	return provider.NewWithUserAgent(version, userAgent)()
}
