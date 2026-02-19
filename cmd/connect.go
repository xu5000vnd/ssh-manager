package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
	"ssh-manager/internal/sshcmd"
)

var connectCmd = &cobra.Command{
	Use:   "connect <name>",
	Short: "Connect directly by connection name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, pw, err := loadConfigWithPassword()
		if err != nil {
			return err
		}

		idx := findConnectionIndexByName(cfg.Connections, args[0])
		if idx < 0 {
			s := closestName(cfg.Connections, args[0])
			if s != "" {
				return fmt.Errorf("connection %q not found, did you mean %q", args[0], s)
			}
			return fmt.Errorf("connection %q not found", args[0])
		}

		conn := cfg.Connections[idx]
		conn.ConnectCount++
		conn.LastConnectedAt = time.Now().Format(time.RFC3339)
		cfg.Connections[idx] = conn

		if err := config.Save(cfg, path, pw); err != nil {
			return err
		}

		argsOut := sshcmd.BuildCommand(conn)
		return sshcmd.Exec(argsOut)
	},
}
