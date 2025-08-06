// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.

package config

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/goo-apps/vpnctl/internal/model"
)

var (
	VPN_BINARY_PATH            string
	VPN_GUI_PATH               string
	VPN_CONNECTION_RETRY_COUNT int
	SQLITE_DB_PATH             string
	APPLICATION_ENVIRONMENT    string
	KEYRING_SERVICE_NAME       string
	KEYRING_ENCRYPTION_KEY     string
	LOGGER_LEVEL               int
	APPLICATION_VERSION        string
)

type ConfigReader struct {
	data map[string]interface{}
}

// Embed the config file content (adjust path as needed)
//
//go:embed resource.toml
var embeddedConfig []byte

// LoadConfig loads config from embedded file or optionally from external path if given
func LoadConfig(path string) (*model.Config, error) {
	cfg := &model.Config{}

	if path == "" {
		// Load from embedded
		err := toml.Unmarshal(embeddedConfig, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedded config: %w", err)
		}
		return cfg, nil
	}

	// If user explicitly passed path, load from file system (optional fallback)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	err = toml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Load returns a ConfigReader from TOML file path.
// func Load(path string) (*ConfigReader, error) {
// 	defaultConfigPath := filepath.Join(os.Getenv("HOME"), ".vpnctl", "resource.toml")

// 	if path == "" {
// 		path = defaultConfigPath
// 	}

// 	data := make(map[string]interface{})
// 	if _, err := os.Stat(path); os.IsNotExist(err) {
// 		return nil, fmt.Errorf("config file not found: %s", path)
// 	}

// 	if _, err := toml.DecodeFile(path, &data); err != nil {
// 		return nil, fmt.Errorf("failed to parse TOML: %w", err)
// 	}

// 	return &ConfigReader{data}, nil
// }

// get traverses dot-separated keys like "redis.pool_size"
func (c *ConfigReader) get(path string) interface{} {
	keys := strings.Split(path, ".")
	var value interface{} = c.data

	for _, key := range keys {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value, ok = m[key]
		if !ok {
			return nil
		}
	}
	return value
}

func (c *ConfigReader) GetString(path string) string {
	if val := c.get(path); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (c *ConfigReader) GetInt(path string) int {
	if val := c.get(path); val != nil {
		switch v := val.(type) {
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			i, _ := strconv.Atoi(v)
			return i
		}
	}
	return -1 // Return -1 to indicate not found or error
}

func (c *ConfigReader) GetBool(path string) bool {
	if val := c.get(path); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Load all configuration at once from the default resource.toml file
func LoadAllConfigAtOnce(configPath string) error {
	vr, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	VPN_BINARY_PATH = vr.VPN.BinaryPath
	VPN_GUI_PATH = vr.VPN.GuiPath
	VPN_CONNECTION_RETRY_COUNT = vr.VPN.ConnectionRetry
	SQLITE_DB_PATH = vr.Sqlite.Path
	APPLICATION_ENVIRONMENT = vr.Application.Environment
	KEYRING_SERVICE_NAME = vr.Keyring.ServiceName
	KEYRING_ENCRYPTION_KEY = vr.Keyring.EncryptionKey
	LOGGER_LEVEL = vr.Logger.LoggerLevel
	APPLICATION_VERSION = vr.Application.Version

	// print all loaded configurations for debugging
	// for key, value := range vr.data {
	// 	logger.Infof("Config: %s = %v", key, value)
	// }

	// vpnCmd = "/opt/cisco/secureclient/bin/vpn"
	// guiApp = "/Applications/Cisco/Cisco Secure Client.app/"

	// Load Vault configuration from environment variables

	// VPN_BINARY_PATH = vr.GetString("vpn.binary_path")
	// if VPN_BINARY_PATH == "" {
	// 	logger.Fatalf("VPN binary path is not set in the configuration file")
	// 	return
	// }

	// VPN_GUI_PATH = vr.GetString("vpn.gui_path")
	// if VPN_GUI_PATH == "" {
	// 	logger.Fatalf("VPN GUI path is not set in the configuration file")
	// 	return
	// }

	// VPN_CONNECTION_RETRY_COUNT = vr.GetInt("vpn.connection_retry")
	// if VPN_CONNECTION_RETRY_COUNT < 0 {
	// 	logger.Fatalf("VPN connection retry count must be greater than or equal to 0")
	// 	return
	// }

	// SQLITE_DB_PATH = vr.GetString("sqlite.path")
	// if SQLITE_DB_PATH == "" {
	// 	logger.Fatalf("SQLite database path is not set in the configuration file")
	// 	return
	// }

	return nil
}
