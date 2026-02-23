package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"ssh-manager/internal/config"
	"ssh-manager/internal/sshcmd"
)

type connectResultMsg struct {
	Logs []string
	Err  error
}

func runConnectPreflight(conn config.Connection) tea.Cmd {
	return func() tea.Msg {
		base := sshcmd.BuildCommand(conn)
		if len(base) < 2 {
			return connectResultMsg{Err: errors.New("invalid ssh command")}
		}

		// Preflight runs a non-interactive verbose SSH probe so user can see progress logs.
		args := []string{
			"-v",
			"-o", "BatchMode=yes",
			"-o", "ConnectTimeout=8",
			"-o", "NumberOfPasswordPrompts=0",
		}
		args = append(args, base[1:]...)
		args = append(args, "exit")

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "ssh", args...)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return connectResultMsg{Err: fmt.Errorf("stderr pipe: %w", err)}
		}

		if err := cmd.Start(); err != nil {
			return connectResultMsg{Err: fmt.Errorf("start ssh preflight: %w", err)}
		}

		var logs []string
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			logs = append(logs, line)
		}
		if err := scanner.Err(); err != nil {
			logs = append(logs, "log read error: "+err.Error())
		}

		if err := cmd.Wait(); err != nil {
			return connectResultMsg{Logs: logs, Err: err}
		}
		return connectResultMsg{Logs: logs}
	}
}

