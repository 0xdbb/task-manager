package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the configuration values from environment variables.
type Config struct {
	DB_URL                 string
	DB_URL_DEV             string
	PRODUCTION             string
	PORT                   string
	TOKEN_SECRET           string
	ACCESS_TOKEN_DURATION  time.Duration
	REFRESH_TOKEN_DURATION time.Duration
	ALLOWED_ORIGINS        string
}

// LoadConfig loads environment variables from the specified file paths (if provided) or defaults to `.env`.
// It returns a Config struct populated with the required environment variables.
func LoadConfig(paths ...string) (*Config, error) {
	// Default to .env if no path is provided
	envFile := ".env"
	if len(paths) > 0 {
		envFile = paths[0]
	}

	// Load environment variables from the .env file
	err := godotenv.Load(envFile)
	if err != nil {
		return nil, fmt.Errorf("could not load config file %s: %w", envFile, err)
	}

	// Parse ACCESS_TOKEN_DURATION
	accessTokenDurationStr := os.Getenv("ACCESS_TOKEN_DURATION")
	refreshTokenDurationStr := os.Getenv("REFRESH_TOKEN_DURATION")
	accessTokenDuration, err := time.ParseDuration(accessTokenDurationStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse access token duration: %w", err)
	}

	refreshTokenDuration, err := time.ParseDuration(refreshTokenDurationStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse access token duration: %w", err)
	}

	// Populate the Config struct
	config := &Config{
		DB_URL:                 os.Getenv("DB_URL"),
		DB_URL_DEV:             os.Getenv("DB_URL_DEV"),
		PRODUCTION:             os.Getenv("PRODUCTION"),
		PORT:                   os.Getenv("PORT"),
		TOKEN_SECRET:           os.Getenv("TOKEN_SECRET"),
		ACCESS_TOKEN_DURATION:  accessTokenDuration,
		REFRESH_TOKEN_DURATION: refreshTokenDuration,
		ALLOWED_ORIGINS:        os.Getenv("ALLOWED_ORIGINS"),
	}

	// Validate required environment variables
	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig checks that all required environment variables are set.
func validateConfig(config *Config) error {
	if config.DB_URL == "" {
		return errors.New("missing required environment variable: DB_URL")
	}
	if config.ACCESS_TOKEN_DURATION == 0 {
		return errors.New("missing or invalid required environment variable: ACCESS_TOKEN_DURATION")
	}

	if config.REFRESH_TOKEN_DURATION == 0 {
		return errors.New("missing or invalid required environment variable: REFRESH_TOKEN_DURATION")
	}

	if config.PORT == "" {
		return errors.New("missing required environment variable: PORT")
	}

	if config.ALLOWED_ORIGINS == "" {
		return errors.New("missing required environment variable: ALLOWED_ORIGINS")
	}

	return nil
}
