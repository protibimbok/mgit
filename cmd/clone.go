package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/prompt"
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

	var sshURL string

	switch {
	case strings.HasPrefix(target, "https://github.com/") || strings.HasPrefix(target, "http://github.com/"):
		path := strings.TrimPrefix(strings.TrimPrefix(target, "https://"), "http://")
		path = strings.TrimPrefix(path, "github.com/")
		path = strings.TrimSuffix(path, ".git")

		if len(cfg.Profiles) == 0 {
			return fmt.Errorf("no profiles found — run 'mgit gen' to create one")
		}
		idx, err := prompt.Select("Select profile", profileLabels(cfg.Profiles))
		if err != nil {
			return err
		}
		sshURL = fmt.Sprintf("git@hub.%s:%s", cfg.Profiles[idx].Key, path)

	case strings.Contains(target, ":"):
		parts := strings.SplitN(target, ":", 2)
		key, path := parts[0], parts[1]
		if cfg.FindByKey(key) == nil {
			return fmt.Errorf("unknown profile key %q — run 'mgit list' to see available profiles", key)
		}
		path = strings.TrimSuffix(path, ".git")
		sshURL = fmt.Sprintf("git@hub.%s:%s", key, path)

	default:
		return fmt.Errorf("invalid format: use <key>:<user>/<repo> or a GitHub HTTPS URL")
	}

	c := exec.Command("git", append([]string{"clone", sshURL}, rest...)...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
