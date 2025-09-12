package config

import (
	"go-pass-keeper/internal/grpcclient/settings"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	manager := NewConfig("testapp")

	t.Run("DefaultConfig", func(t *testing.T) {
		defaultConfig := manager.DefaultConfig()
		if defaultConfig.ServerURL != "localhost" {
			t.Errorf("Expected localhost, got %s", defaultConfig.ServerURL)
		}
		if defaultConfig.Timeout != 30 {
			t.Errorf("Expected timeout 30, got %d", defaultConfig.Timeout)
		}
	})

	t.Run("LoadNonExistentConfig", func(t *testing.T) {
		config := manager.Load()
		if config.ServerURL != "localhost" {
			t.Errorf("Expected default config, got %s", config.ServerURL)
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		testConfig := &settings.Connection{
			ServerURL:  "example.com",
			ServerPort: "9000",
			Timeout:    60,
		}

		if err := manager.Save(testConfig); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loadedConfig := manager.Load()
		if loadedConfig.ServerURL != "example.com" {
			t.Errorf("Expected example.com, got %s", loadedConfig.ServerURL)
		}
		if loadedConfig.Timeout != 60 {
			t.Errorf("Expected timeout 60, got %d", loadedConfig.Timeout)
		}
	})

	t.Run("ConfigFileExists", func(t *testing.T) {
		if _, err := os.Stat(manager.configPath); os.IsNotExist(err) {
			t.Errorf("Config file should exist after save")
		}
	})
}

func TestInvalidJSON(t *testing.T) {
	manager := NewConfig("testapp")

	// Создаем битый JSON файл
	os.MkdirAll(filepath.Dir(manager.configPath), 0755)
	os.WriteFile(manager.configPath, []byte("{invalid json}"), 0644)

	// Должен вернуть конфиг по умолчанию
	config := manager.Load()
	if config.ServerURL != "localhost" {
		t.Errorf("Should return default config for invalid JSON")
	}
}
