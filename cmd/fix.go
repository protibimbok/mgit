package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/prompt"
)

var fixCmd = &cobra.Command{
	Use:   "fix [remote]",
	Short: "Rewrite a GitHub HTTPS remote to SSH using a profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFix,
}

var httpsPattern = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)

func runFix(_ *cobra.Command, args []string) error {
	remote := "origin"
	if len(args) == 1 {
		remote = args[0]
	}

	out, err := exec.Command("git", "remote", "get-url", remote).Output()
	if err != nil {
		return fmt.Errorf("remote %q not found (are you inside a git repo?)", remote)
	}
	remoteURL := strings.TrimSpace(string(out))

	m := httpsPattern.FindStringSubmatch(remoteURL)
	if m == nil {
		return fmt.Errorf("remote %q is not a GitHub HTTPS URL:\n  %s\nNothing to fix.", remote, remoteURL)
	}
	owner, repo := m[1], m[2]

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles found — run 'mgit gen' to create one")
	}

	idx, err := prompt.Select("Select profile for SSH remote", profileLabels(cfg.Profiles))
	if err != nil {
		return err
	}
	key := cfg.Profiles[idx].Key

	sshURL := fmt.Sprintf("git@hub.%s:%s/%s", key, owner, repo)
	if err := exec.Command("git", "remote", "set-url", remote, sshURL).Run(); err != nil {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	fmt.Printf("Remote %q updated:\n  %s\n→ %s\n", remote, remoteURL, sshURL)
	return nil
}
