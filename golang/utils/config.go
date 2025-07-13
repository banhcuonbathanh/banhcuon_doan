package utils

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var AppConfig Config

type Config struct {
	DatabaseURL     string `mapstructure:"DATABASE_URL"`
	GRPCAddress     string `mapstructure:"GRPCAddress"`
	ServerAddress   string `mapstructure:"ServerAddress"`
	AnthropicAPIKey string `mapstructure:"anthropicAPIKey"`
	AnthropicAPIURL string `mapstructure:"anthropicAPIURL"`
	QuanAnAddress   string `mapstructure:"QuanAnAddress"`
	JwtSecret       string `mapstructure:"JWT_SECRET"`
}

func LoadServer() (*Config, error) {
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")

	// Add multiple config paths for different environments
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app/golang")
	viper.AddConfigPath("/app")
	viper.AddConfigPath("./golang")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Override database URL based on environment
	if os.Getenv("APP_ENV") == "docker" {
		// Construct database URL using environment variables
		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"))
		// Override GRPC address for Docker
		cfg.GRPCAddress = ":50051"
	}

	return &cfg, nil
}
