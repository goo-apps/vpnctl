// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.
package model

// USER_CREDENTIAL represents the structure of user credentials.
type USER_CREDENTIAL struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	SecondPassword string `json:"second_password"`
	YFlag          string `json:"y_flag"`
}

type CREDENTIAL_FOR_LOGIN struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Push     string `json:"-"`
	YFlag    string `json:"-"`
}

// Config holds the configuration settings for the application.
type Config struct {
	VPN struct {
		BinaryPath      string `toml:"binary_path"`
		GuiPath         string `toml:"gui_path"`
		ConnectionRetry int    `toml:"connection_retry"`
	} `toml:"vpn"`

	Sqlite struct {
		Path string `toml:"path"`
	} `toml:"sqlite"`

	Application struct {
		Environment string `toml:"environment"`
		Version     string `toml:"version"`
	} `toml:"application"`

	Keyring struct {
		ServiceName   string `toml:"service_name"`
		EncryptionKey string `toml:"encryption_key"`
	} `toml:"keyring"`

	Logger struct {
		LoggerLevel int `toml:"level"`
	} `toml:"logger"`
}

// Credential represents a simple structure for storing user credentials.
type Credential struct {
	Username string
	Password string
}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	HTMLURL    string `json:"html_url"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}
