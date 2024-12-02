package repositories

import (
	"github.com/SteaceP/coderage/models"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository returns a new instance of CommentRepository.
//
// The returned instance is backed by the provided Gorm database connection.
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment in the database.
//
// The comment must not have an ID or else an error will be returned. The
// comment's content, user ID, and post ID must be set. If the comment is
// successfully created, this function will return nil. Otherwise, it will
// return an error describing the reason the creation failed.
func (r *CommentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// FindByID finds a comment by its ID.
//
// It returns the comment and an error. If the comment is not found, the error
// is gorm.ErrRecordNotFound.
func (r *CommentRepository) FindByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// FindByPostID retrieves comments for the given post ID, with pagination.
//
// The comments are ordered by their creation time in descending order.
//
// It returns the comments, the total count of comments and an error.
func (r *CommentRepository) FindByPostID(postID uint, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// Count total comments
	r.db.Model(&models.Comment{}).Where("post_id = ?", postID).Count(&total)

	// Fetch paginated comments
	err := r.db.Where("post_id = ?", postID).
		Preload("User").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error

	return comments, total, err
}

// Update updates an existing comment in the database.
//
// The comment must have an ID or else an error will be returned. The comment's
// content and user ID may be updated. If the comment is successfully updated,
// this function will return nil. Otherwise, it will return an error describing
// the reason the update failed.
func (r *CommentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

// Delete removes a comment from the database by its ID.
//
// Returns an error if the deletion fails.
func (r *CommentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

// FindReplies finds all replies to the given comment.
//
// The replies are ordered by their creation time in ascending order.
func (r *CommentRepository) FindReplies(commentID uint) ([]models.Comment, error) {
	var replies []models.Comment
	err := r.db.Where("parent_id = ?", commentID).
		Preload("User").
		Order("created_at ASC").
		Find(&replies).Error
	return replies, err
}

// UpdateLikeCount updates the like count for a comment. If increment is true, the count is
// incremented by one, otherwise it is decremented by one.
func (r *CommentRepository) UpdateLikeCount(commentID uint, increment bool) error {
	var operation string
	if increment {
		operation = "like_count + 1"
	} else {
		operation = "like_count - 1"
	}

	return r.db.Model(&models.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr(operation)).Error
}
