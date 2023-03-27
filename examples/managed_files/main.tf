variable "gh_token" {
  type = string
}

terraform {
  required_providers {
    git = {
      versions = ["0.1"]
      source   = "github.com/go-pax/git"
    }
  }
}

provider "git" {
  owner = "test-dump"
  token = var.gh_token
}

provider "github" {
  owner = "test-dump"
  token = var.gh_token
}

locals {
  org  = "test-dump"
  repo = "test-git-provider"

  unmanaged = {
    "src/main.hpp" = {
      contents = "#include <vector>\n#include <cstring>\n"
    }
    "src/main.cpp" = {
      contents = "#include \"main.hpp\"\n\nint main(int argc, char *argv[])\n{\n\treturn 0;\n}\n"
    }
  }

}

resource "github_branch" "unmanaged" {
  repository = local.repo
  branch     = "unmanaged"
}

resource "git_files" "unmanaged" {
  lifecycle {
    ignore_changes = all
  }
  depends_on = [
    github_branch.unmanaged

  ]
  hostname     = "github.com"
  repository   = local.repo
  organization = local.org
  branch       = "unmanaged"
  author = {
    name    = "trentmillar"
    email   = "1146672+trentmillar@users.noreply.github.com"
    message = "chore: terraform lifecycle management automated commit"
  }
  file {
    contents = "hello world."
    filepath = "files/hello.txt"
  }
}

resource "git_files" "managed" {
  hostname     = "github.com"
  repository   = local.repo
  organization = local.org
  branch       = "main"
  author = {
    name    = "trentmillar"
    email   = "1146672+trentmillar@users.noreply.github.com"
    message = "chore: terraform lifecycle management automated commit"
  }
  file {
    contents = "managed hello world."
    filepath = "cant_touch_this.txt"
  }

}
