package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/deps"
)

var gitInitCmd = &cobra.Command{
	Use:   "init [key]",
	Short: "Init git repo (if needed) and set user config from a profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGitInit,
}

func runGitInit(_ *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles found — run 'mgit new' to create one")
	}

	if err := deps.RequireGit(); err != nil {
		return err
	}

	// An explicit key always wins; otherwise reuse what the repo already
	// has configured (mgit.profile, remotes, user.email) before asking.
	var profile *config.Profile
	if len(args) == 1 {
		if profile = cfg.FindByKey(args[0]); profile == nil {
			return fmt.Errorf("unknown profile key %q — run 'mgit list' to see available profiles", args[0])
		}
	} else if profile, err = chooseProfile(cfg, "Choose a profile"); err != nil {
		return err
	}

	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		c := exec.Command("git", "init")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
	}

	if err := applyProfileToRepo("", profile); err != nil {
		return err
	}
	fmt.Printf("Git user set to %s <%s>\n", profile.Name, profile.Email)
	return nil
}
