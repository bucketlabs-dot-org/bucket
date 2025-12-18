package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIBase   	string `json:"api_base"`
	APIKey    	string `json:"api_key"`
	DeviceID  	string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Tier      	string `json:"tier"`
	UsedBytes 	int64  `json:"used_bytes"`
	Quota     	int64  `json:"quota"`
}


func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "bucket_config.json"
	}
	return filepath.Join(home, ".config", "bucket", "config.json")
}

func Path() string {
	return configPath()
}

func Load() (*Config, error) {
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		// default config
		return &Config{
			APIBase: "https://api.bucketlabs.org",
			APIKey:  "",
		}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.APIBase == "" {
		cfg.APIBase = "https://api.bucketlabs.org"
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
