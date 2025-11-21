package pkg

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	SecretId string
	Secret   string
	BaseUrl  string
}

func GetConfig() (Config, error) {
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
			return Config{}, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Load values
	cfg := Config{
		SecretId: viper.GetString("secret_id"),
		Secret:   viper.GetString("secret"),
		BaseUrl:  viper.GetString("base_url"),
	}

	// Set default BaseUrl if not provided
	if cfg.BaseUrl == "" {
		cfg.BaseUrl = "http://localhost:9670"
	}

	// Validate required credentials
	if cfg.SecretId == "" || cfg.Secret == "" {
		return Config{}, fmt.Errorf("RS_SECRET_ID and RS_SECRET must be set (via environment variables or config file)")
	}

	return cfg, nil
}

func GetSecretId() string {
	cfg, err := GetConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}
	return cfg.SecretId
}

func GetSecret() string {
	cfg, err := GetConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}
	return cfg.Secret
}

func GetBaseURL() string {
	cfg, err := GetConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}
	return cfg.BaseUrl
}
