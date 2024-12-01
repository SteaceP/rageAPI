package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SteaceP/coderage/internal/models"
	"github.com/SteaceP/coderage/pkg/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// CreateCommentRequest represents the structure for creating a new comment
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// CreateComment handles creating a new comment on a post
func CreateComment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("user_id").(uint)

	// Get post ID from URL
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["postId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get database from context
	db := r.Context().Value("db").(*gorm.DB)

	// Verify post exists
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Create comment
	comment := models.Comment{
		Content: req.Content,
		UserID:  userID,
		PostID:  uint(postID),
	}

	if err := db.Create(&comment).Error; err != nil {
		http.Error(w, "Comment creation failed", http.StatusInternalServerError)
		return
	}

	// Preload user for the response
	if err := db.Preload("User").First(&comment, comment.ID).Error; err != nil {
		http.Error(w, "Failed to fetch comment details", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"message": "Comment created successfully",
		"comment": map[string]interface{}{
			"id":      utils.UintToString(comment.ID),
			"content": comment.Content,
			"user": map[string]string{
				"id":       utils.UintToString(comment.User.ID),
				"username": comment.User.Username,
			},
			"post_id": utils.UintToString(comment.PostID),
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListComments retrieves comments for a specific post
func ListComments(w http.ResponseWriter, r *http.Request) {
	// Get post ID from URL
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["postId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get database from context
	db := r.Context().Value("db").(*gorm.DB)

	// Parse query parameters for pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Verify post exists
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Fetch comments with pagination and preload user
	var comments []models.Comment
	var totalCount int64
	if err := db.Model(&models.Comment{}).Where("post_id = ?", postID).Count(&totalCount).Error; err != nil {
		http.Error(w, "Failed to count comments", http.StatusInternalServerError)
		return
	}

	if err := db.Preload("User").Where("post_id = ?", postID).Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"comments": comments,
		"pagination": map[string]interface{}{
			"total_comments": totalCount,
			"page":           page,
			"limit":          limit,
			"total_pages":    (totalCount + int64(limit) - 1) / int64(limit),
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
