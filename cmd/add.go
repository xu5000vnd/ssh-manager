package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var (
	addName      string
	addHost      string
	addPort      int
	addUser      string
	addKey       string
	addGroup     string
	addProxyJump string
	addExtraArgs string
	addTags      string
	addNotes     string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(addName) == "" {
			return errors.New("--name is required")
		}
		if strings.TrimSpace(addHost) == "" {
			return errors.New("--host is required")
		}
		if addPort < 1 || addPort > 65535 {
			return errors.New("--port must be between 1 and 65535")
		}

		cfg, path, pw, err := loadConfigWithPassword()
		if err != nil {
			return err
		}
		if findConnectionIndexByName(cfg.Connections, addName) >= 0 {
			return fmt.Errorf("connection %q already exists", addName)
		}

		cfg.Connections = append(cfg.Connections, config.Connection{
			ID:        uuid.NewString(),
			Name:      addName,
			Group:     addGroup,
			Host:      addHost,
			Port:      addPort,
			Username:  addUser,
			KeyPath:   addKey,
			ProxyJump: addProxyJump,
			ExtraArgs: addExtraArgs,
			Tags:      addTags,
			Notes:     addNotes,
		})

		if err := config.Save(cfg, path, pw); err != nil {
			return err
		}
		fmt.Printf("Added connection %q\n", addName)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "connection name")
	addCmd.Flags().StringVar(&addHost, "host", "", "host")
	addCmd.Flags().IntVar(&addPort, "port", 22, "port")
	addCmd.Flags().StringVar(&addUser, "user", "", "username")
	addCmd.Flags().StringVar(&addKey, "key", "", "private key path")
	addCmd.Flags().StringVar(&addGroup, "group", "", "group")
	addCmd.Flags().StringVar(&addProxyJump, "proxy-jump", "", "proxy jump")
	addCmd.Flags().StringVar(&addExtraArgs, "extra-args", "", "extra ssh args")
	addCmd.Flags().StringVar(&addTags, "tags", "", "tags")
	addCmd.Flags().StringVar(&addNotes, "notes", "", "notes")
}
