package provider

import (
	"context"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type dataSourceClients struct {
	management *management.Client
	public     *hcti.Client
}

type templateDataSource struct {
	versions bool
	clients  *dataSourceClients
}

func NewTemplateDataSource() datasource.DataSource { return &templateDataSource{} }
func NewTemplateVersionsDataSource() datasource.DataSource {
	return &templateDataSource{versions: true}
}
func (d *templateDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "htmlcsstoimage_template"
	if d.versions {
		resp.TypeName += "_versions"
	}
}
func templateVersionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"version":       schema.Int64Attribute{Computed: true, MarkdownDescription: "Exact template version identifier."},
		"name":          schema.StringAttribute{Computed: true, MarkdownDescription: "Template display name for this version."},
		"description":   schema.StringAttribute{Computed: true, MarkdownDescription: "Template description for this version."},
		"template_type": schema.StringAttribute{Computed: true, MarkdownDescription: "Template type returned by the API."},
		"created_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp returned by the API."},
		"updated_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp returned by the API."},
	}
}
func (d *templateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	a := map[string]schema.Attribute{"id": schema.StringAttribute{Required: true, MarkdownDescription: "Existing template ID, including the t- prefix."}}
	if d.versions {
		a["limit"] = schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum number of versions to return, newest first. Defaults to 1000. Must be positive. Pagination stops when this limit is reached."}
		a["versions"] = schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Available versions, newest first, up to limit. Pagination is followed automatically; entries contain metadata without template contents.", NestedObject: schema.NestedAttributeObject{Attributes: templateVersionAttributes()}}
	} else {
		for k, r := range renderAttributes("template") {
			desc := r.GetMarkdownDescription()
			switch t := r.GetType().(type) {
			case basetypes.StringType:
				a[k] = schema.StringAttribute{Computed: true, MarkdownDescription: desc, Sensitive: r.IsSensitive()}
			case basetypes.Int64Type:
				a[k] = schema.Int64Attribute{Computed: true, MarkdownDescription: desc}
			case basetypes.Float64Type:
				a[k] = schema.Float64Attribute{Computed: true, MarkdownDescription: desc}
			case basetypes.BoolType:
				a[k] = schema.BoolAttribute{Computed: true, MarkdownDescription: desc}
			case types.SetType:
				a[k] = schema.SetAttribute{Computed: true, ElementType: t.ElemType, MarkdownDescription: desc}
			default:
				panic("unsupported template lookup attribute type")
			}
		}
		for k, v := range templateVersionAttributes() {
			if _, ok := a[k]; !ok {
				a[k] = v
			}
		}
		a["version"] = schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Version to read. Omit to resolve the latest version on each refresh. Referencing this output in an image's template_version follows template updates."}
	}
	resp.Schema = schema.Schema{MarkdownDescription: "Read an existing template without creating, updating, rendering, or deleting it. Requires templates:read. Lookup failures remain errors.", Attributes: a}
}
func (d *templateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.clients, ok = req.ProviderData.(*dataSourceClients)
	if !ok {
		resp.Diagnostics.AddError("Invalid provider configuration", "Expected HCTI data source clients.")
	}
}
func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a := config.Attributes()
	id := a["id"].(types.String).ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Invalid template ID", "id must not be empty.")
		return
	}
	if d.versions {
		d.readVersions(ctx, id, a, resp)
	} else {
		v := a["version"].(types.Int64)
		var pin *int64
		if !v.IsNull() && !v.IsUnknown() {
			n := v.ValueInt64()
			if n <= 0 {
				resp.Diagnostics.AddError("Invalid template version", "version must be positive.")
				return
			}
			pin = &n
		}
		remote, err := d.clients.management.GetTemplateDefinition(ctx, id, pin)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read template", apiErrorDetail(err))
			return
		}
		if remote.TemplateType != "html_css" {
			resp.Diagnostics.AddError("Unsupported template type", "The template lookup currently supports HTML/CSS templates. Use template_versions to list metadata for other template types.")
			return
		}
		values, diags := renderReadValues(ctx, "template", remote.RenderDefinition)
		resp.Diagnostics.Append(diags...)
		for k, v := range values {
			a[k] = v
		}
		a["version"] = types.Int64Value(remote.Version)
		a["template_type"] = types.StringValue(remote.TemplateType)
		a["created_at"] = types.StringValue(remote.CreatedAt)
		a["updated_at"] = types.StringValue(remote.UpdatedAt)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	result, diags := types.ObjectValue(config.AttributeTypes(ctx), a)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	}
}
func (d *templateDataSource) readVersions(ctx context.Context, id string, a map[string]attr.Value, resp *datasource.ReadResponse) {
	limit := int64(1000)
	if v := a["limit"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		limit = v.ValueInt64()
	}
	if limit <= 0 {
		resp.Diagnostics.AddError("Invalid template versions limit", "limit must be positive.")
		return
	}
	a["limit"] = types.Int64Value(limit)
	options := hcti.TemplateListOptions{Count: 100}
	versions := []hcti.Template{}
	seen := map[int64]bool{}
	for {
		options.Count = int(min(int64(100), limit-int64(len(versions))))
		page, err := d.clients.public.ListTemplateVersions(ctx, id, options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list template versions", apiErrorDetail(err))
			return
		}
		for _, v := range page.Data {
			if v.ID != id || v.Version <= 0 || seen[v.Version] {
				resp.Diagnostics.AddError("Invalid template versions response", "The API returned mismatched or duplicate template versions.")
				return
			}
			seen[v.Version] = true
			versions = append(versions, v)
			if int64(len(versions)) == limit {
				break
			}
		}
		if int64(len(versions)) == limit {
			break
		}
		next := page.Pagination.NextPageStart
		if next == nil {
			break
		}
		if options.MaxVersion != nil && *next >= *options.MaxVersion {
			resp.Diagnostics.AddError("Invalid template pagination", "The API pagination cursor did not advance.")
			return
		}
		options.MaxVersion = next
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].Version > versions[j].Version })
	attrs := map[string]attr.Type{}
	for k, v := range templateVersionAttributes() {
		attrs[k] = v.GetType()
	}
	items := []attr.Value{}
	for _, v := range versions {
		obj, diags := types.ObjectValue(attrs, map[string]attr.Value{
			"version": types.Int64Value(v.Version), "name": types.StringValue(v.Name), "description": types.StringValue(v.Description), "template_type": types.StringValue(v.TemplateType),
			"created_at": types.StringValue(v.CreatedAt.Format(time.RFC3339Nano)), "updated_at": types.StringValue(v.UpdatedAt.Format(time.RFC3339Nano)),
		})
		resp.Diagnostics.Append(diags...)
		items = append(items, obj)
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: attrs}, items)
	resp.Diagnostics.Append(diags...)
	a["versions"] = list
}
