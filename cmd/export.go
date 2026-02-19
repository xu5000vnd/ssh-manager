package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var exportPath string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export connections to plaintext JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		if exportPath == "" {
			return errors.New("-o/--output is required")
		}
		cfg, _, _, err := loadConfigWithPassword()
		if err != nil {
			return err
		}
		if err := config.ExportJSON(cfg, exportPath); err != nil {
			return err
		}
		fmt.Printf("Exported %d connections to %s\n", len(cfg.Connections), exportPath)
		fmt.Println("Warning: exported file is NOT encrypted")
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportPath, "output", "o", "", "output file path")
}
