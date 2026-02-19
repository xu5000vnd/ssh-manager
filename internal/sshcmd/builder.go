package sshcmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ssh-manager/internal/config"
)

func BuildCommand(conn config.Connection) []string {
	args := []string{"ssh"}

	if conn.ProxyJump != "" {
		args = append(args, "-J", conn.ProxyJump)
	}

	if conn.KeyPath != "" {
		args = append(args, "-i", resolveHome(conn.KeyPath))
	}

	port := conn.Port
	if port == 0 {
		port = 22
	}
	args = append(args, "-p", strconv.Itoa(port))

	if conn.ExtraArgs != "" {
		args = append(args, strings.Fields(conn.ExtraArgs)...)
	}

	target := conn.Host
	if conn.Username != "" {
		target = conn.Username + "@" + conn.Host
	}
	args = append(args, target)

	return args
}

func resolveHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~/"))
		}
	}
	return p
}
