package git

import (
	"fmt"
	"os"
	"strings"
)

type BranchStatus uint8

const (
	NotExist BranchStatus = 0
	Exist    BranchStatus = 1
	Unknown  BranchStatus = 2
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

func (r *GitCommands) getAuthorString(name string, email string) []string {
	return []string{"--author", fmt.Sprintf("%s <%s>", name, email)}
}

func (r *GitCommands) checkout(path string, repo string, branch string, project string) (string, BranchStatus, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", Unknown, err
	}

	// May already be checked out from another project
	if _, err := os.Stat(fmt.Sprintf("%s/.git", path)); err != nil {
		full_repo_arg := ""
		if project != "" {
			full_repo_arg = fmt.Sprintf("https://%s:%s@%s/%s/%s/_git/%s", r.user, r.token, r.hostname, r.organization, project, repo)
		} else {
			full_repo_arg = fmt.Sprintf("https://%s:%s@%s/%s/%s", r.user, r.token, r.hostname, r.organization, repo)
		}
		if _, err := gitCommand(path, "clone", "--", full_repo_arg, "."); err != nil {
			return "", Unknown, err
		}
	}

	if _, err := gitCommand(path, "checkout", "--guess", branch, "--"); err != nil {
		return "", NotExist, err
	}

	var head string
	if out, err := gitCommand(path, "rev-parse", "HEAD"); err != nil {
		return "", NotExist, err
	} else {
		head = strings.TrimRight(string(out), "\n")
	}

	return head, Exist, nil
}
