package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type AgentConfig struct {
	HubURL      string `json:"hub_url"`
	AgentAPIKey string `json:"agent_api_key"`
}

const AgentConfigPath = "/etc/syslantern-agent/config.json"

func ParseConfig() (AgentConfig, error) {
	b, err := os.ReadFile(AgentConfigPath)
	if err != nil {
		return AgentConfig{}, err
	}

	var cfg AgentConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return AgentConfig{}, err
	}
	if cfg.HubURL == "" {
		return AgentConfig{}, fmt.Errorf("missing hub URL in %s", AgentConfigPath)
	}
	if cfg.AgentAPIKey == "" {
		return AgentConfig{}, fmt.Errorf("missing agent API key, run syslantern set apikey <key>")
	}
	return cfg, nil
}

func SaveConfig(cfg AgentConfig) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(AgentConfigPath, append(b, '\n'), 0600)
}

func SetAPIKey(value string) error {
	// Ignore load errors so the installer can create the config file on first setup.
	cfg, err := ParseConfig()
	if err != nil {
		cfg = AgentConfig{HubURL: "", AgentAPIKey: ""}
	}
	cfg.AgentAPIKey = value
	return SaveConfig(cfg)
}
