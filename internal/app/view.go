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
	default:
		return "Not implemented"
	}
}
