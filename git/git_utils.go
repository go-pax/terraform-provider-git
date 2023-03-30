package git

import (
	"fmt"
	"github.com/go-pax/terraform-provider-git/utils/mutexkv"
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
		return out, fmt.Errorf("Error while running git %[1]s: %[4]w\nWorking dir: %[2]s\nOutput: %[3]s", strings.Join(args, " "), cwd, string(out), err)
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

var gitfileMutexKV = mutexkv.NewMutexKV()

func lockCheckout(checkout_dir string) {
	gitfileMutexKV.Lock(checkout_dir)
}

func unlockCheckout(checkout_dir string) {
	gitfileMutexKV.Unlock(checkout_dir)
}
