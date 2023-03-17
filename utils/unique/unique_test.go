package unique

import (
	"strings"
	"testing"
)

func TestUnique(t *testing.T) {
	iterations := 10000
	uniques := make(map[string]struct{})
	var name string
	for i := 0; i < iterations; i++ {
		name = UniqueId()

		if _, ok := uniques[name]; ok {
			t.Fatalf("Got duplicated id! %s", name)
		}

		if !strings.HasPrefix(name, "git_") {
			t.Fatalf("Unique ID didn't have git_ prefix! %s", name)
		}

		uniques[name] = struct{}{}
	}
}
