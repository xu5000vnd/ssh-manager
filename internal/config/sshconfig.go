package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseSSHConfig(path string) ([]Connection, error) {
	resolved := expandPath(path)
	f, err := os.Open(resolved)
	if err != nil {
		return nil, fmt.Errorf("open ssh config: %w", err)
	}
	defer f.Close()

	var out []Connection
	var currentAliases []string
	current := Connection{Port: 22}
	inBlock := false

	flush := func() {
		if !inBlock || len(currentAliases) == 0 {
			return
		}
		for _, alias := range currentAliases {
			if shouldSkipHostAlias(alias) {
				continue
			}
			conn := current
			conn.Name = alias
			if conn.Host == "" {
				conn.Host = alias
			}
			out = append(out, conn)
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch key {
		case "host":
			flush()
			inBlock = true
			current = Connection{Port: 22}
			currentAliases = strings.Fields(value)
		case "hostname":
			if inBlock {
				current.Host = value
			}
		case "user":
			if inBlock {
				current.Username = value
			}
		case "port":
			if inBlock {
				if p, err := strconv.Atoi(value); err == nil {
					current.Port = p
				}
			}
		case "identityfile":
			if inBlock {
				current.KeyPath = value
			}
		case "proxyjump":
			if inBlock {
				current.ProxyJump = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan ssh config: %w", err)
	}
	flush()

	return out, nil
}

func GenerateSSHConfig(connections []Connection) string {
	var b strings.Builder
	for i, c := range connections {
		if strings.TrimSpace(c.Name) == "" {
			continue
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Host ")
		b.WriteString(c.Name)
		b.WriteString("\n")
		if c.Host != "" {
			b.WriteString("  HostName ")
			b.WriteString(c.Host)
			b.WriteString("\n")
		}
		if c.Username != "" {
			b.WriteString("  User ")
			b.WriteString(c.Username)
			b.WriteString("\n")
		}
		port := c.Port
		if port == 0 {
			port = 22
		}
		b.WriteString("  Port ")
		b.WriteString(strconv.Itoa(port))
		b.WriteString("\n")
		if c.KeyPath != "" {
			b.WriteString("  IdentityFile ")
			b.WriteString(c.KeyPath)
			b.WriteString("\n")
		}
		if c.ProxyJump != "" {
			b.WriteString("  ProxyJump ")
			b.WriteString(c.ProxyJump)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func shouldSkipHostAlias(alias string) bool {
	alias = strings.TrimSpace(alias)
	return alias == "*" || strings.ContainsAny(alias, "*?")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
