package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/term"
	"ssh-manager/internal/config"
)

func getMasterPassword() ([]byte, error) {
	if v := os.Getenv("SSH_MANAGER_PASSWORD"); v != "" {
		return []byte(v), nil
	}

	fmt.Fprint(os.Stdout, "Master password: ")
	line, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stdout)
	pw := strings.TrimSpace(string(line))
	if pw == "" {
		return nil, errors.New("password cannot be empty")
	}
	return []byte(pw), nil
}

func loadConfigWithPassword() (*config.Config, string, []byte, error) {
	path, err := config.ConfigPath()
	if err != nil {
		return nil, "", nil, err
	}
	pw, err := getMasterPassword()
	if err != nil {
		return nil, "", nil, err
	}
	cfg, err := config.Load(path, pw)
	if err != nil {
		return nil, "", nil, err
	}
	return cfg, path, pw, nil
}

func findConnectionIndexByName(conns []config.Connection, name string) int {
	name = strings.TrimSpace(strings.ToLower(name))
	for i, c := range conns {
		if strings.ToLower(strings.TrimSpace(c.Name)) == name {
			return i
		}
	}
	return -1
}

func closestName(conns []config.Connection, name string) string {
	if len(conns) == 0 {
		return ""
	}
	name = strings.ToLower(strings.TrimSpace(name))
	scores := make([]struct {
		name  string
		score int
	}, 0, len(conns))
	for _, c := range conns {
		n := strings.ToLower(strings.TrimSpace(c.Name))
		score := commonPrefix(name, n)
		scores = append(scores, struct {
			name  string
			score int
		}{name: c.Name, score: score})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if scores[0].score == 0 {
		return ""
	}
	return scores[0].name
}

func commonPrefix(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	count := 0
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			break
		}
		count++
	}
	return count
}

func askYesNo(prompt string) (bool, error) {
	fmt.Fprint(os.Stdout, prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "y" || ans == "yes", nil
}
