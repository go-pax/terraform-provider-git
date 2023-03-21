package git

import (
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/hashcode"
	"github.com/go-pax/terraform-provider-git/utils/mutexkv"
	"github.com/hashicorp/errwrap"
	"os/exec"
	"strings"
)

func gitCommand(cwd string, args ...string) ([]byte, error) {
	command := exec.Command("git", args...)
	if cwd != "" {
		command.Dir = cwd
	}
	out, err := command.CombinedOutput()
	if err != nil {
		return out, errwrap.Wrapf(fmt.Sprintf("Error while running git %s: {{err}}\nWorking dir: %s\nOutput: %s", strings.Join(args, " "), cwd, string(out)), err)
	} else {
		return out, err
	}
}

func flatten(args ...interface{}) []string {
	ret := make([]string, 0, len(args))

	for _, arg := range args {
		switch arg := arg.(type) {
		default:
			panic("can only handle strings and []strings")
		case string:
			ret = append(ret, arg)
		case []string:
			ret = append(ret, arg...)
		}
	}

	return ret
}

func hashString(v interface{}) int {
	switch v := v.(type) {
	default:
		panic(fmt.Sprintf("unexpectedtype %T", v))
	case string:
		return hashcode.HashcodeString(v)
	}
}

// This is a global MutexKV for use within this plugin.
var gitfileMutexKV = mutexkv.NewMutexKV()

func lockCheckout(checkout_dir string) {
	gitfileMutexKV.Lock(checkout_dir)
}

func unlockCheckout(checkout_dir string) {
	gitfileMutexKV.Unlock(checkout_dir)
}
