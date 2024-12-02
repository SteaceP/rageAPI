package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SteaceP/coderage/models"
	"github.com/SteaceP/coderage/types"
	"github.com/SteaceP/coderage/utils"

	"gorm.io/gorm"
)

// GetUserProfile retrieves a user's profile details
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userIDValue := r.Context().Value(types.KeyUserID)
	if userIDValue == nil {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Invalid user ID type", http.StatusInternalServerError)
		return
	}

	// Get database from context
	dbValue := r.Context().Value(types.KeyDB)
	if dbValue == nil {
		http.Error(w, "Database not found in context", http.StatusInternalServerError)
		return
	}
	db, ok := dbValue.(*gorm.DB)
	if !ok {
		http.Error(w, "Invalid database type", http.StatusInternalServerError)
		return
	}

	// Find user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare response
	response := map[string]string{
		"id":       utils.UintToString(user.ID),
		"username": user.Username,
		"email":    user.Email,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
