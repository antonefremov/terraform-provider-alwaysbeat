// Package provider implements the Alwaysbeat Terraform provider:
// checks-as-code over the Alwaysbeat JSON API, authenticated with a Alwaysbeat API key.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/antonefremov/terraform-provider-alwaysbeat/internal/client"
)

// Ensure the implementation satisfies the framework interface.
var _ provider.Provider = &alwaysbeatProvider{}

type alwaysbeatProvider struct {
	version string
}

// New returns the provider factory used by main and by acceptance tests.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &alwaysbeatProvider{version: version}
	}
}

func (p *alwaysbeatProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "alwaysbeat"
	resp.Version = p.version
}

// providerModel maps the provider configuration block.
type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

func (p *alwaysbeatProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage [Alwaysbeat](https://alwaysbeat.com) cron/heartbeat checks as code.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "API base URL. Defaults to the production endpoint; override for staging/local. May also be set via `ALWAYSBEAT_ENDPOINT`.",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Alwaysbeat API key (`dmf_...`), created in the dashboard under **API keys**. May also be set via `ALWAYSBEAT_API_KEY` (preferred, keeps it out of state/config).",
			},
		},
	}
}

func (p *alwaysbeatProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Precedence: explicit config value, then environment variable, then
	// (for endpoint only) the built-in production default.
	endpoint := os.Getenv("ALWAYSBEAT_ENDPOINT")
	if !cfg.Endpoint.IsNull() {
		endpoint = cfg.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = client.DefaultEndpoint
	}

	apiKey := os.Getenv("ALWAYSBEAT_API_KEY")
	if !cfg.APIKey.IsNull() {
		apiKey = cfg.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API key",
			"Set the provider `api_key` argument or the ALWAYSBEAT_API_KEY environment variable. Create a key in the dashboard under \"API keys\".",
		)
		return
	}

	c := client.New(endpoint, apiKey)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *alwaysbeatProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCheckResource,
	}
}

func (p *alwaysbeatProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCheckDataSource,
	}
}
