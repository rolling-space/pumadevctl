package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds user-configurable settings loaded from XDG config.
// Currently only the mappings directory and port allocation defaults are supported.
// JSON example (~/.config/pumadevctl/config.json):
// {
//   "dir": "/Users/alice/.puma-dev",
//   "port_min": 36000,
//   "port_max": 37000,
//   "port_block_size": 10
// }
// All fields are optional; sensible defaults are applied.
// If XDG variable is not set, falls back to ~/.config.
// The config file itself is optional.
//
// This package avoids introducing extra dependencies by using JSON.

type AppConfig struct {
	Dir           string `json:"dir"`
	PortMin       int    `json:"port_min"`
	PortMax       int    `json:"port_max"`
	PortBlockSize int    `json:"port_block_size"`
}

// DefaultAppConfig returns built-in defaults matching previous behavior.
func DefaultAppConfig() AppConfig {
	home, _ := os.UserHomeDir()
	return AppConfig{
		Dir:           filepath.Join(home, ".puma-dev"),
		PortMin:       36000,
		PortMax:       37000,
		PortBlockSize: 10,
	}
}

// XDGConfigDir returns the directory to place pumadevctl's config, respecting XDG.
func XDGConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "pumadevctl")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pumadevctl")
}

// ConfigPath returns the path to pumadevctl's JSON config file.
func ConfigPath() string { return filepath.Join(XDGConfigDir(), "config.json") }

// LoadAppConfig reads the JSON config file if present and merges it with defaults.
func LoadAppConfig() (AppConfig, error) {
	cfg := DefaultAppConfig()
	path := ConfigPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	var fileCfg AppConfig
	if err := json.Unmarshal(b, &fileCfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	// Merge: only override when non-zero values are provided.
	if fileCfg.Dir != "" {
		cfg.Dir = fileCfg.Dir
	}
	if fileCfg.PortMin != 0 {
		cfg.PortMin = fileCfg.PortMin
	}
	if fileCfg.PortMax != 0 {
		cfg.PortMax = fileCfg.PortMax
	}
	if fileCfg.PortBlockSize != 0 {
		cfg.PortBlockSize = fileCfg.PortBlockSize
	}
	return cfg, nil
}

// ResolveDir validates the directory coming from flag or config.
// If dirFlag is empty, it falls back to the directory in the XDG-based config.
func ResolveDir(dirFlag string) (string, error) {
	var dir string
	if dirFlag != "" {
		dir = dirFlag
	} else {
		cfg, err := LoadAppConfig()
		if err != nil {
			return "", err
		}
		dir = cfg.Dir
	}
	if dir == "" {
		return "", fmt.Errorf("no directory specified")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("directory %s does not exist", abs)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("path %s is not a directory", abs)
	}
	return abs, nil
}
