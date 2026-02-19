package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var upgradeSource string

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade ssh-manager using make/install from source repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := resolveUpgradeSource(upgradeSource)
		if err != nil {
			return err
		}

		makeFile := filepath.Join(src, "make", "install")
		if _, err := os.Stat(makeFile); err != nil {
			return fmt.Errorf("upgrade file not found: %s", makeFile)
		}

		c := exec.Command("make", "-f", "make/install", "upgrade")
		c.Dir = src
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	upgradeCmd.Flags().StringVar(&upgradeSource, "source", "", "source repo path containing make/install")
}

func resolveUpgradeSource(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	cwd, err := os.Getwd()
	if err == nil {
		if _, statErr := os.Stat(filepath.Join(cwd, "make", "install")); statErr == nil {
			return cwd, nil
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		defaultPath := filepath.Join(home, "ssh-manager")
		if _, statErr := os.Stat(filepath.Join(defaultPath, "make", "install")); statErr == nil {
			return defaultPath, nil
		}
	}

	return "", errors.New("cannot find source repo; run from repo folder or use --source /path/to/ssh-manager")
}

