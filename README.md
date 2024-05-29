# Git Terraform Provider

This document provides an overview of the Git Terraform Provider.

## Description

Git Terraform Provider allows managing files over various Git services with Terraform. It includes different types of resources such as creating files and managing existing ones.

## Core Components:

#### Provider Block

The provider blocks indicate the configuration for the Github, Azure DevOps, Bitbucket, or any git provider token, the repository owner, and project.

##### Example use:

```terraform
provider "git" {
  owner = "<owner>"
  token = "<your_github_token>"
}
```
Replace <owner> and <your_github_token> with actual values.

### Resource "git_files"

It represents the files in a designated repository.
Example use:

```terraform
resource "git_files" "test" {
  hostname   = "github.com"
  repository = "repository_name"
  organization = "organization_name"
  branch     = "branch_name"
  author = {
    name  = "author_name"
    email = "author_email"
    message = "author_commit_message"
  }
  file {
    contents = "--\nname: some manifest\n"
    filepath = "file.yml"
  }
  file {
    contents = "<html>...</html>"
    filepath = "path/to/file.htm"
  }  
}
```

Replace placeholder values with actual repository, organization, branch, author details, and file content.

## Tests

The git_files resource offers unit tests to validate:

- Fields are properly set especially the id.
- Addition of multiple files with different auth modes i.e., for anonymous, individual, and organization accounts.
- Committed changes for files with newly created branches via GitHub.
- (Disabled tests for release-related cases with defined expected error outcomes).

## Environment Variables

The following environment variables are used in the test cases:

```
# for testRepository
export GITHUB_TEMPLATE_REPOSITORY=<your_repository>
export GIT_TOKEN=<your_pat>
```

Replace <your_repository> with your Github repository.

## Note

Please replace placeholders in the examples with your actual values before use. The unit tests are written to run locally are therefore may require adjustments based on your CI/CD platform and configuration.
