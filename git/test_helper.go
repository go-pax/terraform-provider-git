package git

import (
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/unique"
	"os"
)

type TestHelper struct{}

type TestFile struct {
	Path string
}

func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

func (t TestHelper) GenerateFile(content string) (*TestFile, error) {
	fileName := unique.PrefixedUniqueId("test_file_")
	file, err := os.CreateTemp("", fileName)
	if err != nil {
		return nil, fmt.Errorf("creating test file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return nil, err
	}
	return &TestFile{
		Path: file.Name(),
	}, nil
}

func (t TestHelper) GenerateBranchName() string {
	return unique.PrefixedUniqueId("branch_")
}
