## GitHub example usage

variable "gh_token" {
  type = string
}

terraform {
  required_providers {
    git = {
      version = "~> 0.1"
      source  = "go-pax/git"
    }
  }
}

provider "git" {
  owner = "test-dump"
  token = var.gh_token
}

resource "git_files" "test" {
  hostname     = "github.com"
  repository   = "test-git-provider"
  organization = "test-dump"
  branch       = "branch_1"
  author = {
    name    = "username"
    email   = "1146672+username@users.noreply.github.com"
    message = "chore: terraform lifecycle management automated commit"
  }
  file {
    contents = "#include <vector>\n#include <cstring>\n"
    filepath = "src/main.hpp"
  }
  file {
    contents = "#include \"main.hpp\"\n\nint main(int argc, char *argv[])\n{\n\treturn 0;\n}\n"
    filepath = "src/main.cpp"
  }
}


## Bitbucket Server example usage

# This provider whilst it is targetted at Github as it stands you *can* use it with Bitbucket server, see below example to make this work. The key is to replace the org with the relevant path as below:
# full_repo_arg := fmt.Sprintf("https://%s:%s@%s/%s/%s", r.user, r.token, r.hostname, r.organization, repo)

# Since it builds it up based off the user / token / hostname and org we can manipulate this and change org to be the matching URL as per bitbucket or other git
# https://<BITBUCKET_SERVER_URL>/scm/<PROJECT>/<REPO>.git


provider "git" {
  owner = "<BITBUCKET_USERNAME>"
  token = "<BITBUCKET_HTTPS_ACCESS_TOKEN>"
}

terraform {
  required_providers {
    git = {
      version = "~> 0.1"
      source  = "go-pax/git"
    }
  }
}

# https://<BITBUCKET_SERVER_URL>/scm/<PROJECT>/<REPO>.git
resource "aws_codecommit_repository" "this" {
  repository_name = "repo_name"
  description     = "repo_name repository"
  default_branch  = "main"
}

resource "git_files" "this" {
  hostname     = "<BITBUCKET_SERVER_URL>"
  repository   = "<REPO>"
  organization = "scm/<PROJECT>"
  branch       = "main"
  author = {
    name    = "example"
    email   = "example@example.com"
    message = "automated commit"
  }
  file {
    contents = "managed file"
    filepath = "managed_file.txt"
  }
}
