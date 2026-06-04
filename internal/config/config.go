package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	LogLines int `json:"logLines"`
}

func DefaultConfig() Config {
	return Config{
		LogLines: 500,
	}
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "terraforge", "config.json")
}

func Load() Config {
	cfg := DefaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func Save(cfg Config) error {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
