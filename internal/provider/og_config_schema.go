package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (*OGConfigResource) Metadata(_ context.Context, _ resource.MetadataRequest, r *resource.MetadataResponse) {
	r.TypeName = "htmlcsstoimage_og_config"
}
func (*OGConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, r *resource.SchemaResponse) {
	stable := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	r.Schema = schema.Schema{MarkdownDescription: "Manage an Open Graph image configuration. Changes, including configuration type, update in place. Import by management ID, not domain_id. This resource never renders images. Headers are sensitive and stored in provider state.", Attributes: map[string]schema.Attribute{
		"id":                           schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "Management ID used for CRUD and import."},
		"domain_id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Serving identifier used in OG image URLs. Distinct from the management ID."},
		"config_type":                  schema.StringAttribute{Required: true, MarkdownDescription: "Rendering source: html_css or templated. Changing type updates this configuration in place."},
		"name":                         schema.StringAttribute{Required: true, MarkdownDescription: "Display name, 1–255 characters after trimming."},
		"description":                  schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description, up to 1023 characters. Omission clears it."},
		"base_url":                     schema.StringAttribute{Required: true, MarkdownDescription: "HTTPS origin of the source website, up to 2048 characters, without a path, credentials, query, or fragment."},
		"disabled":                     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Disable serving images. Defaults to false."},
		"optimization_mode":            schema.StringAttribute{Optional: true, MarkdownDescription: "no_optimization keeps dimensions; post_process adapts after rendering; set_viewport renders at social-platform viewport sizes. Omission lets the API choose its default (currently post_process)."},
		"effective_optimization_mode":  schema.StringAttribute{Computed: true, MarkdownDescription: "Optimization mode returned by the API."},
		"refresh_interval_s":           schema.Int64Attribute{Optional: true, MarkdownDescription: "Seconds before cached images may refresh: 1800–31536000, subject to the plan minimum. Omission lets the API choose its default (currently 86400)."},
		"effective_refresh_interval_s": schema.Int64Attribute{Computed: true, MarkdownDescription: "Refresh interval returned by the API, in seconds."},
		"extract_values":               schema.BoolAttribute{Optional: true, MarkdownDescription: "HTML/CSS only. Extract image options from page metadata, overriding default_options. Omission uses false."},
		"default_options":              schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "HTML/CSS only. Default page rendering options. Unspecified fields are sent as null; the API owns render defaults.", Attributes: ogOptionAttributes()},
		"template_id":                  schema.StringAttribute{Optional: true, MarkdownDescription: "Required for templated configurations. Template ID including its t- prefix."},
		"template_version":             schema.Int64Attribute{Optional: true, MarkdownDescription: "Templated only. Positive int64 version. Omission follows the latest version at serving time."},
		"template_values_mapping": schema.ListNestedAttribute{Optional: true, MarkdownDescription: "Templated only. Up to 32 ordered mappings from page metadata to distinct template fields.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"template_key": schema.StringAttribute{Required: true, MarkdownDescription: "Destination template field path, up to 128 characters. Paths must not overlap."},
			"meta_key":     schema.StringAttribute{Optional: true, MarkdownDescription: "Source meta tag name, up to 136 characters. Supply exactly one of meta_key or fallback."},
			"fallback":     schema.StringAttribute{Optional: true, MarkdownDescription: "Metadata fallback: titles or descriptions. This selects page metadata, not a literal value."},
		}}},
		"headers":                   schema.MapAttribute{Optional: true, Sensitive: true, ElementType: types.StringType, MarkdownDescription: "Templated only. HTTP extraction headers, up to 20. Names: 512 ASCII characters; values: 8192 UTF-8 bytes. Omission clears them."},
		"additional_header_origins": schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Templated only. Up to 20 additional exact HTTP(S) origins allowed to receive headers, each up to 512 UTF-8 bytes. Omission clears them."},
		"created_at":                schema.StringAttribute{Computed: true, PlanModifiers: stable, MarkdownDescription: "Creation timestamp."},
		"updated_at":                schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
	}}
}

var ogMappingTypes = map[string]attr.Type{"template_key": types.StringType, "meta_key": types.StringType, "fallback": types.StringType}
var ogMappingType = types.ObjectType{AttrTypes: ogMappingTypes}

func ogOptionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"css":                            schema.StringAttribute{Optional: true, MarkdownDescription: "CSS injected into the loaded page to override its styles."},
		"device_scale":                   schema.Float64Attribute{Optional: true, MarkdownDescription: "Adjusts the pixel ratio used for the screenshot. Minimum: 0.1. Maximum: 3. HTML and template renders default to 2; URL renders default to 1."},
		"max_wait_ms":                    schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets a limit on how long to wait before taking the screenshot when the page continues loading irrelevant content. Minimum: 500. Maximum: 10000 and subject to the account plan limit."},
		"ms_delay":                       schema.Int64Attribute{Optional: true, MarkdownDescription: "Adds extra time in milliseconds before taking the screenshot so JavaScript can execute. Minimum: 0. Maximum: 10000."},
		"render_when_ready":              schema.BoolAttribute{Optional: true, MarkdownDescription: "Waits until the page signals that the screenshot is ready. The image fails if the readiness signal is never sent."},
		"selector":                       schema.StringAttribute{Optional: true, MarkdownDescription: "A CSS selector for an element in the HTML. We’ll crop the image to this specific element."},
		"viewport_height":                schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the height of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."},
		"viewport_width":                 schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the width of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."},
		"disable_twemoji":                schema.BoolAttribute{Optional: true, MarkdownDescription: "Disables the Twemoji fallback and renders emoji using native fonts instead."},
		"color_scheme":                   schema.StringAttribute{Optional: true, MarkdownDescription: "Sets Chrome's preferred color scheme. Valid values: light or dark."},
		"timezone":                       schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the IANA timezone used by Chrome while rendering. Must be a recognized IANA timezone identifier."},
		"block_consent_banners":          schema.BoolAttribute{Optional: true, MarkdownDescription: "Attempt to block cookie/consent banners from displaying."},
		"identify_as_hcti":               schema.BoolAttribute{Optional: true, MarkdownDescription: "Identify the top-level page navigation as an HCTI screenshot request using the X-HCTI-SCREENSHOT header."},
		"headers":                        schema.MapAttribute{Optional: true, ElementType: types.StringType, Sensitive: true, MarkdownDescription: "HTTP headers to include on top-level page navigations to the requested URL's origin and any additional_header_origins. Supports up to 20 headers with names up to 512 ASCII characters and values up to 8192 UTF-8 bytes."},
		"additional_header_origins":      schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Additional exact HTTP or HTTPS origins allowed to receive custom headers. Supports up to 20 unique origins of up to 512 UTF-8 bytes each. Origins must use the format scheme://host[:port] without a path; duplicates are ignored."},
		"include_headers_on_subrequests": schema.BoolAttribute{Optional: true, MarkdownDescription: "Include custom headers on subrequests to the requested URL's origin and any additional_header_origins. Defaults to false. Requires at least one header."},
		"viewport_mobile":                schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the page uses mobile viewport behavior, including its viewport meta tag."},
		"viewport_landscape":             schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport is in landscape orientation."},
		"viewport_touch":                 schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport supports touch events."},
		"media_type":                     schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the CSS media type used while rendering the page. Valid values: print or screen."},
		"proxy_id":                       schema.StringAttribute{Optional: true, MarkdownDescription: "Configured proxy ID. Must refer to an enabled proxy."},
		"storage_destination_id":         schema.StringAttribute{Optional: true, MarkdownDescription: "Configured storage destination ID. Must be enabled and allow HCTI storage."},
		"transparent_background":         schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the image is rendered with a transparent background."},
	}
}

func ogOptionTypes() map[string]attr.Type {
	result := map[string]attr.Type{}
	for name, a := range ogOptionAttributes() {
		result[name] = a.GetType()
	}
	return result
}
