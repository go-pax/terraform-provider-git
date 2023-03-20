package git

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/shurcooL/githubv4"
)

type Config struct {
	Token    string
	Owner    string
	Org      string
	Insecure bool
}

type Owner struct {
	name           string
	id             int64
	client         *githubv4.Client
	Context        context.Context
	IsOrganization bool
	token          string
}

// Meta returns the meta parameter that is passed into subsequent resources
// https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ConfigureFunc
func (c *Config) Meta() (interface{}, error) {

	var client *http.Client
	if c.Anonymous() {
		client = c.AnonymousHTTPClient()
	} else {
		client = c.AuthenticatedHTTPClient()
	}

	qlClient, err := c.NewGraphQLClient(client)
	if err != nil {
		return nil, err
	}

	var owner Owner
	owner.client = qlClient

	owner.token = c.Token

	if c.Anonymous() {
		log.Printf("[INFO] No token present; configuring anonymous owner.")
		return &owner, nil
	} else {
		_, err = c.ConfigureOwner(&owner)
		if err != nil {
			return &owner, err
		}
		log.Printf("[INFO] Token present; configuring authenticated owner: %s", owner.name)
		return &owner, nil
	}
}

func (c *Config) ConfigureOwner(owner *Owner) (*Owner, error) {

	ctx := context.Background()

	owner.name = c.Owner
	if owner.name == "" {
		var query struct {
			Viewer struct {
				Login string
			}
		}
		err := owner.client.Query(ctx, &query, nil)
		if err != nil {
			return nil, err
		}
		owner.name = query.Viewer.Login
	}

	return owner, nil
}

func (c *Config) Anonymous() bool {
	return c.Token == ""
}

func (c *Config) AnonymousHTTPClient() *http.Client {
	client := &http.Client{Transport: &http.Transport{}}
	return HTTPClient(client)
}

func (c *Config) AuthenticatedHTTPClient() *http.Client {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	client := oauth2.NewClient(ctx, ts)

	return HTTPClient(client)
}

func HTTPClient(client *http.Client) *http.Client {

	client.Transport = NewEtagTransport(client.Transport)
	client.Transport = logging.NewLoggingHTTPTransport(client.Transport)

	return client
}

func (c *Config) NewGraphQLClient(client *http.Client) (*githubv4.Client, error) {

	uv4, err := url.Parse("https://api.github.com/graphql")
	if err != nil {
		return nil, err
	}

	return githubv4.NewEnterpriseClient(uv4.String(), client), nil
}
