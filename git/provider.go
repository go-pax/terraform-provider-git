package git

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_TOKEN", nil),
				Description: descriptions["token"],
			},
			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_OWNER", nil),
				Description: descriptions["owner"],
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_ORGANIZATION", nil),
				Description: descriptions["organization"],
				Deprecated:  "Use owner (or GITHUB_OWNER) instead of organization (or GITHUB_ORGANIZATION)",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["insecure"],
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			//"github_release": resourceGithubRelease(),
			"git_files": resourceGitFiles(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			//"gitfile_release": dataSourceGithubRelease(),
		},
	}

	p.ConfigureContextFunc = providerConfigure(p)

	return p
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The PAT used to connect to GitHub. Anonymous mode is enabled if `token` is not set.",
		"owner": "The GitHub owner name to manage. " +
			"Use this field instead of `organization` when managing individual accounts.",
		"organization": "The GitHub organization name to manage. " +
			"Use this field instead of `owner` when managing organization accounts.",
		"insecure": "Enable `insecure` mode for testing purposes",
	}
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		owner := d.Get("owner").(string)
		token := d.Get("token").(string)
		insecure := d.Get("insecure").(bool)
		org := d.Get("organization").(string)
		if org != "" {
			log.Printf("[INFO] Selecting organization attribute as owner: %s", org)
			owner = org
		}

		config := Config{
			Token:    token,
			Insecure: insecure,
			Owner:    owner,
		}

		meta, err := config.Meta()
		if err != nil {
			diags = append(
				diags,
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to set config metadata",
					Detail:   "failed to return meta without error: " + err.Error(),
				},
			)
			return nil, diags
		}

		meta.(*Owner).Context = ctx

		return meta, nil
	}
}
