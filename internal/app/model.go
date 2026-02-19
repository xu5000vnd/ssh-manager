package app

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ssh-manager/internal/config"
	"ssh-manager/internal/sshcmd"
)

type Screen int

const (
	ScreenUnlock Screen = iota
	ScreenHome
	ScreenDetail
	ScreenForm
	ScreenImport
	ScreenExport
	ScreenPassword
)

type autoLockMsg struct {
	Token int
}

type Model struct {
	Screen         Screen
	Path           string
	Config         *config.Config
	PasswordInput  textinput.Model
	Status         string
	Password       []byte
	IdleTimeout    time.Duration
	Width          int
	Height         int
	SelectedIndex  int
	ActivityToken  int
	ConnectArgs    []string
}

func NewModel() (Model, error) {
	path, err := config.ConfigPath()
	if err != nil {
		return Model{}, err
	}
	input := textinput.New()
	input.Placeholder = "Master password"
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '*'
	input.Focus()
	return Model{
		Screen:        ScreenUnlock,
		Path:          path,
		Config:        config.New(),
		PasswordInput: input,
		IdleTimeout:   5 * time.Minute,
	}, nil
}

func Run() error {
	m, err := NewModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	final, ok := finalModel.(Model)
	if !ok {
		return nil
	}
	if len(final.ConnectArgs) > 0 {
		return sshcmd.Exec(final.ConnectArgs)
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	return m.lockTickCmd(m.ActivityToken)
}

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36"))
var statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

func checkFirstRun(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func (m Model) lockTickCmd(token int) tea.Cmd {
	return tea.Tick(m.IdleTimeout, func(time.Time) tea.Msg {
		return autoLockMsg{Token: token}
	})
}

func renderConnectionRows(cfg *config.Config, selected int) string {
	if len(cfg.Connections) == 0 {
		return "No connections. Use CLI `ssh-manager add ...` to add one.\n"
	}
	out := ""
	for i, c := range cfg.Connections {
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		out += fmt.Sprintf("%s%d. %s (%s@%s:%d)\n", prefix, i+1, c.Name, c.Username, c.Host, c.Port)
	}
	return out
}
