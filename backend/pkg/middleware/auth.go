package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/SteaceP/coderage/pkg/utils"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt"
)

type contextKey string // Define a custom type for context keys

const (
	keyUserID contextKey = "user_id"
	keyDB     contextKey = "db"
)

func AuthMiddleware(db *gorm.DB) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check for authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization token", http.StatusUnauthorized)
				return
			}

			// Validate token format
			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			// Validate token
			token, err := utils.ValidateJWTToken(bearerToken[1])
			if err != nil || token == nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Validate claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Validate user ID
			userIDFloat, ok := claims["user_id"]
			if !ok {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}
			if userIDInt, ok := userIDFloat.(int64); ok {
				userID := uint(userIDInt)
				if userID == 0 {
					http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
					return
				}

				// Check database connection
				if db == nil {
					http.Error(w, "Database connection is unavailable", http.StatusInternalServerError)
					return
				}

				// Attach user ID to request context
				ctx := context.WithValue(r.Context(), keyUserID, userID)
				ctx = context.WithValue(ctx, keyDB, db)

				// Call next handler
				next.ServeHTTP(w, r.WithContext(ctx))
			} else if userIDFloat64, ok := userIDFloat.(float64); ok {
				userID := uint(userIDFloat64)
				if userID == 0 {
					http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
					return
				}

				// Check database connection
				if db == nil {
					http.Error(w, "Database connection is unavailable", http.StatusInternalServerError)
					return
				}

				// Attach user ID to request context
				ctx := context.WithValue(r.Context(), keyUserID, userID)
				ctx = context.WithValue(ctx, keyDB, db)

				// Call next handler
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}
		}
	}
}
