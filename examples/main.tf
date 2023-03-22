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

resource "random_string" "test" {
  length  = 10
  special = false
  lower   = true
}

locals {
  org    = "test-dump"
  repo   = "test-git-provider"
  branch = format("branch-%s", random_string.test.result)

  files = {
    01 = {
      "src/main.hpp" = {
        contents = "#include <vector>\n#include <cstring>\n"
      }
      "src/main.cpp" = {
        contents = "#include \"main.hpp\"\n\nint main(int argc, char *argv[])\n{\treturn 0;}\n"
      }
    }
  }
}

resource "github_branch" "test" {
  repository = local.repo
  branch     = local.branch
}

resource "git_files" "test" {
  hostname     = "github.com"
  repository   = local.repo
  organization = local.org
  branch       = github_branch.test.branch
  author = {
    name    = "trentmillar"
    email   = "1146672+trentmillar@users.noreply.github.com"
    message = "chore: terraform lifecycle management automated commit"
  }
  dynamic "file" {
    for_each = local.files.01[each.key]
    content {
      content  = file.value.content
      filepath = file.value.path
    }
  }
  file {
    contents = "hello world."
    filepath = "files/1.txt"
  }
  file {
    contents = "hello world.\n\t"
    filepath = "files/2"
  }
  file {
    contents = "#include <vector>\n#include <cstring>\n"
    filepath = "includes/code.h"
  }
}
