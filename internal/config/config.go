package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Host struct {
	User     string `yaml:"user"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	BasePath string `yaml:"base_path"`
}

type Folder struct {
	LocalPath    string `yaml:"local_path"`
	RemoteHost   string `yaml:"remote_host"`
	RemoteSubpath string `yaml:"remote_subpath"`
	Mode         string `yaml:"mode"` // "push", "pull", "bidirectional"
	Delete       bool   `yaml:"delete"`
	Checksum     bool   `yaml:"checksum"`
}

type Config struct {
	Hosts   map[string]Host   `yaml:"hosts"`
	Folders map[string]Folder `yaml:"folders"`
}

var globalConfig *Config

func Load() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, ".mirusync", "config.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	globalConfig = &cfg
	return globalConfig, nil
}

func GetHost(name string) (*Host, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	host, ok := cfg.Hosts[name]
	if !ok {
		return nil, fmt.Errorf("host '%s' not found in configuration", name)
	}

	return &host, nil
}

func GetFolder(name string) (*Folder, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	folder, ok := cfg.Folders[name]
	if !ok {
		return nil, fmt.Errorf("folder '%s' not found in configuration", name)
	}

	return &folder, nil
}

func GetConfigPath() string {
	if viper.ConfigFileUsed() != "" {
		return viper.ConfigFileUsed()
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mirusync", "config.yaml")
}

