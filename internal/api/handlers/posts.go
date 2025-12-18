package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yourusername/college-event-backend/internal/models"
)

// PostsHandler handles post-related requests
type PostsHandler struct {
	db *sql.DB
}

// NewPostsHandler creates a new posts handler
func NewPostsHandler(db *sql.DB) *PostsHandler {
	return &PostsHandler{db: db}
}

// CreatePost creates a new post (admin-only)
// POST /api/v1/admin/posts
func (h *PostsHandler) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	creatorID := userID.(uuid.UUID)

	// Validate content type matches URLs
	if req.ContentType == models.ContentTypeImage && req.ImageURL == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("image_url required for image content"),
		})
		return
	}
	if req.ContentType == models.ContentTypeVideo && (req.VideoURL == nil || req.ThumbnailURL == nil) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("video_url and thumbnail_url required for video content"),
		})
		return
	}

	// Insert post
	query := `
		INSERT INTO posts (
			created_by, club_id, house_id, content_type, 
			image_url, video_url, thumbnail_url, duration_seconds,
			description, hashtags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at, storage_class
	`

	var post models.Post
	post.CreatedBy = creatorID
	post.ClubID = req.ClubID
	post.HouseID = req.HouseID
	post.ContentType = req.ContentType
	post.ImageURL = req.ImageURL
	post.VideoURL = req.VideoURL
	post.ThumbnailURL = req.ThumbnailURL
	post.DurationSecs = req.DurationSecs
	post.Description = req.Description
	post.Hashtags = req.Hashtags
	if post.Hashtags == nil {
		post.Hashtags = []string{}
	}

	err := h.db.QueryRow(
		query,
		creatorID, req.ClubID, req.HouseID, req.ContentType,
		req.ImageURL, req.VideoURL, req.ThumbnailURL, req.DurationSecs,
		req.Description, pq.Array(post.Hashtags),
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt, &post.StorageClass)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to create post: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Post created successfully",
		Data:    post,
	})
}

// ListPosts lists posts with pagination
// GET /api/v1/posts
func (h *PostsHandler) ListPosts(c *gin.Context) {
	var query models.ListPostsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid query parameters"),
		})
		return
	}

	// Default pagination
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	offset := (query.Page - 1) * query.PageSize

	// Build query with filters
	whereClause := "WHERE p.deleted_at IS NULL"
	args := []interface{}{}
	argCount := 1

	if query.Hashtag != nil {
		whereClause += " AND $" + strconv.Itoa(argCount) + " = ANY(p.hashtags)"
		args = append(args, *query.Hashtag)
		argCount++
	}

	if query.ClubID != nil {
		whereClause += " AND p.club_id = $" + strconv.Itoa(argCount)
		args = append(args, *query.ClubID)
		argCount++
	}

	if query.HouseID != nil {
		whereClause += " AND p.house_id = $" + strconv.Itoa(argCount)
		args = append(args, *query.HouseID)
		argCount++
	}

	if query.Search != nil {
		searchTerm := "%" + strings.ToLower(*query.Search) + "%"
		whereClause += " AND LOWER(p.description) LIKE $" + strconv.Itoa(argCount)
		args = append(args, searchTerm)
		argCount++
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM posts p " + whereClause
	err := h.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to get post count"),
		})
		return
	}

	// Get posts
	postsQuery := `
		SELECT 
			p.id, p.created_by, p.club_id, p.house_id,
			p.content_type, p.image_url, p.video_url, p.thumbnail_url, p.duration_seconds,
			p.description, p.hashtags, p.created_at, p.updated_at,
			p.archived_at, p.storage_class,
			p.like_count, p.comment_count, p.share_count, p.view_count,
			u.id, u.full_name, u.avatar_url, u.role
		FROM posts p
		JOIN users u ON p.created_by = u.id
		` + whereClause + `
		ORDER BY p.created_at DESC
		LIMIT $` + strconv.Itoa(argCount) + ` OFFSET $` + strconv.Itoa(argCount+1)

	args = append(args, query.PageSize, offset)

	rows, err := h.db.Query(postsQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch posts"),
		})
		return
	}
	defer rows.Close()

	posts := []models.PostResponse{}
	for rows.Next() {
		var pr models.PostResponse
		var hashtags pq.StringArray

		err := rows.Scan(
			&pr.ID, &pr.CreatedBy, &pr.ClubID, &pr.HouseID,
			&pr.ContentType, &pr.ImageURL, &pr.VideoURL, &pr.ThumbnailURL, &pr.DurationSecs,
			&pr.Description, &hashtags, &pr.CreatedAt, &pr.UpdatedAt,
			&pr.ArchivedAt, &pr.StorageClass,
			&pr.LikeCount, &pr.CommentCount, &pr.ShareCount, &pr.ViewCount,
			&pr.Creator.ID, &pr.Creator.FullName, &pr.Creator.AvatarURL, &pr.Creator.Role,
		)
		if err != nil {
			continue
		}

		pr.Hashtags = hashtags

		// Check if current user liked this post
		userID, exists := c.Get("user_id")
		if exists {
			pr.IsLikedByMe = h.checkUserLikedPost(pr.ID, userID.(uuid.UUID))
			pr.IsSharedByMe = h.checkUserSharedPost(pr.ID, userID.(uuid.UUID))
		}

		posts = append(posts, pr)
	}

	totalPages := (totalCount + query.PageSize - 1) / query.PageSize

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PostsListResponse{
			Posts:      posts,
			Page:       query.Page,
			PageSize:   query.PageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	})
}

// GetPost gets a single post by ID with comments
// GET /api/v1/posts/:id
func (h *PostsHandler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	// Get post with creator info
	query := `
		SELECT 
			p.id, p.created_by, p.club_id, p.house_id,
			p.content_type, p.image_url, p.video_url, p.thumbnail_url, p.duration_seconds,
			p.description, p.hashtags, p.created_at, p.updated_at,
			p.archived_at, p.storage_class,
			p.like_count, p.comment_count, p.share_count, p.view_count,
			u.id, u.full_name, u.avatar_url, u.role
		FROM posts p
		JOIN users u ON p.created_by = u.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	var pr models.PostResponse
	var hashtags pq.StringArray

	err = h.db.QueryRow(query, postID).Scan(
		&pr.ID, &pr.CreatedBy, &pr.ClubID, &pr.HouseID,
		&pr.ContentType, &pr.ImageURL, &pr.VideoURL, &pr.ThumbnailURL, &pr.DurationSecs,
		&pr.Description, &hashtags, &pr.CreatedAt, &pr.UpdatedAt,
		&pr.ArchivedAt, &pr.StorageClass,
		&pr.LikeCount, &pr.CommentCount, &pr.ShareCount, &pr.ViewCount,
		&pr.Creator.ID, &pr.Creator.FullName, &pr.Creator.AvatarURL, &pr.Creator.Role,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("Post not found"),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch post"),
		})
		return
	}

	pr.Hashtags = hashtags

	// Check if current user liked/shared
	userID, exists := c.Get("user_id")
	if exists {
		pr.IsLikedByMe = h.checkUserLikedPost(pr.ID, userID.(uuid.UUID))
		pr.IsSharedByMe = h.checkUserSharedPost(pr.ID, userID.(uuid.UUID))
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    pr,
	})
}

// UpdatePost updates a post (admin-only)
// PUT /api/v1/admin/posts/:id
func (h *PostsHandler) UpdatePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request"),
		})
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Description != nil {
		updates = append(updates, "description = $"+strconv.Itoa(argCount))
		args = append(args, *req.Description)
		argCount++
	}

	if req.Hashtags != nil {
		updates = append(updates, "hashtags = $"+strconv.Itoa(argCount))
		args = append(args, pq.Array(*req.Hashtags))
		argCount++
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("No fields to update"),
		})
		return
	}

	query := "UPDATE posts SET " + strings.Join(updates, ", ") +
		" WHERE id = $" + strconv.Itoa(argCount) + " AND deleted_at IS NULL"
	args = append(args, postID)

	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to update post"),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("Post not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Post updated successfully",
	})
}

// DeletePost soft deletes a post (admin-only)
// DELETE /api/v1/admin/posts/:id
func (h *PostsHandler) DeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	query := "UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL"
	result, err := h.db.Exec(query, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to delete post"),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("Post not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Post deleted successfully",
	})
}

// HardDeletePost permanently deletes a post (admin-only)
// DELETE /api/v1/admin/posts/:id/hard
func (h *PostsHandler) HardDeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	query := "SELECT hard_delete_post($1)"
	var success bool
	err = h.db.QueryRow(query, postID).Scan(&success)
	if err != nil || !success {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to delete post"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Post permanently deleted",
	})
}

// ToggleLike toggles like on a post
// POST /api/v1/posts/:id/like
func (h *PostsHandler) ToggleLike(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Check if already liked
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)"
	h.db.QueryRow(checkQuery, postID, uid).Scan(&exists)

	if exists {
		// Unlike
		_, err = h.db.Exec("DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2", postID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   strPtr("Failed to unlike post"),
			})
			return
		}
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Post unliked",
			Data:    gin.H{"liked": false},
		})
	} else {
		// Like
		_, err = h.db.Exec("INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)", postID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   strPtr("Failed to like post"),
			})
			return
		}
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Post liked",
			Data:    gin.H{"liked": true},
		})
	}
}

// AddComment adds a comment to a post
// POST /api/v1/posts/:id/comment
func (h *PostsHandler) AddComment(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	query := `
		INSERT INTO post_comments (post_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	var comment models.PostComment
	err = h.db.QueryRow(query, postID, uid, req.Content).
		Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to add comment"),
		})
		return
	}

	comment.PostID = postID
	comment.UserID = uid
	comment.Content = req.Content

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Comment added",
		Data:    comment,
	})
}

// DeleteComment deletes a comment (user can delete own comments)
// DELETE /api/v1/posts/:id/comments/:comment_id
func (h *PostsHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid comment ID"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	userRole, _ := c.Get("user_role")

	// Check if user owns the comment or is admin
	var ownerID uuid.UUID
	err = h.db.QueryRow("SELECT user_id FROM post_comments WHERE id = $1", commentID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("Comment not found"),
		})
		return
	}

	// Allow deletion if owner or admin
	isOwner := ownerID == uid
	isAdmin := userRole.(models.UserRole) == models.RoleAdmin || userRole.(models.UserRole) == models.RoleFaculty

	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error:   strPtr("You can only delete your own comments"),
		})
		return
	}

	query := "UPDATE post_comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1"
	_, err = h.db.Exec(query, commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to delete comment"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Comment deleted",
	})
}

// TrackShare tracks a post share
// POST /api/v1/posts/:id/share
func (h *PostsHandler) TrackShare(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	var req models.CreateShareRequest
	c.ShouldBindJSON(&req) // Optional binding

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	query := "INSERT INTO post_shares (post_id, user_id, share_method) VALUES ($1, $2, $3)"
	_, err = h.db.Exec(query, postID, uid, req.ShareMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to track share"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Share tracked",
	})
}

// TrackView tracks a post view
// POST /api/v1/posts/:id/view
func (h *PostsHandler) TrackView(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid post ID"),
		})
		return
	}

	userID, exists := c.Get("user_id")
	var uid *uuid.UUID
	if exists {
		u := userID.(uuid.UUID)
		uid = &u
	}

	query := "INSERT INTO post_views (post_id, user_id) VALUES ($1, $2)"
	_, err = h.db.Exec(query, postID, uid)
	if err != nil {
		// Silently fail for views (non-critical)
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "View tracked",
	})
}

// Helper functions
func (h *PostsHandler) checkUserLikedPost(postID, userID uuid.UUID) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)"
	h.db.QueryRow(query, postID, userID).Scan(&exists)
	return exists
}

func (h *PostsHandler) checkUserSharedPost(postID, userID uuid.UUID) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM post_shares WHERE post_id = $1 AND user_id = $2)"
	h.db.QueryRow(query, postID, userID).Scan(&exists)
	return exists
}
