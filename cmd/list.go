package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE:  runList,
}

func runList(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles found. Run 'mgit gen' to create one.")
		return nil
	}
	fmt.Printf("%-12s  %-16s  %-28s  %s\n", "KEY", "LABEL", "EMAIL", "SSH KEY")
	fmt.Printf("%-12s  %-16s  %-28s  %s\n", "---", "-----", "-----", "-------")
	for _, p := range cfg.Profiles {
		fmt.Printf("%-12s  %-16s  %-28s  %s\n", p.Key, p.Label, p.Email, p.SSHKey)
	}
	return nil
}
