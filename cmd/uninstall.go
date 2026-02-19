package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uninstallForce bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall ssh-manager and remove local encrypted data",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		bin := filepath.Join(home, ".local", "bin", "ssh-manager")
		legacy := filepath.Join(home, ".local", "bin", "ssh-management")
		dataDir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		configPath := filepath.Join(dataDir, "ssh-manager", "config.enc")

		if !uninstallForce {
			ok, err := askYesNo(fmt.Sprintf("Uninstall binary and delete all local data at %s(.1/.2/.3)? [y/N]: ", configPath))
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Canceled")
				return nil
			}
		}

		_ = os.Remove(bin)
		_ = os.Remove(legacy)
		_ = os.Remove(configPath)
		_ = os.Remove(configPath + ".1")
		_ = os.Remove(configPath + ".2")
		_ = os.Remove(configPath + ".3")

		fmt.Printf("Uninstalled binary from %s\n", filepath.Dir(bin))
		fmt.Printf("Deleted local encrypted data: %s(.1/.2/.3)\n", configPath)
		return nil
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "skip confirmation")
}
