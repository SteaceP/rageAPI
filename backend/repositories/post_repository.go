package repositories

import (
	"strings"

	"github.com/SteaceP/coderage/models"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	// Generate slug if not provided
	if post.Slug == "" {
		post.Slug = generateSlug(post.Title)
	}
	return r.db.Create(post).Error
}

func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.
		Preload("User").
		Preload("Comments").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := r.db.
		Where("slug = ?", slug).
		Preload("User").
		Preload("Comments").
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) List(page, pageSize int, filters map[string]interface{}) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	// Base query
	query := r.db.Model(&models.Post{})

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if tags, ok := filters["tags"].([]string); ok && len(tags) > 0 {
		query = query.Where("? = ANY(tags)", tags[0])
	}

	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	query.Count(&total)

	// Fetch paginated posts
	err := query.
		Preload("User").
		Order("published_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) Update(post *models.Post) error {
	// Update slug if title changes
	if post.Title != "" {
		post.Slug = generateSlug(post.Title)
	}
	return r.db.Save(post).Error
}

func (r *PostRepository) Delete(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *PostRepository) IncrementViewCount(postID uint) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PostRepository) UpdateCommentCount(postID uint, increment bool) error {
	var operation string
	if increment {
		operation = "comment_count + 1"
	} else {
		operation = "comment_count - 1"
	}

	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr(operation)).Error
}

// Helper function to generate URL-friendly slug
func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace non-alphanumeric characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove any non-alphanumeric or hyphen characters
	var result []rune
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}

	return string(result)
}
