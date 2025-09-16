package config

import (
	"encoding/json"
	"go-pass-keeper/internal/grpcclient/settings"
	"os"
	"path/filepath"
)

type Config struct {
	configPath string
}

func NewConfig(appName string) *Config {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, appName, "config.json")
	return &Config{configPath: configPath}
}

func (cm *Config) Load() *settings.Connection {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return cm.DefaultConfig()
	}

	var config settings.Connection
	if json.Unmarshal(data, &config) != nil {
		return cm.DefaultConfig()
	}

	return &config
}

func (cm *Config) Save(config *settings.Connection) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(cm.configPath), 0755)
	return os.WriteFile(cm.configPath, data, 0644)
}

func (cm *Config) DefaultConfig() *settings.Connection {
	return &settings.Connection{
		ServerURL:  "localhost",
		ServerPort: "8080",
		Timeout:    30,
	}
}
