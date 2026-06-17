package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/protibimbok/mgit/cmd"
	"github.com/protibimbok/mgit/internal/deps"
)

// set via -ldflags "-X main.version=x.y.z"
var version = "dev"

var mgitCommands = map[string]bool{
	"gen":     true,
	"init":    true,
	"clone":   true,
	"del":     true,
	"list":    true,
	"fix":     true,
	"help":    true,
	"version": true,
}

func main() {
	if len(os.Args) > 1 {
		first := os.Args[1]
		if first != "--help" && first != "-h" && first != "--version" && !mgitCommands[first] {
			gitArgs := os.Args[1:]
			if newArgs, ok, err := cmd.TryInterceptGitArgs(gitArgs); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			} else if ok {
				gitArgs = newArgs
			}

			if err := deps.RequireGit(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			c := exec.Command("git", gitArgs...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					os.Exit(e.ExitCode())
				}
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}
	cmd.Execute(version)
}
