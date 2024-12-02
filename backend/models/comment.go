package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Content   string    `json:"content" validate:"required,min=1,max=500"`
	UserID    uint      `json:"user_id" validate:"required"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	PostID    uint      `json:"post_id" validate:"required"`
	Post      Post      `json:"post" gorm:"foreignKey:PostID"`
	ParentID  *uint     `json:"parent_id,omitempty"` // For nested comments
	Parent    *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies   []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
	Status    string    `json:"status" validate:"oneof=published hidden deleted" default:"published"`
	LikeCount int       `json:"like_count" gorm:"default:0"`
}

// TableName overrides the table name used by Comment to `comments`
func (Comment) TableName() string {
	return "comments"
}
