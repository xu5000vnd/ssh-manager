package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"ssh-manager/internal/config"
)

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change master password",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		fmt.Print("Current password: ")
		currentRaw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println()
		current := strings.TrimSpace(string(currentRaw))

		cfg, err := config.Load(path, []byte(current))
		if err != nil {
			return fmt.Errorf("verify current password: %w", err)
		}

		fmt.Print("New password: ")
		newPwRaw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Print("Confirm new password: ")
		confirmRaw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println()
		newPw := strings.TrimSpace(string(newPwRaw))
		confirm := strings.TrimSpace(string(confirmRaw))

		if newPw == "" {
			return errors.New("new password cannot be empty")
		}
		if newPw != confirm {
			return errors.New("new passwords do not match")
		}

		if err := config.Save(cfg, path, []byte(newPw)); err != nil {
			return err
		}
		fmt.Println("Master password changed")
		return nil
	},
}
