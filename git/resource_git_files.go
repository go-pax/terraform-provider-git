package git

import (
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/map_type"
	"github.com/go-pax/terraform-provider-git/utils/unique"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"path"
	"strings"
)

type Author struct {
	name    string
	email   string
	message string
}

func resourceGitFiles() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			/*"checkout_dir": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},*/
			/*"filepath": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"contents": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},*/
			/*"author": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},*/
			"author": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"file": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filepath": {
							Type:     schema.TypeString,
							Required: true,
						},
						"contents": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		Create: fileCreateUpdate,
		Read:   fileRead,
		Delete: fileDelete,
		Exists: fileExists,
	}
}

func fileCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	hostname := d.Get("hostname").(string)
	org := d.Get("organization").(string)
	branch := d.Get("branch").(string)
	repo := d.Get("repository").(string)

	checkout_dir := unique.UniqueId()
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	var author map[string]string
	var val, aok = d.GetOk("author")
	if aok {
		author = map_type.ToTypedObject(val.(map[string]interface{}))
	} else {
		return fmt.Errorf("author argument missing")
	}

	var err error
	if err = commands.configureAuthor(author["name"], author["email"]); err != nil {
		return err
	}

	if _, err = commands.checkout(checkout_dir, repo, branch); err != nil {
		return err
	}

	files := d.Get("file")
	for _, v := range files.(*schema.Set).List() {
		file := map_type.ToTypedObject(v.(map[string]interface{}))
		filepath := file["filepath"]
		contents := file["contents"]

		if err := os.MkdirAll(path.Dir(path.Join(checkout_dir, filepath)), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(checkout_dir, filepath), []byte(contents), 0666); err != nil {
			return err
		}

		if _, err := gitCommand(checkout_dir, "add", "--", filepath); err != nil {
			return err
		}

		commit_message := author["message"]
		commit_body := fmt.Sprintf("The following files are managed by terraform:\n%s", filepath)
		if _, err := gitCommand(checkout_dir, flatten("commit", "-m", commit_message, "-m", commit_body, "--allow-empty", "--", filepath)...); err != nil {
			return err
		}
	}

	if _, err := gitCommand(checkout_dir, "push", "origin", "HEAD"); err != nil {
		return err
	}

	var sha string
	if out, err := gitCommand(checkout_dir, "rev-parse", "HEAD"); err != nil {
		return err
	} else {
		sha = strings.TrimRight(string(out), "\n")
	}

	d.SetId(fmt.Sprintf("%s", sha))
	return nil
}

func fileRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func fileExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	hostname := d.Get("hostname").(string)
	org := d.Get("organization").(string)
	branch := d.Get("branch").(string)
	repo := d.Get("repository").(string)

	checkout_dir := unique.UniqueId()
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	a := d.Get("author")
	author := map_type.ToTypedObject(a.(map[string]interface{}))

	var err error
	if err = commands.configureAuthor(author["name"], author["email"]); err != nil {
		return false, err
	}

	if _, err = commands.checkout(checkout_dir, repo, branch); err != nil {
		return false, err
	}

	files := d.Get("file")
	for _, v := range files.(*schema.Set).List() {
		file := map_type.ToTypedObject(v.(map[string]interface{}))
		filepath := file["filepath"]
		contents := file["contents"]

		var out []byte
		var err error
		if out, err = os.ReadFile(path.Join(checkout_dir, filepath)); err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if string(out) != contents {
			return false, nil
		}
	}
	return true, nil
}

func fileDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
