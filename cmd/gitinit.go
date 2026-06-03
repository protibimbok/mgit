package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/prompt"
)

var gitInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Init git repo (if needed) and set user config from a profile",
	RunE:  runGitInit,
}

func runGitInit(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles found — run 'mgit gen' to create one")
	}

	labels := profileLabels(cfg.Profiles)
	idx, err := prompt.Select("Choose a profile", labels)
	if err != nil {
		return err
	}
	profile := cfg.Profiles[idx]

	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		c := exec.Command("git", "init")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
	}

	for _, kv := range [][2]string{{"user.name", profile.Name}, {"user.email", profile.Email}} {
		c := exec.Command("git", "config", kv[0], kv[1])
		if err := c.Run(); err != nil {
			return err
		}
	}
	fmt.Printf("Git user set to %s <%s>\n", profile.Name, profile.Email)
	return nil
}
