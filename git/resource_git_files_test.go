package git

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestGitFileResource(t *testing.T) {

	testReleaseRepository := os.Getenv("GITHUB_TEMPLATE_REPOSITORY")
	testReleaseOwner := testOrganizationFunc()

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
		`, testReleaseRepository, testReleaseOwner)

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

	t.Run("queries latest release", func(t *testing.T) {

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
		`, testReleaseRepository, testReleaseOwner)

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

	/*t.Run("queries release by ID or tag", func(t *testing.T) {

		config := fmt.Sprintf(`
			data "gitfile_release" "by_id" {
				repository = "%[1]s"
				owner = "%[2]s"
				retrieve_by = "id"
				release_id = "%[3]s"
			}

			data "gitfile_release" "by_tag" {
				repository = "%[1]s"
				owner = "%[2]s"
				retrieve_by = "tag"
				release_tag = data.gitfile_release.by_id.release_tag
			}
		`, testReleaseRepository, testReleaseOwner, testReleaseID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.gitfile_release.by_id", "id", testReleaseID,
			),
			resource.TestCheckResourceAttr(
				"data.gitfile_release.by_tag", "id", testReleaseID,
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					skipUnlessMode(t, mode)
					testAccPreCheck(t)
				},
				Providers: testAccProviders,
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

	t.Run("errors when querying with non-existent ID", func(t *testing.T) {

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
