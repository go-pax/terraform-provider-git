package git

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccGitFileResource tests the behavior of the git_files resource when creating files in different scenarios.
// The function runs a series of sub-tests to cover the different test cases.
func TestAccGitFileResource(t *testing.T) {

	testRepository := os.Getenv("GITHUB_TEMPLATE_REPOSITORY")
	testOwner := testOrganizationFunc()
	testHelper := NewTestHelper()

	/*t.Run("testing git_files resource in Azure DevOps w/ one file", func(t *testing.T) {

		config := fmt.Sprintf(`
			provider "git" {
				owner = "%[1]s"
				token = "%[2]s"
			}

			resource "git_files" "test" {
				hostname = "dev.azure.com"
				project = "%[3]s"
				organization = "%[1]s"
				branch = "%[4]s-branch"
				repository = "%[4]s"
				author = {
					name = "trentmillar"
					email = "1146672+trentmillar@users.noreply.github.com"
					message = "chore: terraform lifecycle management automated commit"
				}
				file {
					contents = "hello world."
					filepath = "files/go/here/helloworld.txt"
				}
			}
		`, "YOUR ORG IN AZDO", "PAT", "PROJECT", "REPO NAME")

		resource.UnitTest(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							rs := s.RootModule().Resources["git_files.test"]
							att := rs.Primary.Attributes["id"]
							if att == "" {
								return fmt.Errorf("expected 'id' to have a value")
							}
							return nil
						},
					),
				},
			},
		})
	})*/

	t.Run("testing git_files resource w/ one file", func(t *testing.T) {

		config := fmt.Sprintf(`
			provider "git" {}

			resource "git_files" "test" {
				hostname = "github.com"
				repository = "%s"
				organization = "%s"
				branch = "main-patch"
				author = {
					name = "trentmillar"
					email = "1146672+trentmillar@users.noreply.github.com"
					message = "chore: terraform lifecycle management automated commit"
				}
				file {
					contents = "hello world."
					filepath = "files/go/here/helloworld.txt"
				}
			}
		`, testRepository, testOwner)

		resource.UnitTest(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							rs := s.RootModule().Resources["git_files.test"]
							att := rs.Primary.Attributes["id"]
							if att == "" {
								return fmt.Errorf("expected 'id' to have a value")
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("add multiple files w/ different auth schemes", func(t *testing.T) {

		config := fmt.Sprintf(`
			provider "git" {}
			resource "git_files" "test" {
				hostname = "github.com"
				repository = "%s"
				organization = "%s"
				branch = "main-patch-2"
				author = {
					name = "trentmillar"
					email = "1146672+trentmillar@users.noreply.github.com"
					message = "chore: terraform lifecycle management automated commit"
				}
				file {
					contents = "hello world."
					filepath = "files/test/2.txt"
				}
				file {
					contents = "hello world.\n\t"
					filepath = "files/test/3.txt"
				}
			}
		`, testRepository, testOwner)

		check := resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				rs := s.RootModule().Resources["git_files.test"]
				att := rs.Primary.Attributes["id"]
				if att == "" {
					return fmt.Errorf("expected 'id' to have a value")
				}
				return nil
			},
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: providerFactories,
				PreCheck: func() {
					skipUnlessMode(t, mode)
					testAccPreCheck(t)
				},
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			testCase(t, anonymous)
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("add files w/ github created branch", func(t *testing.T) {

		_ = testHelper.GenerateBranchName()

		config := fmt.Sprintf(`
			resource "git_files" "test" {
				lifecycle { ignore_changes = all }
				hostname = "github.com"
				repository = "%[2]s"
				organization = "%[1]s"
				branch = "multi-files"
				author = {
					name = "trentmillar"
					email = "1146672+trentmillar@users.noreply.github.com"
					message = "chore: terraform lifecycle management automated commit"
				}
				file {
					contents = "hello world."
					filepath = "files/test/2.txt"
				}
				file {
					contents = "hello world.\n\t"
					filepath = "files/test/3.txt"
				}
			}
		`, testOwner, testRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"git_files.test", "organization", testOwner,
			),
			resource.TestCheckResourceAttr(
				"git_files.test", "repository", testRepository,
			),
			resource.TestCheckResourceAttr(
				"git_files.test", "branch", "multi-files",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					skipUnlessMode(t, mode)
					testAccPreCheck(t)
				},
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			testCase(t, anonymous)
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	/*t.Run("errors when querying with non-existent ID", func(t *testing.T) {

		config := `
			data "gitfile_release" "test" {
				repository = "test"
				owner = "test"
				retrieve_by = "id"
			}
		`

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					skipUnlessMode(t, mode)
					testAccPreCheck(t)
				},
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile("`release_id` must be set when `retrieve_by` = `id`"),
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			testCase(t, anonymous)
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("errors when querying with non-existent repository", func(t *testing.T) {

		config := `
			data "gitfile_release" "test" {
				repository = "test"
				owner = "test"
				retrieve_by = "latest"
			}
		`
		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					skipUnlessMode(t, mode)
					testAccPreCheck(t)
				},
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile(`Not Found`),
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			testCase(t, anonymous)
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("errors when querying with non-existent tag", func(t *testing.T) {

		config := `
			data "gitfile_release" "test" {
				repository = "test"
				owner = "test"
				retrieve_by = "tag"
			}
		`
		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					skipUnlessMode(t, mode)
					testAccPreCheck(t)
				},
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile("`release_tag` must be set when `retrieve_by` = `tag`"),
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			testCase(t, anonymous)
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})*/
}
