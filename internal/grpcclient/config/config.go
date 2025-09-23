package config

import (
	"encoding/json"
	"go-pass-keeper/internal/grpcclient/settings"
	"os"
	"path/filepath"
)

// Config - модель конфига
type Config struct {
	configPath string
}

// NewConfig - метод создания нового конфига из файла
func NewConfig(appName string) *Config {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, appName, "config.json")
	return &Config{configPath: configPath}
}

// Load - загрузка конфига из файла
func (cm *Config) Load() *settings.Settings {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return cm.DefaultConfig()
	}

	config := cm.DefaultConfig()
	if json.Unmarshal(data, config) != nil {
		return config
	}

	return config
}

// Load - сохранение конфига в файл
func (cm *Config) Save(config *settings.Settings) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(cm.configPath), 0755)
	return os.WriteFile(cm.configPath, data, 0644)
}

// DefaultConfig - дефолтный конфиг
func (cm *Config) DefaultConfig() *settings.Settings {
	return &settings.Settings{
		ServerURL:  "localhost",
		ServerPort: "8080",
		Timeout:    30,
		Secret:     "secret",
	}
}
