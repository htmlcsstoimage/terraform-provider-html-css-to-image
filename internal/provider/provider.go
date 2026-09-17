package provider

import (
	"context"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type Provider struct{ version, userAgent string }
type config struct {
	APIID   types.String `tfsdk:"api_id"`
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return NewWithUserAgent(version, "HCTITerraform/"+version)
}

func NewWithUserAgent(version, userAgent string) func() provider.Provider {
	return func() provider.Provider { return &Provider{version: version, userAgent: userAgent} }
}
func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, r *provider.MetadataResponse) {
	r.TypeName = "htmlcsstoimage"
	r.Version = p.version
}
func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, r *provider.SchemaResponse) {
	r.Schema = schema.Schema{MarkdownDescription: "Manage HTML/CSS to Image resources.", Attributes: map[string]schema.Attribute{
		"api_id":   schema.StringAttribute{Optional: true, MarkdownDescription: "API ID. Falls back to HCTI_API_ID."},
		"api_key":  schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "API key. Falls back to HCTI_API_KEY."},
		"base_url": schema.StringAttribute{Optional: true, MarkdownDescription: "API origin. Defaults to https://hcti.io."},
	}}
}
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var c config
	resp.Diagnostics.Append(req.Config.Get(ctx, &c)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if c.APIID.IsUnknown() || c.APIKey.IsUnknown() || c.BaseURL.IsUnknown() {
		resp.Diagnostics.AddError("Unknown provider configuration", "API credentials and base_url must be known before configuring the provider.")
		return
	}
	id, key, origin := os.Getenv("HCTI_API_ID"), os.Getenv("HCTI_API_KEY"), "https://hcti.io"
	if !c.APIID.IsNull() {
		id = c.APIID.ValueString()
	}
	if !c.APIKey.IsNull() {
		key = c.APIKey.ValueString()
	}
	if !c.BaseURL.IsNull() {
		origin = c.BaseURL.ValueString()
	}
	if id == "" || key == "" {
		resp.Diagnostics.AddError("Missing API credentials", "Set api_id and api_key or HCTI_API_ID and HCTI_API_KEY.")
		return
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		resp.Diagnostics.AddError("Invalid API origin", "base_url must be an HTTP(S) origin without credentials, a path, query, or fragment.")
		return
	}
	client := management.NewClient(id, key, management.WithBaseURL(origin), management.WithUserAgentSuffix(p.userAgent))
	resp.ResourceData = client
	resp.DataSourceData = &dataSourceClients{management: client, public: hcti.NewClient(id, key, hcti.WithBaseURL(origin), hcti.WithUserAgentSuffix(p.userAgent))}
}
func (*Provider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewProxyResource, NewAPIKeyResource, NewOGConfigResource, NewStorageDestinationResource, NewTemplateResource, NewImageHTMLCSSResource, NewImageURLResource, NewImageTemplatedResource}
}
func (*Provider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewAWSStorageExternalIDDataSource, NewTemplateDataSource, NewTemplateVersionsDataSource}
}
