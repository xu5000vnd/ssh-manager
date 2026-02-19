package app

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestCheckHealthReachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
			t.Skip("network listen not permitted in this environment")
		}
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	ok, err := CheckHealth("127.0.0.1", addr.Port, time.Second)
	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected reachable=true")
	}
}

func TestCheckHealthUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
			t.Skip("network listen not permitted in this environment")
		}
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	ok, err := CheckHealth("127.0.0.1", port, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for unreachable target")
	}
	if ok {
		t.Fatal("expected reachable=false")
	}
}
