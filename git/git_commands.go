package git

import (
	"fmt"
	"os"
	"strings"
)

type GitCommands struct {
	user         string
	token        string
	organization string
	hostname     string
}

func NewGitCommands(user string, token string, org string, hostname string) *GitCommands {
	return &GitCommands{
		user:         user,
		token:        token,
		organization: org,
		hostname:     hostname,
	}
}

func (r *GitCommands) configureAuthor(name string, email string) error {
	if _, err := gitCommand("", "config", "--global", "user.name", fmt.Sprintf("\"%s\"", name)); err != nil {
		return err
	}
	if _, err := gitCommand("", "config", "--global", "user.email", fmt.Sprintf("\"%s\"", email)); err != nil {
		return err
	}
	return nil
}

func (r *GitCommands) checkout(path string, repo string, branch string) (string, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	// May already be checked out from another project
	if _, err := os.Stat(fmt.Sprintf("%s/.git", path)); err != nil {
		full_repo_arg := fmt.Sprintf("https://%s:%s@%s/%s/%s", r.user, r.token, r.hostname, r.organization, repo)
		if _, err := gitCommand(path, "clone", "-b", branch, "--", full_repo_arg, "."); err != nil {
			return "", err
		}
	}
	var head string
	if out, err := gitCommand(path, "rev-parse", "HEAD"); err != nil {
		return "", err
	} else {
		head = strings.TrimRight(string(out), "\n")
	}

	return head, nil
}
