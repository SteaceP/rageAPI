package services

import (
	"errors"

	"github.com/SteaceP/coderage/internal/models"
	"github.com/SteaceP/coderage/internal/repositories"
	"github.com/SteaceP/coderage/pkg/utils"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService returns a new instance of UserService with the provided UserRepository.
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserProfile retrieves a user's profile by their ID.
func (s *UserService) GetUserProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

// UpdateProfile updates a user's profile information.
func (s *UserService) UpdateProfile(user *models.User) error {
	// Validate input
	if err := validateUserUpdate(user); err != nil {
		return err
	}

	// Fetch existing user
	existingUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return errors.New("user not found")
	}

	// Update allowed fields
	existingUser.FirstName = user.FirstName
	existingUser.LastName = user.LastName
	existingUser.Bio = user.Bio
	existingUser.ProfilePicture = user.ProfilePicture
	existingUser.TwitterHandle = user.TwitterHandle
	existingUser.LinkedInProfile = user.LinkedInProfile
	existingUser.PersonalWebsite = user.PersonalWebsite

	return s.userRepo.Update(existingUser)
}

// ChangePassword updates a user's password in the database.
func (s *UserService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	// Fetch user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if !utils.CheckPasswordHash(currentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Update password
	return s.userRepo.UpdatePassword(userID, newPassword)
}

// ListUsers retrieves users with pagination and filters.
func (s *UserService) ListUsers(page, pageSize int, filters map[string]interface{}) ([]models.User, int64, error) {
	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.userRepo.List(page, pageSize, filters)
}

func (s *UserService) VerifyUser(userID uint) error {
	return s.userRepo.VerifyUser(userID)
}

// DeactivateUser sets a user's IsActive field to false, deactivating the account.
func (s *UserService) DeactivateUser(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	user.IsActive = false
	return s.userRepo.Update(user)
}

// DeleteUser removes a user from the database by its ID.
func (s *UserService) DeleteUser(userID uint) error {
	return s.userRepo.Delete(userID)
}

// validateUserUpdate validates a user's update data.
func validateUserUpdate(user *models.User) error {
	// Validate first name and last name
	if user.FirstName != "" && (len(user.FirstName) < 2 || len(user.FirstName) > 50) {
		return errors.New("first name must be between 2 and 50 characters")
	}

	if user.LastName != "" && (len(user.LastName) < 2 || len(user.LastName) > 50) {
		return errors.New("last name must be between 2 and 50 characters")
	}

	// Validate bio
	if user.Bio != "" && len(user.Bio) > 500 {
		return errors.New("bio cannot exceed 500 characters")
	}

	return nil
}

// validatePassword validates a password according to the following rules:
func validatePassword(password string) error {
	// Check password complexity
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Additional complexity checks can be added here
	// For example, checking for uppercase, lowercase, numbers, special characters
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case (char >= '!' && char <= '/') || (char >= ':' && char <= '@'):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("password must include uppercase, lowercase, number, and special character")
	}

	return nil
}
