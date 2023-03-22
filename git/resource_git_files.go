package git

import (
	"context"
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/map_type"
	"github.com/go-pax/terraform-provider-git/utils/unique"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
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
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
	}
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("failed to configure author %s: %s", author["name"], author["email"])
	}

	if _, err = commands.checkout(checkout_dir, repo, branch); err != nil {
		return diag.Errorf("failed to checkout branch %s: %s", branch, repo)
	}

	files := d.Get("file")
	is_clean := true
	for _, v := range files.(*schema.Set).List() {
		file := map_type.ToTypedObject(v.(map[string]interface{}))
		filepath := file["filepath"]

		if err := os.Remove(path.Join(checkout_dir, filepath)); err != nil {
			return diag.Errorf("failed to delete file %s: %s", filepath, err)
		}
		is_clean = false
	}

	if is_clean {
		return nil
	}

	if _, err := gitCommand(checkout_dir, "add", "--", "."); err != nil {
		return diag.Errorf("failed to add files to git: %s", err)
	}
	if _, err := gitCommand(checkout_dir, flatten("commit", "-m", "automated delete", "--allow-empty", "--")...); err != nil {
		return diag.Errorf("failed to commit to git: %s", err)
	}
	if _, err := gitCommand(checkout_dir, "push", "origin", "HEAD"); err != nil {
		return diag.Errorf("failed to push commit")
	}
	return nil
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("failed to configure author %s: %s", author["name"], author["email"])
	}

	if _, err = commands.checkout(checkout_dir, repo, branch); err != nil {
		return diag.Errorf("failed to checkout branch %s: %s", branch, repo)
	}

	is_clean := true
	updated := []string{}
	if d.HasChange("file") {
		files, _ := d.GetChange("file")

		for _, v := range files.(*schema.Set).List() {
			file := map_type.ToTypedObject(v.(map[string]interface{}))
			filepath := file["filepath"]

			if err := os.Remove(path.Join(checkout_dir, filepath)); err != nil {
				return diag.Errorf("failed to delete file %s: %s", filepath, err)
			}
			updated = append(updated, filepath)
		}
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
				is_clean = false
				if err := os.MkdirAll(path.Dir(path.Join(checkout_dir, filepath)), 0755); err != nil {
					return diag.Errorf("failed to create file directory: %s", filepath)
				}
				if err := os.WriteFile(path.Join(checkout_dir, filepath), []byte(contents), 0666); err != nil {
					return diag.Errorf("failed to create file: %s", filepath)
				}
				if _, err := gitCommand(checkout_dir, "add", "--", filepath); err != nil {
					return diag.Errorf("failed to add file to git: %s", filepath)
				}
				updated = append(updated, filepath)
				continue
			}
			return diag.Errorf("General os error: %s", filepath)
		}
		if string(out) != contents {
			log.Printf("[INFO] File contents changed: %s", filepath)
			is_clean = false
			if err := os.WriteFile(path.Join(checkout_dir, filepath), []byte(contents), 0666); err != nil {
				return diag.Errorf("failed to update file: %s", filepath)
			}
			if _, err := gitCommand(checkout_dir, "add", "--", filepath); err != nil {
				return diag.Errorf("failed to update file to git: %s", filepath)
			}
			updated = append(updated, filepath)
		}
	}

	if is_clean {
		var sha string
		if out, err := gitCommand(checkout_dir, "rev-parse", "HEAD"); err != nil {
			return diag.Errorf("failed to get revision")
		} else {
			sha = strings.TrimRight(string(out), "\n")
		}
		d.SetId(sha)
		return nil
	}

	commit_message := author["message"]
	commit_body := fmt.Sprintf("The following files are managed by terraform:\n%s", strings.Join(updated, "\n"))
	if _, err := gitCommand(checkout_dir, flatten("commit", "-m", commit_message, "-m", commit_body, "--allow-empty", "--")...); err != nil {
		return diag.Errorf("failed to commit to git: %s", err)
	}
	if _, err := gitCommand(checkout_dir, "push", "origin", "HEAD"); err != nil {
		return diag.Errorf("failed to push commit")
	}
	var sha string
	if out, err := gitCommand(checkout_dir, "rev-parse", "HEAD"); err != nil {
		return diag.Errorf("failed to get revision")
	} else {
		sha = strings.TrimRight(string(out), "\n")
	}
	d.SetId(sha)
	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("failed to configure author %s: %s", author["name"], author["email"])
	}

	if _, err = commands.checkout(checkout_dir, repo, branch); err != nil {
		return diag.Errorf("failed to checkout branch %s: %s", branch, err)
	}

	files := d.Get("file")
	for _, v := range files.(*schema.Set).List() {
		file := map_type.ToTypedObject(v.(map[string]interface{}))
		filepath := file["filepath"]
		contents := file["contents"]

		if err := os.MkdirAll(path.Dir(path.Join(checkout_dir, filepath)), 0755); err != nil {
			return diag.Errorf("failed to create file directory: %s", filepath)
		}
		if err := os.WriteFile(path.Join(checkout_dir, filepath), []byte(contents), 0666); err != nil {
			return diag.Errorf("failed to create file: %s", filepath)
		}

		if _, err := gitCommand(checkout_dir, "add", "--", filepath); err != nil {
			return diag.Errorf("failed to add file to git: %s", filepath)
		}

		commit_message := author["message"]
		commit_body := fmt.Sprintf("The following files are managed by terraform:\n%s", filepath)
		if _, err := gitCommand(checkout_dir, flatten("commit", "-m", commit_message, "-m", commit_body, "--allow-empty", "--", filepath)...); err != nil {
			return diag.Errorf("failed to commit file to git: %s", filepath)
		}
	}

	if _, err := gitCommand(checkout_dir, "push", "origin", "HEAD"); err != nil {
		return diag.Errorf("failed to push commit")
	}

	var sha string
	if out, err := gitCommand(checkout_dir, "rev-parse", "HEAD"); err != nil {
		return diag.Errorf("failed to get revision")
	} else {
		sha = strings.TrimRight(string(out), "\n")
	}

	d.SetId(sha)
	return nil
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	hostname := d.Get("hostname").(string)
	org := d.Get("organization").(string)
	branch := d.Get("branch").(string)
	repo := d.Get("repository").(string)

	checkout_dir := unique.UniqueId()
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	// a := d.Get("author")
	// author := map_type.ToTypedObject(a.(map[string]interface{}))
	//
	// var err error
	// if err = commands.configureAuthor(author["name"], author["email"]); err != nil {
	// 	return false, err
	// }

	if _, err := commands.checkout(checkout_dir, repo, branch); err != nil {
		return diag.Errorf("failed to checkout branch %s: %s", branch, repo)
	}

	files := d.Get("file")
	is_clean := true
	for _, v := range files.(*schema.Set).List() {
		file := map_type.ToTypedObject(v.(map[string]interface{}))
		filepath := file["filepath"]
		contents := file["contents"]

		var out []byte
		var err error
		if out, err = os.ReadFile(path.Join(checkout_dir, filepath)); err != nil {
			if os.IsNotExist(err) {
				log.Printf("[INFO] Expected file doesn't exist: %s", filepath)
				is_clean = false
			}
			return diag.Errorf("General os error: %s", filepath)
		}
		if string(out) != contents {
			log.Printf("[INFO] File contents changed: %s", filepath)
			is_clean = false
		}
	}

	if !is_clean {
		d.SetId("")
		return nil
	}

	if out, err := gitCommand(checkout_dir, "rev-parse", "HEAD"); err != nil {
		return diag.Errorf("Unable to get revision git.")
	} else {
		id := d.Id()
		log.Printf("[INFO] Remote branch revision: %s", id)

		sha := strings.TrimRight(string(out), "\n")
		log.Printf("[INFO] Local branch revision: %s", sha)

		if id != sha {
			log.Printf("[INFO] Remote revision not the same as local revision: %s", id)
			d.SetId(sha)
			return nil
		}
	}

	d.SetId(d.Id())
	return nil
}
