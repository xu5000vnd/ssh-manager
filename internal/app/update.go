package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"ssh-manager/internal/config"
	"ssh-manager/internal/sshcmd"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case autoLockMsg:
		if msg.Token != m.ActivityToken {
			return m, nil
		}
		m.Password = nil
		m.Screen = ScreenUnlock
		m.Status = "Session locked due to inactivity"
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		m.ActivityToken++
	case connectResultMsg:
		m.Connecting = false
		m.ConnectLogs = msg.Logs
		if msg.Err != nil {
			m.PreflightOK = false
			m.ConnectErr = msg.Err.Error()
			m.Status = "Preflight failed. Check logs and press Esc to return."
			return m, m.lockTickCmd(m.ActivityToken)
		}

		m.PreflightOK = true
		m.ConnectErr = ""
		m.Status = "Preflight successful. Starting SSH session..."
		if m.Config == nil || len(m.Config.Connections) == 0 {
			m.Status = "No connection to connect"
			return m, m.lockTickCmd(m.ActivityToken)
		}
		conn := m.Config.Connections[m.SelectedIndex]
		conn.ConnectCount++
		conn.LastConnectedAt = time.Now().Format(time.RFC3339)
		m.Config.Connections[m.SelectedIndex] = conn
		if err := config.Save(m.Config, m.Path, m.Password); err != nil {
			m.Status = "Failed to save: " + err.Error()
			return m, m.lockTickCmd(m.ActivityToken)
		}
		return m, tea.Quit
	}

	if m.Screen == ScreenUnlock {
		m.PasswordInput, cmd = m.PasswordInput.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" {
			pw := []byte(strings.TrimSpace(m.PasswordInput.Value()))
			if len(pw) == 0 {
				m.Status = "Password cannot be empty"
				return m, nil
			}
			if checkFirstRun(m.Path) {
				cfg := config.New()
				if err := config.Save(cfg, m.Path, pw); err != nil {
					m.Status = "Failed to initialize config: " + err.Error()
					return m, nil
				}
				importStatus, err := autoImportIfExists(cfg, m.Path, pw)
				if err != nil {
					m.Status = "Auto-import failed: " + err.Error()
					return m, nil
				}
				m.Config = cfg
				m.Password = pw
				m.Screen = ScreenHome
				m.SelectedIndex = 0
				if importStatus != "" {
					m.Status = importStatus
				} else {
					m.Status = "Initialized new encrypted config"
				}
				m.PasswordInput.SetValue("")
				return m, m.lockTickCmd(m.ActivityToken)
			}
			cfg, err := config.Load(m.Path, pw)
			if err != nil {
				m.Status = "Wrong password or corrupt config"
				m.PasswordInput.SetValue("")
				return m, nil
			}
			importStatus, err := autoImportIfExists(cfg, m.Path, pw)
			if err != nil {
				m.Status = "Auto-import failed: " + err.Error()
				return m, nil
			}
			m.Config = cfg
			m.Password = pw
			m.Screen = ScreenHome
			m.SelectedIndex = 0
			if importStatus != "" {
				m.Status = importStatus
			} else {
				m.Status = "Unlocked"
			}
			m.PasswordInput.SetValue("")
			return m, m.lockTickCmd(m.ActivityToken)
		}
		return m, tea.Batch(cmd, m.lockTickCmd(m.ActivityToken))
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		if km.String() == "esc" {
			if m.Screen == ScreenConnectLog {
				m.Screen = ScreenHome
				m.Connecting = false
				m.ConnectErr = ""
				m.PreflightOK = false
			} else {
				m.Screen = ScreenHome
			}
			return m, m.lockTickCmd(m.ActivityToken)
		}
		if m.Screen == ScreenHome {
			switch km.String() {
			case "up", "k":
				if m.SelectedIndex > 0 {
					m.SelectedIndex--
				}
			case "down", "j":
				if m.Config != nil && m.SelectedIndex < len(m.Config.Connections)-1 {
					m.SelectedIndex++
				}
			case "enter", "c":
				if m.Config == nil || len(m.Config.Connections) == 0 {
					m.Status = "No connection to connect"
					return m, m.lockTickCmd(m.ActivityToken)
				}
				conn := m.Config.Connections[m.SelectedIndex]
				m.Screen = ScreenConnectLog
				m.Connecting = true
				m.PreflightOK = false
				m.ConnectErr = ""
				m.ConnectLogs = []string{"Starting SSH preflight..."}
				m.ConnectArgs = sshcmd.BuildCommand(conn)
				m.Status = "Collecting SSH logs..."
				return m, tea.Batch(runConnectPreflight(conn), m.lockTickCmd(m.ActivityToken))
			}
		}
		if m.Screen == ScreenConnectLog {
			return m, m.lockTickCmd(m.ActivityToken)
		}
	}

	return m, m.lockTickCmd(m.ActivityToken)
}
