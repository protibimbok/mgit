package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/deps"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <key>:<user>/<repo> [git-args...]",
	Short: "Clone a repo using a profile key, or pick a profile for an HTTPS URL",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runClone,
}

func runClone(_ *cobra.Command, args []string) error {
	target := args[0]
	rest := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	sshURL, transformed, err := resolveRemoteURL(target, cfg, "Select profile")
	if err != nil {
		return err
	}
	if !transformed {
		return fmt.Errorf("invalid format: use <key>:<user>/<repo> or a GitHub HTTPS URL")
	}

	if err := deps.RequireGit(); err != nil {
		return err
	}

	c := exec.Command("git", append([]string{"clone", sshURL}, rest...)...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
