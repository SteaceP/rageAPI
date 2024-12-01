package services

import (
	"errors"
	"time"

	"github.com/SteaceP/coderage/internal/models"
	"github.com/SteaceP/coderage/internal/repositories"
	"go.uber.org/zap"
)

type PostService struct {
	postRepo    *repositories.PostRepository
	userRepo    *repositories.UserRepository
	commentRepo *repositories.CommentRepository
	logger      *zap.Logger
}

// NewPostService returns a new instance of PostService, which is used to manage the
// lifecycle of posts.
//
// The returned instance is backed by the provided PostRepository, UserRepository,
// CommentRepository, and logger.
func NewPostService(
	postRepo *repositories.PostRepository,
	userRepo *repositories.UserRepository,
	commentRepo *repositories.CommentRepository,
	logger *zap.Logger,
) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
		logger:      logger,
	}
}

// CreatePost creates a new post in the database.
//
// It first validates the post's fields, and returns an error if any of them are
// invalid. It then sets the post's published date to the current time if it is
// zero. It also ensures that the post is associated with a valid user.
//
// Finally, it creates the post in the database and returns an error if that
// fails.
func (s *PostService) CreatePost(post *models.Post) error {
	// Validate post
	if err := validatePost(post); err != nil {
		return err
	}

	// Set published date if not set
	if post.PublishedAt.IsZero() {
		post.PublishedAt = time.Now()
	}

	// Ensure post is associated with a valid user
	_, err := s.userRepo.FindByID(post.UserID)
	if err != nil {
		return errors.New("invalid user")
	}

	return s.postRepo.Create(post)
}

// GetPost retrieves a post from the database using either its ID or slug.
//
// The method accepts an identifier which can be either a uint (representing the post ID)
// or a string (representing the post slug). It attempts to find the post using the given
// identifier and returns the post along with an error, if any.
//
// If the identifier is not of a valid type (neither uint nor string), it returns an error
// indicating the invalid identifier type.
//
// Upon successfully retrieving the post, it increments the post's view count. Any error
// encountered during the increment of the view count is logged but does not affect the
// retrieval process.
func (s *PostService) GetPost(identifier interface{}) (*models.Post, error) {
	var post *models.Post
	var err error

	switch v := identifier.(type) {
	case uint:
		post, err = s.postRepo.FindByID(v)
	case string:
		post, err = s.postRepo.FindBySlug(v)
	default:
		return nil, errors.New("invalid identifier type")
	}

	if err != nil {
		return nil, err
	}

	// Log any error from incrementing view count
	if err := s.postRepo.IncrementViewCount(post.ID); err != nil {
		s.logger.Error("Failed to increment view count",
			zap.Uint("post_id", post.ID),
			zap.Error(err),
		)
	}

	return post, nil
}

// ListPosts retrieves posts with pagination and preload user
//
// It expects the following query parameters:
// page: int - Page number to retrieve (default: 1)
// limit: int - Number of posts to retrieve per page (default: 10, max: 100)
//
// The response will be a JSON object with the following structure:
//
//	{
//	    "posts": [
//	        {
//	            "id": "<post id>",
//	            "title": "<post title>",
//	            "content": "<post content>",
//	            "user": {
//	                "id": "<user id>",
//	                "username": "<user username>",
//	                "email": "<user email>"
//	            }
//	        },
//	        ...
//	    ],
//	    "pagination": {
//	        "total_posts": <total count of posts>,
//	        "page": <current page number>,
//	        "limit": <limit of posts per page>,
//	        "total_pages": <total number of pages>
//	    }
//	}
func (s *PostService) ListPosts(page, pageSize int, filters map[string]interface{}) ([]models.Post, int64, error) {
	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.postRepo.List(page, pageSize, filters)
}

// UpdatePost updates a post in the database.
//
// It first validates the post's fields, and returns an error if any of them are
// invalid. It then ensures that the post exists in the database, and returns an
// error if it does not. It then updates the post's fields and saves it back to the
// database, returning an error if that fails.
func (s *PostService) UpdatePost(post *models.Post) error {
	// Validate post
	if err := validatePost(post); err != nil {
		return err
	}

	// Ensure the post exists
	existingPost, err := s.postRepo.FindByID(post.ID)
	if err != nil {
		return errors.New("post not found")
	}

	// Update fields
	existingPost.Title = post.Title
	existingPost.Content = post.Content
	existingPost.Excerpt = post.Excerpt
	existingPost.Status = post.Status
	existingPost.Tags = post.Tags
	existingPost.FeaturedImage = post.FeaturedImage
	existingPost.MetaTitle = post.MetaTitle
	existingPost.MetaDescription = post.MetaDescription

	return s.postRepo.Update(existingPost)
}

// DeletePost removes a post from the database by its ID.
//
// It verifies the existence of the post before attempting to delete it.
// If the post is not found, it returns an error indicating that the post
// was not found. If the post exists, it deletes the post and returns an
// error if the deletion fails.
func (s *PostService) DeletePost(postID uint) error {
	// Check if post exists
	_, err := s.postRepo.FindByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	return s.postRepo.Delete(postID)
}

// AddComment creates a new comment in the database.
//
// It first validates the comment's fields, and returns an error if any of them are
// invalid. It then ensures that the post exists in the database, and returns an
// error if it does not. It then creates the comment in the database and returns an
// error if that fails. Finally, it increments the post's comment count and returns
// an error if that fails.
func (s *PostService) AddComment(comment *models.Comment) error {
	// Validate comment
	if err := validateComment(comment); err != nil {
		return err
	}

	// Ensure post exists
	_, err := s.postRepo.FindByID(comment.PostID)
	if err != nil {
		return errors.New("post not found")
	}

	// Create comment
	if err := s.commentRepo.Create(comment); err != nil {
		return err
	}

	// Update post comment count
	return s.postRepo.UpdateCommentCount(comment.PostID, true)
}

// validatePost validates a post's fields, and returns an error if any of them
// are invalid.
//
// The following are the validation rules:
//
// - The title and content are required.
// - The title must be between 5 and 200 characters long.
func validatePost(post *models.Post) error {
	if post.Title == "" {
		return errors.New("title is required")
	}

	if post.Content == "" {
		return errors.New("content is required")
	}

	if len(post.Title) < 5 || len(post.Title) > 200 {
		return errors.New("title must be between 5 and 200 characters")
	}

	return nil
}

// validateComment validates a comment's fields, and returns an error if any of them
// are invalid.
//
// The following are the validation rules:
//
// - The comment content is required.
// - The comment must be max 500 characters.
func validateComment(comment *models.Comment) error {
	if comment.Content == "" {
		return errors.New("comment content is required")
	}

	if len(comment.Content) > 500 {
		return errors.New("comment must be max 500 characters")
	}

	return nil
}
