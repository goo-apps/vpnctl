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
	Push     string `json:"push"`
	YFlag    string `json:"y_flag"`
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
	} `toml:"application"`
}

// Credential represents a simple structure for storing user credentials.
type Credential struct {
	Username string
	Password string
}
