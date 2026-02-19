package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, pw, err := loadConfigWithPassword()
		if err != nil {
			return err
		}
		idx := findConnectionIndexByName(cfg.Connections, args[0])
		if idx < 0 {
			return fmt.Errorf("connection %q not found", args[0])
		}

		ok, err := askYesNo(fmt.Sprintf("Delete %q? [y/N]: ", cfg.Connections[idx].Name))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Canceled")
			return nil
		}

		cfg.Connections = append(cfg.Connections[:idx], cfg.Connections[idx+1:]...)
		if err := config.Save(cfg, path, pw); err != nil {
			return err
		}
		fmt.Printf("Deleted connection %q\n", args[0])
		return nil
	},
}
