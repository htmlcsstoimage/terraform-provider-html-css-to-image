package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (*ProxyResource) Metadata(_ context.Context, _ resource.MetadataRequest, r *resource.MetadataResponse) {
	r.TypeName = "htmlcsstoimage_proxy"
}
func (*ProxyResource) Schema(_ context.Context, _ resource.SchemaRequest, r *resource.SchemaResponse) {
	r.Schema = schema.Schema{
		MarkdownDescription: "Manage a rendering proxy. Updates preserve its ID. Import by proxy ID; imported passwords are unavailable and may be retained on unrelated updates. Passwords are sensitive but stored in provider state.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Proxy ID."},
			"name":           schema.StringAttribute{Required: true, MarkdownDescription: "Display name, 3–500 characters after trimming."},
			"url":            schema.StringAttribute{Required: true, MarkdownDescription: "HTTP(S) proxy origin without credentials, port, query, fragment, or a non-root path."},
			"port":           schema.Int64Attribute{Optional: true, MarkdownDescription: "Port from 1 to 65535. Omitted sends null; the API chooses the scheme default."},
			"effective_port": schema.Int64Attribute{Computed: true, MarkdownDescription: "Port returned by the API."},
			"disabled":       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Disable the proxy. Defaults to false."},
			"bypass_hosts":   schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Up to 100 hostnames or URLs to bypass. Host comparison ignores case, URL paths, duplicates, and surrounding whitespace. Omission clears the list."},
			"authentication": schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "Proxy credentials. Omit the whole object to remove authentication.", Attributes: map[string]schema.Attribute{
				"username": schema.StringAttribute{Required: true, MarkdownDescription: "Username, up to 512 characters. Empty is valid; whitespace is preserved."},
				"password": schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Password, up to 484 UTF-8 bytes. Empty is valid. Required for creation or a username change; omitted on update retains the password only when the existing username matches."},
			}},
			"created_at": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Creation timestamp."},
			"updated_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
		}}
}
