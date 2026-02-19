package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.enc")
	password := []byte("pw")

	cfg := &Config{
		Version: 1,
		Connections: []Connection{{
			ID:       "1",
			Name:     "prod",
			Host:     "example.com",
			Port:     22,
			Username: "admin",
		}},
	}

	if err := Save(cfg, path, password); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := Load(path, password)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(got.Connections) != 1 || got.Connections[0].Name != "prod" {
		t.Fatalf("unexpected loaded config: %#v", got)
	}
}

func TestFirstRunNoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.enc")

	cfg, err := Load(path, []byte("pw"))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Version != 1 {
		t.Fatalf("got version %d, want 1", cfg.Version)
	}
	if len(cfg.Connections) != 0 {
		t.Fatalf("expected no connections")
	}
}

func TestBackupsAreCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.enc")
	password := []byte("pw")

	for i := 0; i < 4; i++ {
		cfg := &Config{Version: 1, Connections: []Connection{{Name: "n"}}}
		if err := Save(cfg, path, password); err != nil {
			t.Fatalf("Save failed at %d: %v", i, err)
		}
	}

	for _, suffix := range []string{".1", ".2", ".3"} {
		if _, err := os.Stat(path + suffix); err != nil {
			t.Fatalf("expected backup %s: %v", suffix, err)
		}
	}
}

func TestAtomicWriteNoCorruptionOnDirPath(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(badPath, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	cfg := &Config{Version: 1}
	err := Save(cfg, badPath, []byte("pw"))
	if err == nil {
		t.Fatal("expected save error when path is directory")
	}
}
