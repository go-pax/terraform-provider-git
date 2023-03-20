package git

import (
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/unique"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"path"
)

func resourceGitFiles() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			/*"checkout_dir": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},*/
			"filepath": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"contents": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"author": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Default:  "github.com",
				Optional: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		Create: fileCreateUpdate,
		Read:   fileRead,
		Delete: fileDelete,
		Exists: fileExists,
	}
}

func fileCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	checkout_dir := unique.UniqueId()
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)

	hostname := d.Get("hostname").(string)
	org := d.Get("organization").(string)
	branch := d.Get("branch").(string)
	repo := d.Get("repository").(string)
	filepath := d.Get("filepath").(string)
	contents := d.Get("contents").(string)

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	var head string
	var err error
	if head, err = commands.checkout(checkout_dir, repo, branch); err != nil {
		return err
	}
	println("%s", head)

	if err := os.MkdirAll(path.Dir(path.Join(checkout_dir, filepath)), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(checkout_dir, filepath), []byte(contents), 0666); err != nil {
		return err
	}

	if _, err := gitCommand(checkout_dir, "add", "--", filepath); err != nil {
		return err
	}

	id := fmt.Sprintf("%d-%s", hashString(contents), filepath)

	d.SetId(id)
	return nil
}

func fileRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func fileExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	checkout_dir := d.Get("checkout_dir").(string)
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)
	filepath := d.Get("path").(string)

	var out []byte
	var err error
	if out, err = os.ReadFile(path.Join(checkout_dir, filepath)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	if string(out) == d.Get("contents").(string) {
		return true, nil
	} else {
		return false, nil
	}
}

func fileDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
