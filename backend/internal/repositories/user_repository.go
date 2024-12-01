package repositories

import (
	"time"

	"github.com/SteaceP/coderage/internal/models"
	"github.com/SteaceP/coderage/pkg/utils"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(user *models.User) error {
	// Hash password before storing
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return r.db.Create(user).Error
}

// FindByID finds a user by its ID.
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Posts").
		Preload("Comments").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by its username.
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by its email address.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update saves the changes made to an existing user in the database.
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// UpdatePassword updates a user's password by hashing the given new password and
// storing it in the database. It takes the ID of the user to update and the new
// password as arguments. It returns an error if the update fails.
func (r *UserRepository) UpdatePassword(userID uint, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

// List retrieves users with pagination and filters.
//
// The function takes page and pageSize as parameters, and an optional filters
// map. The filters map can contain the following keys:
//
//   - role: string - Filter users by role. If empty, all roles are returned.
//   - is_active: bool - Filter users by active or inactive status. If false,
//     only inactive users are returned. If true, only active users
//     are returned. If not provided, all users are returned.
//
// The response will be a tuple containing the paginated users, the total count
// of users matching the filters, and an error. If the fetch operation fails,
// the error is not nil.
func (r *UserRepository) List(page, pageSize int, filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Base query
	query := r.db.Model(&models.User{})

	// Apply filters
	if role, ok := filters["role"].(string); ok && role != "" {
		query = query.Where("role = ?", role)
	}

	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}

	// Count total
	query.Count(&total)

	// Fetch paginated users
	err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error

	return users, total, err
}

// UpdateLastLogin updates the last login timestamp for a user by their ID.
func (r *UserRepository) UpdateLastLogin(userID uint) error {
	now := time.Now()
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login", now).Error
}

// VerifyUser updates the verification status of a user in the database.
func (r *UserRepository) VerifyUser(userID uint) error {
	now := time.Now()
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"verified_at": now,
			"is_active":   true,
		}).Error
}

// Delete removes a user from the database by its ID.
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}
