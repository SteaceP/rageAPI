package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username       string     `json:"username" gorm:"uniqueIndex" validate:"required,min=3,max=50"`
	Email          string     `json:"email" gorm:"uniqueIndex" validate:"required,email"`
	Password       string     `json:"-"` // Stored hash, never returned in JSON
	FirstName      string     `json:"first_name,omitempty" validate:"max=50"`
	LastName       string     `json:"last_name,omitempty" validate:"max=50"`
	Bio            string     `json:"bio,omitempty" validate:"max=500"`
	ProfilePicture string     `json:"profile_picture,omitempty"`
	Role           string     `json:"role" validate:"oneof=user editor admin" default:"user"`
	LastLogin      *time.Time `json:"last_login,omitempty"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	Posts          []Post     `json:"posts,omitempty"`
	Comments       []Comment  `json:"comments,omitempty"`
	// Social links
	TwitterHandle   string `json:"twitter_handle,omitempty"`
	LinkedInProfile string `json:"linkedin_profile,omitempty"`
	PersonalWebsite string `json:"personal_website,omitempty"`
}

// TableName overrides the table name used by User to `users`
func (User) TableName() string {
	return "users"
}
