package app

import (
	"errors"
	"fmt"
	"os"

	"ssh-manager/internal/config"
)

const autoImportFile = "ssh-connections.json"

func autoImportIfExists(cfg *config.Config, encryptedPath string, password []byte) (string, error) {
	if _, err := os.Stat(autoImportFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("check auto-import file: %w", err)
	}

	incoming, err := config.ImportJSON(autoImportFile)
	if err != nil {
		return "", fmt.Errorf("auto-import parse: %w", err)
	}

	newItems, duplicates := config.DetectDuplicates(cfg.Connections, incoming.Connections)
	if len(newItems) == 0 {
		return fmt.Sprintf("Auto-import: %s found (%d duplicates, 0 new)", autoImportFile, len(duplicates)), nil
	}

	cfg.Connections = append(cfg.Connections, newItems...)
	if err := config.Save(cfg, encryptedPath, password); err != nil {
		return "", fmt.Errorf("auto-import save: %w", err)
	}

	return fmt.Sprintf("Auto-imported %d connections from %s (%d duplicates skipped)", len(newItems), autoImportFile, len(duplicates)), nil
}

