package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/htmlcsstoimage/go-client/management"
)

type AWSStorageExternalIDDataSource struct{ client *management.Client }

var _ datasource.DataSourceWithConfigure = (*AWSStorageExternalIDDataSource)(nil)

func NewAWSStorageExternalIDDataSource() datasource.DataSource {
	return &AWSStorageExternalIDDataSource{}
}
func (*AWSStorageExternalIDDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, r *datasource.MetadataResponse) {
	r.TypeName = "htmlcsstoimage_aws_storage_external_id"
}
func (*AWSStorageExternalIDDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, r *datasource.SchemaResponse) {
	r.Schema = schema.Schema{MarkdownDescription: "Read the organization's AWS external ID and HCTI writer role ARN for an IAM role trust policy before creating a storage destination. This does not create a destination or test a connection.", Attributes: map[string]schema.Attribute{
		"external_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Organization-specific external ID to require in the AWS role trust policy's sts:ExternalId condition."},
		"writer_role_arn": schema.StringAttribute{Computed: true, MarkdownDescription: "HCTI IAM writer role ARN to allow as the AWS principal in the role trust policy."},
	}}
}
func (r *AWSStorageExternalIDDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*dataSourceClients)
	if !ok {
		resp.Diagnostics.AddError("Invalid provider configuration", "Expected an HCTI management client.")
		return
	}
	r.client = c.management
}
func (r *AWSStorageExternalIDDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	v, err := r.client.GetAWSExternalID(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read AWS external ID", apiErrorDetail(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		ExternalID    types.String `tfsdk:"external_id"`
		WriterRoleARN types.String `tfsdk:"writer_role_arn"`
	}{types.StringValue(v.ExternalID), types.StringValue(v.WriterRoleARN)})...)
}
