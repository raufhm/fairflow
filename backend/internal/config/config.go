package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	// Server
	Port int

	// Security
	JWTSecret string

	// Database
	DatabaseURL string

	// Storage
	BackupDir string
	DataDir   string

	// Application
	Environment string // development, staging, production
	LogLevel    string
}

var globalConfig *Config

// Load loads configuration from .env file and environment variables
// Configuration is loaded once and cached for subsequent calls
// Priority: Environment variables > .env file > defaults
func Load() *Config {
	// Return cached config if already loaded
	if globalConfig != nil {
		return globalConfig
	}

	// Allow environment variable overrides (highest priority)
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("PORT", 3000)
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("DATA_DIR", "./data")

	// Configure Viper for .env file loading
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")                // Look in current directory
	viper.AddConfigPath("..")               // Look in parent directory (for backend/)
	viper.AddConfigPath("../..")            // Look two levels up

	// Try to load .env file (optional for local development)
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found, using environment variables and defaults")
	} else {
		log.Printf("Loaded configuration from: %s", viper.ConfigFileUsed())
	}

	// Build config struct
	cfg := &Config{
		Port:        viper.GetInt("PORT"),
		Environment: viper.GetString("ENVIRONMENT"),
		LogLevel:    viper.GetString("LOG_LEVEL"),
		JWTSecret:   viper.GetString("JWT_SECRET"),
		DatabaseURL: viper.GetString("DATABASE_URL"),
		DataDir:     viper.GetString("DATA_DIR"),
	}

	// Set backup directory with fallback
	if backupDir := viper.GetString("BACKUP_DIR"); backupDir != "" {
		cfg.BackupDir = backupDir
	} else {
		cfg.BackupDir = cfg.DataDir + "/backups"
	}

	// Validate required fields
	if cfg.JWTSecret == "" {
		panic("JWT_SECRET is required. Set it in .env.development or as environment variable")
	}

	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required. Set it in .env.development or as environment variable")
	}

	// Cache the config
	globalConfig = cfg

	return cfg
}

// GetConfig returns the cached configuration
// Panics if Load() hasn't been called yet
func GetConfig() *Config {
	if globalConfig == nil {
		panic("Configuration not loaded. Call Load() first")
	}
	return globalConfig
}

// Reload forces a fresh load of configuration (useful for testing)
func Reload() *Config {
	globalConfig = nil
	return Load()
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535")
	}
	return nil
}
