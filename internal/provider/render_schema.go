package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *RenderResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "htmlcsstoimage_" + r.resourceName()
}
func (r *RenderResource) resourceName() string {
	if r.kind == "template" {
		return "template"
	}
	return "image_" + r.kind
}
func renderAttributes(kind string) map[string]schema.Attribute {
	a := map[string]schema.Attribute{}
	switch kind {
	case "template":
		a["html"] = schema.StringAttribute{Required: true, MarkdownDescription: "HTML to render for the template. Use Handlebars placeholders for values that will be substituted when an image is rendered. Must be non-empty, contain at least one Handlebars placeholder, and compile as valid Handlebars."}
		a["name"] = schema.StringAttribute{Optional: true, MarkdownDescription: "The name of the template, used to identify it in your account. Maximum: 64 characters."}
		a["description"] = schema.StringAttribute{Optional: true, MarkdownDescription: "An optional description of the template for your reference. Maximum: 1024 characters."}
		a["css"] = schema.StringAttribute{Optional: true, MarkdownDescription: "CSS used to style images rendered from the template. Handlebars expressions are not supported in CSS; put dynamic CSS in the HTML instead."}
		a["device_scale"] = schema.Float64Attribute{Optional: true, MarkdownDescription: "Adjusts the pixel ratio used for the screenshot. Minimum: 0.1. Maximum: 3. HTML and template renders default to 2; URL renders default to 1."}
		a["google_fonts"] = schema.SetAttribute{Optional: true, MarkdownDescription: "Google fonts to load. Separate multiple fonts with a pipe, such as 'Roboto|OpenSans', and set font-family in the CSS to use them.", ElementType: types.StringType}
		a["max_wait_ms"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets a limit on how long to wait before taking the screenshot when the page continues loading irrelevant content. Minimum: 500. Maximum: 10000 and subject to the account plan limit."}
		a["ms_delay"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Adds extra time in milliseconds before taking the screenshot so JavaScript can execute. Minimum: 0. Maximum: 10000."}
		a["render_when_ready"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Waits until the page signals that the screenshot is ready. The image fails if the readiness signal is never sent."}
		a["selector"] = schema.StringAttribute{Optional: true, MarkdownDescription: "A CSS selector for an element in the HTML. We’ll crop the image to this specific element."}
		a["viewport_height"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the height of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."}
		a["viewport_width"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the width of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."}
		a["disable_twemoji"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Disables the Twemoji fallback and renders emoji using native fonts instead."}
		a["color_scheme"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the preferred color scheme: light or dark."}
		a["timezone"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the IANA timezone used by Chrome while rendering. Must be a recognized IANA timezone identifier."}
		a["viewport_mobile"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the page uses mobile viewport behavior, including its viewport meta tag."}
		a["viewport_landscape"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport is in landscape orientation."}
		a["viewport_touch"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport supports touch events."}
		a["media_type"] = schema.StringAttribute{Optional: true, MarkdownDescription: "CSS media type to emulate: screen or print."}
		a["proxy_id"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Specifies which configured organization proxy to use when rendering."}
		a["storage_destination_id"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Specifies which configured organization storage destination receives the rendered image."}
		a["jumbo_max_height"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum height of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_width. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels."}
		a["jumbo_max_width"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum width of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_height. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels."}
		a["transparent_background"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the image is rendered with a transparent background."}
	case "html_css":
		a["html"] = schema.StringAttribute{Required: true, MarkdownDescription: "HTML to render and take a screenshot of. HTML fragments are rendered in a wrapper document unless a complete HTML document is supplied. Required for HTML image requests."}
		a["css"] = schema.StringAttribute{Optional: true, MarkdownDescription: "CSS used to style the rendered HTML."}
		a["device_scale"] = schema.Float64Attribute{Optional: true, MarkdownDescription: "Adjusts the pixel ratio used for the screenshot. Minimum: 0.1. Maximum: 3. HTML and template renders default to 2; URL renders default to 1."}
		a["google_fonts"] = schema.SetAttribute{Optional: true, MarkdownDescription: "Google fonts to load. Separate multiple fonts with a pipe, such as 'Roboto|OpenSans', and set font-family in the CSS to use them.", ElementType: types.StringType}
		a["max_wait_ms"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets a limit on how long to wait before taking the screenshot when the page continues loading irrelevant content. Minimum: 500. Maximum: 10000 and subject to the account plan limit."}
		a["metadata"] = schema.MapAttribute{Optional: true, MarkdownDescription: "Custom key-value metadata stored with the image.", ElementType: types.StringType}
		a["ms_delay"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Adds extra time in milliseconds before taking the screenshot so JavaScript can execute. Minimum: 0. Maximum: 10000."}
		a["render_when_ready"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Waits until the page signals that the screenshot is ready. The image fails if the readiness signal is never sent."}
		a["selector"] = schema.StringAttribute{Optional: true, MarkdownDescription: "A CSS selector for an element in the HTML. We’ll crop the image to this specific element."}
		a["viewport_height"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the height of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."}
		a["viewport_width"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the width of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."}
		a["pdf_options"] = schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "Rendering option.", Attributes: renderPDFAttributes()}
		a["disable_twemoji"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Disables the Twemoji fallback and renders emoji using native fonts instead."}
		a["max_render_once"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Ensure the image is only ever rendered and saved one time. This is an advanced option not applicable to most requests."}
		a["color_scheme"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the preferred color scheme: light or dark."}
		a["timezone"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the IANA timezone used by Chrome while rendering. Must be a recognized IANA timezone identifier."}
		a["viewport_mobile"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the page uses mobile viewport behavior, including its viewport meta tag."}
		a["viewport_landscape"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport is in landscape orientation."}
		a["viewport_touch"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport supports touch events."}
		a["media_type"] = schema.StringAttribute{Optional: true, MarkdownDescription: "CSS media type to emulate: screen or print."}
		a["proxy_id"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Specifies which configured organization proxy to use when rendering."}
		a["storage_destination_id"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Specifies which configured organization storage destination receives the rendered image."}
		a["jumbo_max_height"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum height of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_width. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels."}
		a["jumbo_max_width"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum width of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_height. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels."}
		a["transparent_background"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the image is rendered with a transparent background."}
		a["format"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Rendering URL format: png, jpg, jpeg, webp, or pdf. Does not restrict later renders to this format."}
	case "url":
		a["url"] = schema.StringAttribute{Required: true, MarkdownDescription: "Public HTTP or HTTPS URL to capture. Required for URL image requests."}
		a["css"] = schema.StringAttribute{Optional: true, MarkdownDescription: "CSS injected into the loaded URL to override styles on the page."}
		a["device_scale"] = schema.Float64Attribute{Optional: true, MarkdownDescription: "Adjusts the pixel ratio used for the screenshot. Minimum: 0.1. Maximum: 3. HTML and template renders default to 2; URL renders default to 1."}
		a["full_screen"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Take a screenshot of the entire screen after scrolling down and back to the top."}
		a["max_wait_ms"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets a limit on how long to wait before taking the screenshot when the page continues loading irrelevant content. Minimum: 500. Maximum: 10000 and subject to the account plan limit."}
		a["metadata"] = schema.MapAttribute{Optional: true, MarkdownDescription: "Custom key-value metadata stored with the image.", ElementType: types.StringType}
		a["ms_delay"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Adds extra time in milliseconds before taking the screenshot so JavaScript can execute. Minimum: 0. Maximum: 10000."}
		a["render_when_ready"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Waits until the page signals that the screenshot is ready. The image fails if the readiness signal is never sent."}
		a["selector"] = schema.StringAttribute{Optional: true, MarkdownDescription: "A CSS selector for an element in the HTML. We’ll crop the image to this specific element."}
		a["viewport_height"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the height of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."}
		a["viewport_width"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Sets the width of Chrome's viewport and disables automatic cropping. Minimum: 1. Maximum: 6000. Both viewport dimensions must be supplied together."}
		a["pdf_options"] = schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "Rendering option.", Attributes: renderPDFAttributes()}
		a["disable_twemoji"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Disables the Twemoji fallback and renders emoji using native fonts instead."}
		a["max_render_once"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Ensure the image is only ever rendered and saved one time. This is an advanced option not applicable to most requests."}
		a["color_scheme"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the preferred color scheme: light or dark."}
		a["timezone"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Sets the IANA timezone used by Chrome while rendering. Must be a recognized IANA timezone identifier."}
		a["block_consent_banners"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Attempt to block cookie/consent banners from displaying."}
		a["identify_as_hcti"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Identify the top-level page navigation as an HCTI screenshot request using the X-HCTI-SCREENSHOT header."}
		a["headers"] = schema.MapAttribute{Optional: true, MarkdownDescription: "HTTP headers to include on top-level page navigations to the requested URL's origin and any additional_header_origins. Supports up to 20 headers with names up to 512 ASCII characters and values up to 8192 UTF-8 bytes. For GET and form-encoded requests, repeat this parameter using the format `headers=name:value`.", Sensitive: true, ElementType: types.StringType}
		a["additional_header_origins"] = schema.SetAttribute{Optional: true, MarkdownDescription: "Additional exact HTTP or HTTPS origins allowed to receive custom headers. Supports up to 20 unique origins of up to 512 UTF-8 bytes each. Origins must use the format scheme://host[:port] without a path; duplicates are ignored. For GET and form-encoded requests, repeat this parameter for each origin.", ElementType: types.StringType}
		a["include_headers_on_subrequests"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Include custom headers on subrequests to the requested URL's origin and any additional_header_origins. Defaults to false. Requires at least one header."}
		a["viewport_mobile"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the page uses mobile viewport behavior, including its viewport meta tag."}
		a["viewport_landscape"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport is in landscape orientation."}
		a["viewport_touch"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the emulated viewport supports touch events."}
		a["media_type"] = schema.StringAttribute{Optional: true, MarkdownDescription: "CSS media type to emulate: screen or print."}
		a["proxy_id"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Specifies which configured organization proxy to use when rendering."}
		a["storage_destination_id"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Specifies which configured organization storage destination receives the rendered image."}
		a["jumbo_max_height"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum height of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_width. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels."}
		a["jumbo_max_width"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum width of the rendered image in jumbo mode. Jumbo rendering consumes additional renders and requires jumbo_max_height. Supply both jumbo dimensions. Each must be greater than 0 and no more than 80000; at least one must exceed 8000; total area cannot exceed 400000000 pixels."}
		a["transparent_background"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Specifies whether the image is rendered with a transparent background."}
		a["format"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Rendering URL format: png, jpg, jpeg, webp, or pdf. Does not restrict later renders to this format."}
	case "templated":
		a["template_values"] = schema.StringAttribute{Required: true, MarkdownDescription: "Values substituted into the template for this render. Must be a non-empty JSON object. Include the values needed by the template.", Sensitive: true}
		a["format"] = schema.StringAttribute{Optional: true, MarkdownDescription: "Rendering URL format: png, jpg, jpeg, webp, or pdf. Does not restrict later renders to this format."}
		a["template_id"] = schema.StringAttribute{Required: true, MarkdownDescription: "Template ID including the t- prefix."}
		a["template_version"] = schema.Int64Attribute{Optional: true, MarkdownDescription: "Optional version to pin. Omission selects the latest version when creating the image."}
	}
	return a
}
func renderPDFAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"page_width":       schema.StringAttribute{Optional: true, MarkdownDescription: "Page width with px, in, cm, mm, or pt units."},
		"page_height":      schema.StringAttribute{Optional: true, MarkdownDescription: "Page height with px, in, cm, mm, or pt units."},
		"scale":            schema.Float64Attribute{Optional: true, MarkdownDescription: "PDF scale from 0.1 to 2."},
		"print_background": schema.BoolAttribute{Optional: true, MarkdownDescription: "Print page backgrounds. Omission lets the API choose."},
		"margins": schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "PDF margins. Supply all four sides with units.", Attributes: map[string]schema.Attribute{
			"top":    schema.StringAttribute{Required: true, MarkdownDescription: "Margin with px, in, cm, mm, or pt units."},
			"right":  schema.StringAttribute{Required: true, MarkdownDescription: "Margin with px, in, cm, mm, or pt units."},
			"bottom": schema.StringAttribute{Required: true, MarkdownDescription: "Margin with px, in, cm, mm, or pt units."},
			"left":   schema.StringAttribute{Required: true, MarkdownDescription: "Margin with px, in, cm, mm, or pt units."},
		}}}
}
func (r *RenderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	a := renderAttributes(r.kind)
	a["id"] = schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Stable resource ID. Import using this ID."}
	a["created_at"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."}
	if r.kind == "template" {
		a["version"] = schema.Int64Attribute{Computed: true, MarkdownDescription: "Latest saved template version."}
		a["template_type"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Template type; this resource manages html_css templates."}
		a["updated_at"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."}
	} else {
		a["image_url"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Rendering operation URL. Creation does not render bytes."}
		a["public_url"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Public GET rendering URL, or null when HCTI storage is disabled."}
		a["render_method"] = schema.StringAttribute{Computed: true, MarkdownDescription: "HTTP method required to render: GET or PUT."}
		a["last_render_stored_at"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Last HCTI stored-render timestamp, when available."}
		a["saved_to_storage_destination_at"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Last base-image save to custom storage, when available."}
		a["og_config_id"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Associated OG configuration ID, if any."}
		a["render_requires_auth"] = schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether rendering requires API authentication."}
		a["og_config_content_version"] = schema.Int64Attribute{Computed: true, MarkdownDescription: "Associated OG content version, if any."}
		if r.kind == "templated" {
			a["resolved_template_version"] = schema.Int64Attribute{Computed: true, MarkdownDescription: "Actual template version used at creation; does not pin an omitted template_version."}
		}
		a["storage_destination_hcti_storage_disabled"] = schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether rendered output is saved only to custom storage, recorded when the image was created."}
	}
	resp.Schema = schema.Schema{MarkdownDescription: "Manage a saved rendering definition. Images are created without rendering and replaced when inputs change. Templates retain their ID and create a new version on edits. Unspecified nullable inputs are sent as null; the API owns defaults.", Attributes: a}
}
