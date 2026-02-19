package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUpgradeSourceExplicit(t *testing.T) {
	got, err := resolveUpgradeSource("/tmp/custom-src")
	if err != nil {
		t.Fatalf("resolveUpgradeSource returned error: %v", err)
	}
	if got != "/tmp/custom-src" {
		t.Fatalf("got %q want %q", got, "/tmp/custom-src")
	}
}

func TestResolveUpgradeSourceFromCWD(t *testing.T) {
	tmp := t.TempDir()
	makeDir := filepath.Join(tmp, "make")
	if err := os.MkdirAll(makeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(makeDir, "install"), []byte(""), 0o600); err != nil {
		t.Fatalf("write install marker: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, err := resolveUpgradeSource("")
	if err != nil {
		t.Fatalf("resolveUpgradeSource returned error: %v", err)
	}
	gotReal, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("eval symlinks (got): %v", err)
	}
	tmpReal, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatalf("eval symlinks (tmp): %v", err)
	}
	if gotReal != tmpReal {
		t.Fatalf("got %q want %q", gotReal, tmpReal)
	}
}

func TestResolveUpgradeSourceError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if _, err := resolveUpgradeSource(""); err == nil {
		t.Fatal("expected error when source cannot be discovered")
	}
}
