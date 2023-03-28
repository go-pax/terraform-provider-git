package git

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

var providerFactories = map[string]func() (*schema.Provider, error){
	"git": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func init() {
	testAccProvider = Provider() // .(*schema.Provider)
	testAccProviders = map[string]*schema.Provider{
		"git": testAccProvider,
	}
}
