package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ssh-manager/internal/config"
)

func TestAutoImportIfExistsMissingFile(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg := config.New()
	status, err := autoImportIfExists(cfg, filepath.Join(tmp, "config.enc"), []byte("pw"))
	if err != nil {
		t.Fatalf("autoImportIfExists returned error: %v", err)
	}
	if status != "" {
		t.Fatalf("expected empty status when file missing, got %q", status)
	}
}

func TestAutoImportIfExistsMergesAndSaves(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	jsonData := `{
  "version": 1,
  "connections": [
    {"id":"1","name":"dup","host":"example.com","port":22,"username":"u"},
    {"id":"2","name":"new-item","host":"new.example.com","port":2222,"username":"admin"}
  ]
}`
	if err := os.WriteFile(filepath.Join(tmp, autoImportFile), []byte(jsonData), 0o600); err != nil {
		t.Fatalf("write import file: %v", err)
	}

	cfg := &config.Config{
		Version: 1,
		Connections: []config.Connection{{
			ID: "x", Name: "dup", Host: "example.com", Port: 22, Username: "u",
		}},
	}
	encPath := filepath.Join(tmp, "config.enc")
	pw := []byte("pw")

	status, err := autoImportIfExists(cfg, encPath, pw)
	if err != nil {
		t.Fatalf("autoImportIfExists returned error: %v", err)
	}
	if !strings.Contains(status, "Auto-imported 1 connections") {
		t.Fatalf("unexpected status: %q", status)
	}
	if len(cfg.Connections) != 2 {
		t.Fatalf("expected 2 connections after merge, got %d", len(cfg.Connections))
	}

	loaded, err := config.Load(encPath, pw)
	if err != nil {
		t.Fatalf("load encrypted config failed: %v", err)
	}
	if len(loaded.Connections) != 2 {
		t.Fatalf("expected saved config to have 2 connections, got %d", len(loaded.Connections))
	}
}
