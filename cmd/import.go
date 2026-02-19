package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var importCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import connections from JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, pw, err := loadConfigWithPassword()
		if err != nil {
			return err
		}

		incoming, err := config.ImportJSON(args[0])
		if err != nil {
			return err
		}

		newItems, duplicates := config.DetectDuplicates(cfg.Connections, incoming.Connections)
		fmt.Printf("Found %d connections (%d duplicates)\n", len(incoming.Connections), len(duplicates))
		fmt.Print("Choose action: [m]erge, [r]eplace, [c]ancel: ")

		r := bufio.NewReader(os.Stdin)
		choice, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "m", "merge":
			cfg.Connections = append(cfg.Connections, newItems...)
		case "r", "replace":
			cfg.Connections = incoming.Connections
		default:
			fmt.Println("Canceled")
			return nil
		}

		if err := config.Save(cfg, path, pw); err != nil {
			return err
		}
		fmt.Printf("Imported %d connections\n", len(incoming.Connections))
		return nil
	},
}
