package middleware

import (
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

// ConfigureCORS sets up CORS middleware with configuration from Viper
func ConfigureCORS() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   viper.GetStringSlice("cors.allowed_origins"),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		// Optional: Add debug logging for CORS errors
		Debug: viper.GetBool("cors.debug"),
	})
}
