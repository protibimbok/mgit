//go:build !windows

package deps

import "os/exec"

func lookPath(name string) (string, error) {
	return exec.LookPath(name)
}
