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
  org    = "test-dump"
  repo   = "test-git-provider"

  branches = {
    branch_1 = {
      "src/main.hpp" = {
        contents = "#include <vector>\n#include <cstring>\n"
      }
      "src/main.cpp" = {
        contents = "#include \"main.hpp\"\n\nint main(int argc, char *argv[])\n{\n\treturn 0;\n}\n"
      }
    }
    branch_2 = {
      "src/main.hpp" = {
        contents = "#include <vector>\n#include <cstring>\n"
      }
      "src/main.cpp" = {
        contents = "#include \"main.hpp\"\n\nint main(int argc, char *argv[])\n{\n\treturn 0;\n}\n"
      }
    }
  }
}

resource "random_string" "test" {
  for_each = local.branches
  length  = 10
  special = false
  lower   = true
}

resource "github_branch" "test" {
  for_each = local.branches
  repository = local.repo
  branch     = format("%s-%s", each.key, random_string.test[each.key].result)
}

resource "git_files" "test" {
  for_each = local.branches
  hostname     = "github.com"
  repository   = local.repo
  organization = local.org
  branch       = github_branch.test[each.key].branch
  author = {
    name    = "trentmillar"
    email   = "1146672+trentmillar@users.noreply.github.com"
    message = "chore: terraform lifecycle management automated commit"
  }
  dynamic "file" {
    for_each = each.value
    content {
      contents  = file.value.contents
      filepath = file.key
    }
  }
}
