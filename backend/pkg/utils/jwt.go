package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

// GenerateJWTToken generates a JSON Web Token (JWT) containing the given user ID.
// The token's expiration time is configured using the "jwt.expiration" configuration
// key. If the key is not set, the token will expire after 24 hours. The JWT secret
// is configured using the "jwt.secret" key. If the key is not set, the function
// returns an error.
func GenerateJWTToken(userID uint) (string, error) {
	// Get JWT secret from configuration
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		return "", fmt.Errorf("JWT secret is not configured")
	}

	// Get JWT expiration time from configuration
	expiration := viper.GetInt("jwt.expiration")
	if expiration == 0 {
		// Use a default value if the configuration is missing or invalid
		expiration = 24 * 60 * 60 // 24 hours in seconds
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(expiration) * time.Second).Unix(),
	})

	// Sign and get the complete encoded token as a string
	return token.SignedString([]byte(secret))
}

func ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	// Get JWT secret from configuration
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		return nil, fmt.Errorf("JWT secret is not configured")
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token has expired")
			}
		}
		return token, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// UintToString safely converts a uint to a string.
func UintToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
