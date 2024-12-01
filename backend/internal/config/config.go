package config

import (
	"log"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Default configurations
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("database.type", "postgres")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", 24)
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("cors.allowed_origins", []string{"*"})

	// Read config
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}
}
