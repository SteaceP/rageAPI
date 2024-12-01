package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title           string    `json:"title" validate:"required,min=5,max=200"`
	Slug            string    `json:"slug" gorm:"uniqueIndex"`
	Content         string    `json:"content" validate:"required"`
	Excerpt         string    `json:"excerpt" validate:"max=500"`
	UserID          uint      `json:"user_id" validate:"required"`
	User            User      `json:"user" gorm:"foreignKey:UserID"`
	Comments        []Comment `json:"comments,omitempty"`
	PublishedAt     time.Time `json:"published_at"`
	Status          string    `json:"status" validate:"oneof=draft published archived" default:"draft"`
	Tags            []string  `json:"tags" gorm:"type:text[]"`
	ViewCount       int       `json:"view_count" gorm:"default:0"`
	LikeCount       int       `json:"like_count" gorm:"default:0"`
	CommentCount    int       `json:"comment_count" gorm:"default:0"`
	FeaturedImage   string    `json:"featured_image,omitempty"`
	MetaTitle       string    `json:"meta_title,omitempty" validate:"max=60"`
	MetaDescription string    `json:"meta_description,omitempty" validate:"max=160"`
}

// TableName overrides the table name used by Post to `posts`
func (Post) TableName() string {
	return "posts"
}
