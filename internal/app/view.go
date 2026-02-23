package app

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	if m.Width > 0 && m.Height > 0 && (m.Width < 40 || m.Height < 10) {
		return "Please resize your terminal"
	}

	switch m.Screen {
	case ScreenUnlock:
		return strings.Join([]string{
			titleStyle.Render("SSH Manager"),
			"",
			"Enter master password:",
			m.PasswordInput.View(),
			"",
			statusStyle.Render(m.Status),
			"",
			"Press Enter to unlock, q to quit.",
		}, "\n")
	case ScreenHome:
		return strings.Join([]string{
			titleStyle.Render("Connections"),
			"",
			renderConnectionRows(m.Config, m.SelectedIndex),
			"",
			statusStyle.Render(m.Status),
			"",
			fmt.Sprintf("Total: %d", len(m.Config.Connections)),
			"Keys: Up/Down (or j/k) select, Enter/c connect, q quit.",
		}, "\n")
	case ScreenConnectLog:
		head := "SSH Connect Logs"
		if m.Config != nil && len(m.Config.Connections) > 0 && m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Config.Connections) {
			c := m.Config.Connections[m.SelectedIndex]
			head = fmt.Sprintf("SSH Connect Logs: %s (%s@%s:%d)", c.Name, c.Username, c.Host, c.Port)
		}
		logBody := "(no logs yet)"
		if len(m.ConnectLogs) > 0 {
			logBody = strings.Join(m.ConnectLogs, "\n")
		}
		errLine := ""
		if m.ConnectErr != "" {
			errLine = "Error: " + m.ConnectErr
		}
		bottom := "Press Esc to go back."
		if m.Connecting {
			bottom = "Connecting... collecting logs"
		} else if m.PreflightOK {
			bottom = "Opening SSH session..."
		}
		return strings.Join([]string{
			titleStyle.Render(head),
			"",
			statusStyle.Render(m.Status),
			statusStyle.Render(errLine),
			"",
			logBody,
			"",
			bottom,
		}, "\n")
	default:
		return "Not implemented"
	}
}
