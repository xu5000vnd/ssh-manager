package sshcmd

import (
	"strings"
	"testing"

	"ssh-manager/internal/config"
)

func TestBuildCommandBasic(t *testing.T) {
	conn := config.Connection{Host: "example.com", Username: "alice", Port: 22, KeyPath: "~/.ssh/id_ed25519"}
	args := BuildCommand(conn)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "ssh") || !strings.Contains(joined, "alice@example.com") {
		t.Fatalf("unexpected command: %v", args)
	}
}

func TestBuildCommandWithProxyJump(t *testing.T) {
	conn := config.Connection{Host: "h", Username: "u", Port: 22, ProxyJump: "bastion"}
	args := BuildCommand(conn)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-J bastion") {
		t.Fatalf("missing proxy jump: %v", args)
	}
}

func TestBuildCommandWithExtraArgs(t *testing.T) {
	conn := config.Connection{Host: "h", Port: 22, ExtraArgs: "-L 8080:localhost:80 -v"}
	args := BuildCommand(conn)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-L 8080:localhost:80") || !strings.Contains(joined, "-v") {
		t.Fatalf("missing extra args: %v", args)
	}
	if args[len(args)-1] != "h" {
		t.Fatalf("expected ssh target to be final argument, got: %v", args)
	}
}

func TestBuildCommandDefaultPort(t *testing.T) {
	conn := config.Connection{Host: "h"}
	args := BuildCommand(conn)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-p 22") {
		t.Fatalf("expected default port: %v", args)
	}
}
