package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	// RsSecretID is the Reysys secret ID from RS_SECRET_ID env var
	RsSecretID string

	// RsSecret is the Reysys secret from RS_SECRET env var
	RsSecret string

	// BaseURL is the base URL for API requests
	BaseURL string
)

// Init initializes the configuration using Viper
func Init() error {
	// Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.reysys")
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("RS")
	viper.AutomaticEnv()

	// Bind specific environment variables
	viper.BindEnv("secret_id", "RS_SECRET_ID")
	viper.BindEnv("secret", "RS_SECRET")
	viper.BindEnv("base_url", "RS_BASE_URL")

	// Read config file (optional - won't error if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Load values into global variables
	RsSecretID = viper.GetString("secret_id")
	RsSecret = viper.GetString("secret")
	BaseURL = viper.GetString("base_url")

	// Set default BaseURL if not provided
	if BaseURL == "" {
		BaseURL = "http://localhost:9670"
	}

	// Validate required credentials
	if RsSecretID == "" || RsSecret == "" {
		return fmt.Errorf("RS_SECRET_ID and RS_SECRET must be set (via environment variables or config file)")
	}

	return nil
}

// GetSecretID returns the secret ID
func GetSecretID() string {
	return viper.GetString("secret_id")
}

// GetSecret returns the secret
func GetSecret() string {
	return viper.GetString("secret")
}
