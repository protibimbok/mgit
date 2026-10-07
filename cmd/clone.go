package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/deps"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:                "clone <key>:<user>/<repo> [git-args...]",
	Short:              "Clone a repo using a profile key, or pick a profile for an HTTPS URL",
	Args:               cobra.MinimumNArgs(1),
	RunE:               runClone,
	DisableFlagParsing: true,
}

func runClone(cmd *cobra.Command, args []string) error {
	if args[0] == "-h" || args[0] == "--help" {
		return cmd.Help()
	}
	target := args[0]
	rest := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	sshURL, profile, err := resolveRemoteURL(target, cfg, "Select profile")
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("invalid format: use <key>:<user>/<repo> or a GitHub HTTPS URL")
	}

	if err := deps.RequireGit(); err != nil {
		return err
	}

	c := exec.Command("git", append([]string{"clone", sshURL}, rest...)...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}

	dir := cloneTargetDir(sshURL, rest)
	if dir == "" {
		return nil
	}
	if err := applyProfileToRepo(dir, profile); err != nil {
		fmt.Printf("warning: could not set git user in %s: %v\n", dir, err)
		return nil
	}
	fmt.Printf("Git user in %s set to %s <%s>\n", dir, profile.Name, profile.Email)
	return nil
}

var cloneValueOpts = map[string]bool{
	"--template": true, "-o": true, "--origin": true, "-b": true, "--branch": true,
	"-u": true, "--upload-pack": true, "--reference": true, "--reference-if-able": true,
	"--separate-git-dir": true, "--depth": true, "--shallow-since": true,
	"--shallow-exclude": true, "-c": true, "--config": true, "--filter": true,
	"-j": true, "--jobs": true, "--server-option": true, "--bundle-uri": true,
	"--ref-format": true,
}

func clonePositionalDir(rest []string) string {
	dir := ""
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		switch {
		case a == "--":
			if i+1 < len(rest) {
				dir = rest[len(rest)-1]
			}
			return dir
		case strings.HasPrefix(a, "-"):
			if cloneValueOpts[a] {
				i++ // skip the option's value
			}
		default:
			dir = a
		}
	}
	return dir
}

func cloneTargetDir(sshURL string, rest []string) string {
	dir := clonePositionalDir(rest)
	if dir == "" {
		repo := sshURL[strings.LastIndex(sshURL, ":")+1:]
		dir = strings.TrimSuffix(path.Base(repo), ".git")
	}
	if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
		return dir
	}
	return ""
}
