package git

import (
	"context"
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/map_type"
	"github.com/go-pax/terraform-provider-git/utils/set"
	"github.com/go-pax/terraform-provider-git/utils/unique"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "`object({ name=string, email=string, message=string })` Defines the commit user and message.",
			},
			"branch": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "This is the branch the files will commit into. The branch must exist.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Default:     "github.com",
				Optional:    true,
				ForceNew:    true,
				Description: "Defaults to `github.com` but since this is pure git change to whatever server you are committing into.",
			},
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Respository name you want to commit into.",
			},
			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Sets the organization in git the repository is in.",
			},
			"force_new": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Ensure your files are always pushed into the branch. If the branch is generated in the " +
					"apply and doesn't exist yet set this to true",
			},
			"file": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filepath": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Relative path to the file in the targeted repository.",
						},
						"contents": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "String contents of this file. Bested used with templates",
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

	checkout_dir := path.Join(os.TempDir(), unique.UniqueId())
	if err := os.MkdirAll(checkout_dir, 0755); err != nil {
		return diag.Errorf("failed to create git temp dir: %s", err)
	}
	lockCheckout(checkout_dir)
	defer func() {
		unlockCheckout(checkout_dir)
		_ = os.RemoveAll(checkout_dir)
	}()

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	_, status, err := commands.checkout(checkout_dir, repo, branch)
	switch status {
	case Exist:
		tflog.Info(ctx, "Branch exists for deletion")
		break
	case NotExist:
		tflog.Warn(ctx, fmt.Sprintf("Branch already deleted: %s", branch))
		return nil
	case Unknown:
		if err != nil {
			return diag.Errorf("failed to checkout branch %s: %s", branch, err)
		}
	}

	var deleted_files []string
	files := d.Get("file")
	is_clean := true
	for _, v := range files.(*schema.Set).List() {
		file := map_type.ToTypedObject(v.(map[string]interface{}))
		filepath := file["filepath"]

		if err := os.Remove(path.Join(checkout_dir, filepath)); err != nil {
			return diag.Errorf("failed to delete file %s: %s", filepath, err)
		}
		deleted_files = append(deleted_files, filepath)
		is_clean = false
	}

	if is_clean {
		return nil
	}

	if _, err := gitCommand(checkout_dir, "add", "--", "."); err != nil {
		return diag.Errorf("failed to add files to git: %s", err)
	}

	a := d.Get("author")
	author := map_type.ToTypedObject(a.(map[string]interface{}))
	commit_message := author["message"]
	commit_body := fmt.Sprintf("The following files were deleted by terraform:\n%s", strings.Join(deleted_files, "\n"))
	commit_command := flatten("commit", "-m", commit_message, "-m", commit_body, "--allow-empty")
	commit_command = append(commit_command, commands.getAuthorString(author["name"], author["email"])...)
	if _, err := gitCommand(checkout_dir, commit_command...); err != nil {
		return diag.Errorf("failed to commit file(s) to git: %s", err)
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

	checkout_dir := path.Join(os.TempDir(), unique.UniqueId())
	if err := os.MkdirAll(checkout_dir, 0755); err != nil {
		return diag.Errorf("failed to create git temp dir: %s", err)
	}
	lockCheckout(checkout_dir)
	defer func() {
		unlockCheckout(checkout_dir)
		_ = os.RemoveAll(checkout_dir)
	}()

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	_, status, err := commands.checkout(checkout_dir, repo, branch)
	switch status {
	case NotExist:
		tflog.Warn(ctx, fmt.Sprintf("Branch not found for update: %s", branch))
		d.SetId("")
		return nil
	case Exist:
		tflog.Info(ctx, fmt.Sprintf("Branch exists for update: %s", branch))
		break
	case Unknown:
		if err != nil {
			return diag.Errorf("failed to checkout branch %s: %s", branch, repo)
		}
	}

	is_clean := true
	var updated_files []string
	if d.HasChange("file") {
		files, _ := d.GetChange("file")

		for _, v := range files.(*schema.Set).List() {
			file := map_type.ToTypedObject(v.(map[string]interface{}))
			filepath := file["filepath"]

			if err := os.Remove(path.Join(checkout_dir, filepath)); err != nil {
				return diag.Errorf("failed to delete file %s: %s", filepath, err)
			}

			if _, err := gitCommand(checkout_dir, "add", "--", filepath); err != nil {
				return diag.Errorf("failed to rm file in git: %s", filepath)
			}

			updated_files = append(updated_files, fmt.Sprintf("- %s", filepath))
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
				updated_files = append(updated_files, fmt.Sprintf("+ %s", filepath))
			}
			updated_files = append(updated_files, fmt.Sprintf("? %s", filepath))
			continue
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
			updated_files = append(updated_files, fmt.Sprintf("~ %s", filepath))
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

	updated_files = set.GetSetFromStringArray(updated_files)
	a := d.Get("author")
	author := map_type.ToTypedObject(a.(map[string]interface{}))
	commit_message := author["message"]
	commit_body := fmt.Sprintf("The following files were updated by terraform:\n%s", strings.Join(updated_files, "\n"))
	commit_command := flatten("commit", "-m", commit_message, "-m", commit_body, "--allow-empty")
	commit_command = append(commit_command, commands.getAuthorString(author["name"], author["email"])...)
	commit_command = append(commit_command, "--")
	if _, err := gitCommand(checkout_dir, commit_command...); err != nil {
		return diag.Errorf("failed to commit file(s) to git: %s", err)
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

	checkout_dir := path.Join(os.TempDir(), unique.UniqueId())
	if err := os.MkdirAll(checkout_dir, 0755); err != nil {
		return diag.Errorf("failed to create git temp dir: %s", err)
	}
	lockCheckout(checkout_dir)
	defer func() {
		unlockCheckout(checkout_dir)
		_ = os.RemoveAll(checkout_dir)
	}()

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	_, status, err := commands.checkout(checkout_dir, repo, branch)
	switch status {
	case NotExist:
		tflog.Warn(ctx, fmt.Sprintf("Branch not found for create: %s", branch))
		return diag.Errorf("Branch not found for create %s: %s", branch, repo)
	case Exist:
		tflog.Info(ctx, fmt.Sprintf("Branch exists for update: %s", branch))
		break
	case Unknown:
		if err != nil {
			return diag.Errorf("failed to checkout branch %s: %s", branch, repo)
		}
	}

	var added_files []string
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
		added_files = append(added_files, filepath)
	}

	a := d.Get("author")
	author := map_type.ToTypedObject(a.(map[string]interface{}))
	commit_message := author["message"]
	commit_body := fmt.Sprintf("The following files were created by terraform:\n%s", strings.Join(added_files, "\n"))
	commit_command := flatten("commit", "-m", commit_message, "-m", commit_body, "--allow-empty")
	commit_command = append(commit_command, commands.getAuthorString(author["name"], author["email"])...)
	if _, err := gitCommand(checkout_dir, commit_command...); err != nil {
		return diag.Errorf("failed to commit file(s) to git: %s", err)
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

	checkout_dir := path.Join(os.TempDir(), unique.UniqueId())
	if err := os.MkdirAll(checkout_dir, 0755); err != nil {
		return diag.Errorf("failed to create git temp dir: %s", err)
	}
	lockCheckout(checkout_dir)
	defer func() {
		unlockCheckout(checkout_dir)
		_ = os.RemoveAll(checkout_dir)
	}()

	commands := NewGitCommands(meta.(*Owner).name, meta.(*Owner).token, org, hostname)

	rev, status, err := commands.checkout(checkout_dir, repo, branch)
	switch status {
	case Unknown:
		if err != nil {
			return diag.Errorf("failed to checkout branch %s: %s", branch, repo)
		}
	case Exist:
		log.Printf("[INFO] branch: %s (HEAD): %s", branch, rev)
	case NotExist:
		force := d.Get("force_new").(bool)
		tflog.Warn(ctx, fmt.Sprintf("failed to find remote branch: %s", branch))
		if force {
			// this will create the resource, ignores ignore_changes
			d.SetId("")
		} else {
			d.SetId("-1")
		}
		return nil
	}

	files := d.Get("file")
	clean_files := d.Get("file")
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
				clean_files.(*schema.Set).Remove(v)
				is_clean = false
			}
			is_clean = false
			log.Printf("[WARN] Expected file missing in branch: %s", filepath)
		}
		if string(out) != contents {
			log.Printf("[INFO] File contents changed: %s", filepath)
			is_clean = false
			clean_files.(*schema.Set).Remove(v)
		}
	}

	if !is_clean {
		d.Set("file", clean_files)
		return nil
	}

	current_id := d.Id()
	log.Printf("[INFO] Current branch revision: %s", current_id)
	log.Printf("[INFO] Actual branch revision: %s", rev)

	if current_id != rev {
		log.Printf("[INFO] Remote revision not the same as local revision: %s <-> %s", rev, current_id)
		d.SetId(rev)
		return nil
	}

	d.SetId(current_id)
	return nil
}
