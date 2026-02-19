package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var resetForce bool

var resetCmd = &cobra.Command{
	Use:   "reset-all",
	Short: "Delete all saved connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, pw, err := loadConfigWithPassword()
		if err != nil {
			return err
		}

		if len(cfg.Connections) == 0 {
			fmt.Println("No connections to reset")
			return nil
		}

		if !resetForce {
			ok, err := askYesNo(fmt.Sprintf("Delete ALL %d connections? [y/N]: ", len(cfg.Connections)))
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Canceled")
				return nil
			}
		}

		cfg.Connections = []config.Connection{}
		if err := config.Save(cfg, path, pw); err != nil {
			return err
		}
		fmt.Println("All connections deleted")
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVar(&resetForce, "force", false, "skip confirmation")
}

