terraform {
  required_providers {
    git = {
      version = "~> 0.1"
      source  = "go-pax/git"
    }
  }
}

provider "git" {
  owner = "target-github-org-name"
  token = "ghp_1234567890"
}
