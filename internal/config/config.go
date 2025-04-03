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
	DbUrl                string
	DbUrlDev             string
	RMQAddress           string
	RMQAddressDev        string
	Production           string
	Port                 string
	TokenSecret          string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	WeatherApiKey        string
}

// LoadConfig loads environment variables from the specified file paths (if provided) or defaults to `.env`.
// It returns a Config struct populated with the required environment variables.
func LoadConfig(path ...string) (*Config, error) {
	// Default to .env if no path is provided
	envFile := ".env"
	if len(path) > 0 {
		envFile = path[0]
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
		DbUrl:                os.Getenv("DB_URL"),
		DbUrlDev:             os.Getenv("DB_URL_DEV"),
		Production:           os.Getenv("PRODUCTION"),
		Port:                 os.Getenv("PORT"),
		TokenSecret:          os.Getenv("TOKEN_SECRET"),
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,
		RMQAddress:           os.Getenv("RMQ_ADDRESS"),
		RMQAddressDev:        os.Getenv("RMQ_ADDRESS_DEV"),
		WeatherApiKey:        os.Getenv("WEATHER_API_KEY"),
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
	if config.DbUrl == "" {
		return errors.New("missing required environment variable: DB_URL")
	}
	if config.WeatherApiKey == "" {
		return errors.New("missing required environment variable: WEATHER_API_KEY")
	}
	if config.AccessTokenDuration == 0 {
		return errors.New("missing or invalid required environment variable: ACCESS_TOKEN_DURATION")
	}

	if config.RefreshTokenDuration == 0 {
		return errors.New("missing or invalid required environment variable: REFRESH_TOKEN_DURATION")
	}

	if config.RMQAddress == "" {
		return errors.New("missing or invalid required environment variable: RMQ_ADDRESS")
	}

	if config.RMQAddressDev == "" {
		return errors.New("missing or invalid required environment variable: RMQ_ADDRESS_DEV")
	}

	if config.Port == "" {
		return errors.New("missing required environment variable: PORT")
	}

	return nil
}
