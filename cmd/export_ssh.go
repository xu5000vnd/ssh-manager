package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var exportSSHPath string

var exportSSHCmd = &cobra.Command{
	Use:   "export-ssh-config",
	Short: "Export as plaintext OpenSSH config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if exportSSHPath == "" {
			return errors.New("-o/--output is required")
		}

		cfg, _, _, err := loadConfigWithPassword()
		if err != nil {
			return err
		}
		content := config.GenerateSSHConfig(cfg.Connections)
		if err := os.WriteFile(exportSSHPath, []byte(content), 0o600); err != nil {
			return err
		}
		fmt.Printf("Exported SSH config to %s\n", exportSSHPath)
		fmt.Println("Warning: exported file is plaintext")
		return nil
	},
}

func init() {
	exportSSHCmd.Flags().StringVarP(&exportSSHPath, "output", "o", "", "output file path")
}
