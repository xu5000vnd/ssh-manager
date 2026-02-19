package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func ExportJSON(cfg *Config, path string) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	return nil
}

func ExportSelectedJSON(connections []Connection, path string) error {
	cfg := &Config{Version: 1, Connections: connections}
	return ExportJSON(cfg, path)
}

func ImportJSON(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read import file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse import json: %w", err)
	}

	if cfg.Version == 0 {
		return nil, errors.New("invalid version: must be >= 1")
	}
	if cfg.Connections == nil {
		cfg.Connections = []Connection{}
	}
	return &cfg, nil
}

func DetectDuplicates(existing, incoming []Connection) (newItems, duplicates []Connection) {
	nameSet := map[string]struct{}{}
	tupleSet := map[string]struct{}{}

	for _, c := range existing {
		nameSet[strings.ToLower(strings.TrimSpace(c.Name))] = struct{}{}
		tupleSet[dedupeTuple(c)] = struct{}{}
	}

	for _, c := range incoming {
		_, hasName := nameSet[strings.ToLower(strings.TrimSpace(c.Name))]
		_, hasTuple := tupleSet[dedupeTuple(c)]
		if hasName || hasTuple {
			duplicates = append(duplicates, c)
			continue
		}
		newItems = append(newItems, c)
		nameSet[strings.ToLower(strings.TrimSpace(c.Name))] = struct{}{}
		tupleSet[dedupeTuple(c)] = struct{}{}
	}

	return newItems, duplicates
}

func dedupeTuple(c Connection) string {
	return strings.ToLower(strings.TrimSpace(c.Host)) + "|" +
		fmt.Sprintf("%d", c.Port) + "|" +
		strings.ToLower(strings.TrimSpace(c.Username))
}
