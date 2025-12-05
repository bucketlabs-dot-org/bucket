package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIBase string `json:"api_base"`
	APIKey  string `json:"api_key"`
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "bucket_config.json"
	}
	return filepath.Join(home, ".config", "bucket", "config.json")
}

func Load() (*Config, error) {
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		// default config
		return &Config{
			APIBase: "http://localhost:8080",
			APIKey:  "",
		}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.APIBase == "" {
		cfg.APIBase = "http://localhost:8080"
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
