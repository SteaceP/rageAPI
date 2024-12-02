package services

import (
	"errors"
	"time"

	"github.com/SteaceP/coderage/models"
	"github.com/SteaceP/coderage/repositories"
	"github.com/SteaceP/coderage/utils"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// NewAuthService creates a new instance of AuthService with the provided UserRepository.
func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Register creates a new user in the database.
func (s *AuthService) Register(user *models.User) error {
	// Validate user input
	if err := utils.ValidateStruct(user); len(err) > 0 {
		return errors.New(err[0])
	}

	// Check if username or email already exists
	_, errUsername := s.userRepo.FindByUsername(user.Username)
	if errUsername == nil {
		return errors.New("username already exists")
	}

	_, errEmail := s.userRepo.FindByEmail(user.Email)
	if errEmail == nil {
		return errors.New("email already exists")
	}

	// Create user
	return s.userRepo.Create(user)
}

// Login logs in a user by verifying their email and password.
func (s *AuthService) Login(email, password string) (*TokenDetails, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.CreateTokenPair(user)
}

// CreateTokenPair creates a pair of access and refresh tokens for the given user.
func (s *AuthService) CreateTokenPair(user *models.User) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 24).Unix()
	td.AccessUUID = uuid.New().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUUID = uuid.New().String()

	// Access Token
	atClaims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"authorized": true,
		"exp":        td.AtExpires,
		"uuid":       td.AccessUUID,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	var err error
	td.AccessToken, err = at.SignedString([]byte(viper.GetString("jwt.secret")))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	rtClaims := jwt.MapClaims{
		"user_id": user.ID,
		"uuid":    td.RefreshUUID,
		"exp":     td.RtExpires,
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(viper.GetString("jwt.secret")))
	if err != nil {
		return nil, err
	}

	return td, nil
}

// RefreshToken verifies the given refresh token and generates a new pair of access and refresh tokens.
func (s *AuthService) RefreshToken(refreshToken string) (*TokenDetails, error) {
	// Verify refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(viper.GetString("jwt.secret")), nil
	})

	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Find user
	userID := uint(claims["user_id"].(float64))
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new token pair
	return s.CreateTokenPair(user)
}
