package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SteaceP/coderage/internal/models"
	"github.com/SteaceP/coderage/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUserRequest represents the structure for user registration
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUser handles user registration by creating a new user in the database and
// generating a JSON Web Token to be used for authentication in subsequent
// requests.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Password hashing failed", http.StatusInternalServerError)
		return
	}

	// Create user model
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Get database from context
	dbValue := r.Context().Value("db")
	if dbValue == nil {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}
	db, ok := dbValue.(*gorm.DB)
	if !ok {
		http.Error(w, "Invalid database type", http.StatusInternalServerError)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Create user
	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"message": "User created successfully",
		"token":   token,
		"user": map[string]string{
			"id":       utils.UintToString(user.ID),
			"username": user.Username,
			"email":    user.Email,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUserProfile retrieves a user's profile details
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userIDValue := r.Context().Value("user_id")
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
	dbValue := r.Context().Value("db")
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
