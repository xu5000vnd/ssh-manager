package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}

	dir := filepath.Join(base, "ssh-manager")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	return dir, nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.enc"), nil
}

func atomicWrite(path string, data []byte, perm os.FileMode) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func rotateBackups(path string) error {
	third := path + ".3"
	second := path + ".2"
	first := path + ".1"

	_ = os.Remove(third)
	if _, err := os.Stat(second); err == nil {
		if err := os.Rename(second, third); err != nil {
			return fmt.Errorf("rotate .2 to .3: %w", err)
		}
	}
	if _, err := os.Stat(first); err == nil {
		if err := os.Rename(first, second); err != nil {
			return fmt.Errorf("rotate .1 to .2: %w", err)
		}
	}
	if _, err := os.Stat(path); err == nil {
		if err := os.Rename(path, first); err != nil {
			return fmt.Errorf("rotate config to .1: %w", err)
		}
	}

	return nil
}
