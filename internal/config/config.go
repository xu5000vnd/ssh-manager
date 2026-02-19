package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	cryptoutil "ssh-manager/internal/crypto"
)

type Connection struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Group           string `json:"group,omitempty"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	KeyPath         string `json:"key_path"`
	ProxyJump       string `json:"proxy_jump,omitempty"`
	ExtraArgs       string `json:"extra_args,omitempty"`
	Tags            string `json:"tags,omitempty"`
	Notes           string `json:"notes,omitempty"`
	Favorite        bool   `json:"favorite"`
	ConnectCount    int    `json:"connect_count"`
	LastConnectedAt string `json:"last_connected_at,omitempty"`
}

type Config struct {
	Version     int          `json:"version"`
	Connections []Connection `json:"connections"`
}

func New() *Config {
	return &Config{Version: 1, Connections: []Connection{}}
}

func Load(path string, password []byte) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return New(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	plaintext, err := cryptoutil.Decrypt(b, password)
	if err != nil {
		return nil, fmt.Errorf("decrypt config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(plaintext, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Connections == nil {
		cfg.Connections = []Connection{}
	}

	return &cfg, nil
}

func Save(cfg *Config, path string, password []byte) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		return fmt.Errorf("write config: target is a directory: %s", path)
	}

	if err := rotateBackups(path); err != nil {
		return err
	}

	plaintext, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	ciphertext, err := cryptoutil.Encrypt(plaintext, password)
	if err != nil {
		return fmt.Errorf("encrypt config: %w", err)
	}

	if err := atomicWrite(path, ciphertext, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
