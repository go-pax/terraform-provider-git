variable "gh_token" {
  type = string
}

terraform {
  required_providers {
    git = {
      #      versions = ["0.1"]
      #      source   = "github.com/go-pax/git"
      version = "~> 0.1"
      source  = "go-pax/git"
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

  branches = {
    ignore_changes_1 = {
      "file" = {
        contents = "\n\thello world.\n"
      }
    }
  }
}

resource "random_string" "test" {
  for_each = local.branches
  length   = 10
  special  = false
  lower    = true
}

resource "github_branch" "test" {
  lifecycle {
    ignore_changes = all
  }

  for_each   = local.branches
  repository = local.repo
  branch     = format("%s-%s", each.key, random_string.test[each.key].result)
}

resource "git_files" "test" {
  lifecycle {
    ignore_changes = [file]
  }
  for_each     = local.branches
  hostname     = "github.com"
  repository   = local.repo
  organization = local.org
  branch       = github_branch.test[each.key].branch
  force_new    = false
  author = {
    name    = "trentmillar"
    email   = "1146672+trentmillar@users.noreply.github.com"
    message = "chore: terraform lifecycle management automated commit"
  }
  dynamic "file" {
    for_each = each.value
    content {
      contents = file.value.contents
      filepath = file.key
    }
  }
}
