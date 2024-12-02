package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SteaceP/coderage/models"
	"github.com/SteaceP/coderage/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	var user models.User

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user by email
	db, ok := r.Context().Value("db").(*gorm.DB)
	if !ok {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]string{
		"token":   token,
		"message": "Login successful",
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
