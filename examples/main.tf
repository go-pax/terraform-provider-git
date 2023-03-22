terraform {
  required_providers {
    git = {
      versions = ["0.1"]
      source = "github.com/go-pax/git"
    }
  }
}

provider git {
  owner = "test-dump"
  token = ""
}