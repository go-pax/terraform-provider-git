package git

import (
	"os"
	"testing"
)

var testOrganization string = testOrganizationFunc()
var testToken string = os.Getenv("GITHUB_TOKEN")

const anonymous = "anonymous"
const individual = "individual"
const organization = "organization"

func testOrganizationFunc() string {
	organization := os.Getenv("GITHUB_ORGANIZATION")
	if organization == "" {
		organization = os.Getenv("GITHUB_TEST_ORGANIZATION")
	}
	return organization
}

func skipUnlessMode(t *testing.T, providerMode string) {
	switch providerMode {
	case anonymous:
		if os.Getenv("GITHUB_BASE_URL") != "" &&
			os.Getenv("GITHUB_BASE_URL") != "https://api.github.com/" {
			t.Log("anonymous mode not supported for GHES deployments")
			break
		}

		if os.Getenv("GITHUB_TOKEN") == "" {
			return
		} else {
			t.Log("GITHUB_TOKEN environment variable should be empty")
		}
	case individual:
		if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("GITHUB_OWNER") != "" {
			return
		} else {
			t.Log("GITHUB_TOKEN and GITHUB_OWNER environment variables should be set")
		}
	case organization:
		if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("GITHUB_ORGANIZATION") != "" {
			return
		} else {
			t.Log("GITHUB_TOKEN and GITHUB_ORGANIZATION environment variables should be set")
		}
	}

	t.Skipf("Skipping %s which requires %s mode", t.Name(), providerMode)
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GITHUB_TOKEN"); v == "" {
		t.Fatal("GITHUB_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_ORGANIZATION"); v == "" && os.Getenv("GITHUB_OWNER") == "" {
		t.Fatal("GITHUB_ORGANIZATION or GITHUB_OWNER must be set for acceptance tests")
	}
	//if v := os.Getenv("GITHUB_TEST_USER"); v == "" {
	//	t.Fatal("GITHUB_TEST_USER must be set for acceptance tests")
	//}
	//if v := os.Getenv("GITHUB_TEST_COLLABORATOR"); v == "" {
	//	t.Fatal("GITHUB_TEST_COLLABORATOR must be set for acceptance tests")
	//}
	if v := os.Getenv("GITHUB_TEMPLATE_REPOSITORY"); v == "" {
		t.Fatal("GITHUB_TEMPLATE_REPOSITORY must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ID"); v == "" {
		t.Fatal("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ID must be set for acceptance tests")
	}
}
