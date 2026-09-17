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

func (*APIKeyResource) Metadata(_ context.Context, _ resource.MetadataRequest, r *resource.MetadataResponse) {
	r.TypeName = "htmlcsstoimage_api_key"
}
func (*APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, r *resource.SchemaResponse) {
	stable := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	r.Schema = schema.Schema{MarkdownDescription: "Manage an API key. Updates preserve its ID and secret; destroy disables the key. Import cannot recover the secret. Supplied and generated secrets persist in provider state.", Attributes: map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "Management ID used for import and CRUD; distinct from api_id."},
		"api_id":                 schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "API ID used to authenticate with this key."},
		"api_key":                schema.StringAttribute{Computed: true, Sensitive: true, PlanModifiers: stable, MarkdownDescription: "Secret returned only at creation. Preserved during refresh and updates; unavailable after import."},
		"name":                   schema.StringAttribute{Optional: true, MarkdownDescription: "Display name, up to 255 characters. Omitted, null, or blank uses an API-generated name based on the original creation time."},
		"effective_name":         schema.StringAttribute{Computed: true, MarkdownDescription: "Actual display name returned by the API, including a generated name when name is omitted."},
		"description":            schema.StringAttribute{Optional: true, MarkdownDescription: "Purpose of the key, up to 2000 characters. Omitted, null, or blank clears it."},
		"disabled":               schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Disable authentication with the key. Defaults to false."},
		"all_future_permissions": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Grant all current and future permissions. Requires caller authority. Defaults to false."},
		"permissions":            schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Complete permission set. Required when all_future_permissions is false; [] grants no permissions. Must be omitted or empty when all_future_permissions is true."},
		"effective_permissions":  schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Current grants returned by the API, including expansion of all_future_permissions."},
		"created_at":             schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "Creation timestamp."},
		"updated_at":             schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
	}}
}
