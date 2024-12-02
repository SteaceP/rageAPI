package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/SteaceP/coderage/models"
	"github.com/SteaceP/coderage/types"
	"github.com/SteaceP/coderage/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(types.KeyUserID).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Get database from context
	db, ok := r.Context().Value(types.KeyDB).(*gorm.DB)
	if !ok || db == nil {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}
	// Check user role
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Check if user is an admin
	if user.Role != types.RoleAdmin {
		http.Error(w, "Forbidden: Only admins can create posts", http.StatusForbidden)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" || req.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Create post
	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
	}

	if err := db.Create(&post).Error; err != nil {
		http.Error(w, "Post creation failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"message": "Post created successfully",
		"post": map[string]interface{}{
			"id":      post.ID,
			"title":   post.Title,
			"content": post.Content,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Response encoding failed", http.StatusInternalServerError)
	}
}

func ListPosts(w http.ResponseWriter, r *http.Request) {
	// Get database from context
	db, ok := r.Context().Value(types.KeyDB).(*gorm.DB)
	if !ok {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}

	// Parse query parameters for pagination
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Fetch posts with pagination and preload user
	var posts []models.Post
	var totalCount int64
	if err := db.Model(&models.Post{}).Count(&totalCount).Error; err != nil {
		http.Error(w, "Failed to count posts", http.StatusInternalServerError)
		return
	}

	if err := db.Preload("User").Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"posts": posts,
		"pagination": map[string]interface{}{
			"total_posts": totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetPost retrieves a single post by ID, including the user and comments.
func GetPost(w http.ResponseWriter, r *http.Request) {
	// Get database from context
	db := r.Context().Value(types.KeyDB).(*gorm.DB)
	if db == nil {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}

	// Get post ID from URL
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars[types.IDField], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Fetch post with user
	var post models.Post
	if err := db.Preload("User").Preload("Comments").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value(types.KeyUserID).(uint)

	// Get post ID from URL
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars[types.IDField], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get database from context
	db := r.Context().Value(types.KeyDB).(*gorm.DB)
	if db == nil {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}

	// Find existing post
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}

	// Check if the user owns the post
	if post.UserID != userID {
		http.Error(w, "Unauthorized to update this post", http.StatusForbidden)
		return
	}

	// Parse update request
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update post
	post.Title = req.Title
	post.Content = req.Content
	if err := db.Save(&post).Error; err != nil {
		http.Error(w, "Post update failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"message": "Post updated successfully",
		"post": map[string]string{
			"id":      utils.UintToString(post.ID),
			"title":   post.Title,
			"content": post.Content,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value(types.KeyUserID).(uint)

	// Get post ID from URL
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars[types.IDField], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get database from context
	db, ok := r.Context().Value(types.KeyDB).(*gorm.DB)
	if !ok {
		http.Error(w, "Internal Server Error (Database unavailable)", http.StatusInternalServerError)
		return
	}

	// Find existing post
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}

	// Check if the user owns the post
	if post.UserID != userID {
		http.Error(w, "Unauthorized to delete this post", http.StatusForbidden)
		return
	}

	// Delete post
	if err := db.Delete(&post).Error; err != nil {
		http.Error(w, "Post deletion failed", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post deleted successfully",
	})
}
