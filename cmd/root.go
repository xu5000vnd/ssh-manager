package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ssh-manager/internal/app"
)

var showVersion bool

var rootCmd = &cobra.Command{
	Use:   "ssh-manager",
	Short: "Encrypted SSH connection manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(Version)
			return nil
		}
		return app.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "print version")
	rootCmd.Version = Version
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(importSSHCmd)
	rootCmd.AddCommand(exportSSHCmd)
	rootCmd.AddCommand(passwordCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(upgradeCmd)
}
