package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var importSSHCmd = &cobra.Command{
	Use:   "import-ssh-config [path]",
	Short: "Import from OpenSSH config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, pw, err := loadConfigWithPassword()
		if err != nil {
			return err
		}

		sshPath := "~/.ssh/config"
		if len(args) == 1 {
			sshPath = args[0]
		}

		inConns, err := config.ParseSSHConfig(sshPath)
		if err != nil {
			return err
		}
		incoming := &config.Config{Version: 1, Connections: inConns}

		newItems, duplicates := config.DetectDuplicates(cfg.Connections, incoming.Connections)
		fmt.Printf("Parsed %d connections from %s (%d duplicates)\n", len(incoming.Connections), filepath.Clean(sshPath), len(duplicates))
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
