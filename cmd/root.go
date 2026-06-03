package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mgit",
	Short: "Manage multiple GitHub accounts via SSH",
	Long: `mgit manages multiple GitHub SSH profiles and wraps git.

Unknown subcommands are forwarded directly to git.`,
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(gitInitCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(delCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(fixCmd)
}
