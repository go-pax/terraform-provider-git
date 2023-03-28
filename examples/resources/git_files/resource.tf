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
