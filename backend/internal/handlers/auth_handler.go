package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SteaceP/coderage/internal/models"
	"github.com/SteaceP/coderage/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginRequest represents the structure of a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles user login by verifying the email and password and
// generating a JSON Web Token
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
