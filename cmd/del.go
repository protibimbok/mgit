package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/prompt"
	"github.com/protibimbok/mgit/internal/sshutil"
)

var delCmd = &cobra.Command{
	Use:   "del [key]",
	Short: "Remove a profile and its SSH key",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDel,
}

func runDel(_ *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles found")
	}

	var key string
	if len(args) == 1 {
		key = args[0]
	} else {
		idx, err := prompt.Select("Select profile to delete", profileLabels(cfg.Profiles))
		if err != nil {
			return err
		}
		key = cfg.Profiles[idx].Key
	}

	profile := cfg.FindByKey(key)
	if profile == nil {
		return fmt.Errorf("profile %q not found", key)
	}

	confirmed, err := prompt.Confirm(fmt.Sprintf("Delete profile %q (%s <%s>)?", key, profile.Name, profile.Email))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Cancelled.")
		return nil
	}

	for _, path := range []string{profile.SSHKey, profile.SSHKey + ".pub"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("warning: could not remove %s: %v\n", path, err)
		}
	}
	if err := sshutil.RemoveFromSSHConfig(key); err != nil {
		fmt.Printf("warning: could not update ~/.ssh/config: %v\n", err)
	}

	cfg.Remove(key)
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Profile %q removed.\n", key)
	return nil
}
