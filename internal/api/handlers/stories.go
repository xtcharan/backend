package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yourusername/college-event-backend/internal/models"
)

// StoriesHandler handles story-related requests
type StoriesHandler struct {
	db *sql.DB
}

// NewStoriesHandler creates a new stories handler
func NewStoriesHandler(db *sql.DB) *StoriesHandler {
	return &StoriesHandler{db: db}
}

// CreateStory creates a new 24-hour story (admin-only)
// POST /api/v1/admin/stories
func (h *StoriesHandler) CreateStory(c *gin.Context) {
	var req models.CreateStoryRequest
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
	if req.ContentType == models.ContentTypeVideo && req.VideoURL == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("video_url required for video content"),
		})
		return
	}

	// Insert story
	query := `
		INSERT INTO stories (
			created_by, club_id, house_id, content_type,
			image_url, video_url, thumbnail_url, duration_seconds,
			description, hashtags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, expires_at
	`

	hashtags := req.Hashtags
	if hashtags == nil {
		hashtags = []string{}
	}

	var story models.Story
	err := h.db.QueryRow(
		query,
		creatorID, req.ClubID, req.HouseID, req.ContentType,
		req.ImageURL, req.VideoURL, req.ThumbnailURL, req.DurationSecs,
		req.Description, pq.Array(hashtags),
	).Scan(&story.ID, &story.CreatedAt, &story.ExpiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to create story: " + err.Error()),
		})
		return
	}

	// Populate response
	story.CreatedBy = creatorID
	story.ClubID = req.ClubID
	story.HouseID = req.HouseID
	story.ContentType = req.ContentType
	story.ImageURL = req.ImageURL
	story.VideoURL = req.VideoURL
	story.ThumbnailURL = req.ThumbnailURL
	story.DurationSecs = req.DurationSecs
	story.Description = req.Description
	story.Hashtags = hashtags

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Story created successfully (expires in 24 hours)",
		Data:    story,
	})
}

// ListStories lists active (non-expired) stories
// GET /api/v1/stories
func (h *StoriesHandler) ListStories(c *gin.Context) {
	query := `
		SELECT 
			s.id, s.created_by, s.club_id, s.house_id,
			s.content_type, s.image_url, s.video_url, s.thumbnail_url, s.duration_seconds,
			s.description, s.hashtags, s.created_at, s.expires_at,
			s.view_count, s.like_count,
			u.id, u.full_name, u.avatar_url, u.role
		FROM stories s
		JOIN users u ON s.created_by = u.id
		WHERE s.expires_at > CURRENT_TIMESTAMP
		ORDER BY s.created_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch stories"),
		})
		return
	}
	defer rows.Close()

	stories := []models.StoryResponse{}
	now := time.Now()

	for rows.Next() {
		var sr models.StoryResponse
		var hashtags pq.StringArray

		err := rows.Scan(
			&sr.ID, &sr.CreatedBy, &sr.ClubID, &sr.HouseID,
			&sr.ContentType, &sr.ImageURL, &sr.VideoURL, &sr.ThumbnailURL, &sr.DurationSecs,
			&sr.Description, &hashtags, &sr.CreatedAt, &sr.ExpiresAt,
			&sr.ViewCount, &sr.LikeCount,
			&sr.Creator.ID, &sr.Creator.FullName, &sr.Creator.AvatarURL, &sr.Creator.Role,
		)
		if err != nil {
			continue
		}

		sr.Hashtags = hashtags
		sr.TimeRemaining = int(sr.ExpiresAt.Sub(now).Seconds())

		// Check if current user liked/viewed
		userID, exists := c.Get("user_id")
		if exists {
			uid := userID.(uuid.UUID)
			sr.IsLikedByMe = h.checkUserLikedStory(sr.ID, uid)
			sr.IsViewedByMe = h.checkUserViewedStory(sr.ID, uid)
		}

		stories = append(stories, sr)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    gin.H{"stories": stories},
	})
}

// ToggleLike toggles like on a story
// POST /api/v1/stories/:id/like
func (h *StoriesHandler) ToggleLike(c *gin.Context) {
	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid story ID"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Check if already liked
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM story_likes WHERE story_id = $1 AND user_id = $2)"
	h.db.QueryRow(checkQuery, storyID, uid).Scan(&exists)

	if exists {
		// Unlike
		_, err = h.db.Exec("DELETE FROM story_likes WHERE story_id = $1 AND user_id = $2", storyID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   strPtr("Failed to unlike story"),
			})
			return
		}
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Story unliked",
			Data:    gin.H{"liked": false},
		})
	} else {
		// Like
		_, err = h.db.Exec("INSERT INTO story_likes (story_id, user_id) VALUES ($1, $2)", storyID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   strPtr("Failed to like story"),
			})
			return
		}
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Story liked",
			Data:    gin.H{"liked": true},
		})
	}
}

// TrackView tracks a story view
// POST /api/v1/stories/:id/view
func (h *StoriesHandler) TrackView(c *gin.Context) {
	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid story ID"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Insert view (unique constraint prevents duplicates)
	query := "INSERT INTO story_views (story_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err = h.db.Exec(query, storyID, uid)
	if err != nil {
		// Silently fail
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

// HardDeleteStory permanently deletes a story (admin-only)
// DELETE /api/v1/admin/stories/:id/hard
func (h *StoriesHandler) HardDeleteStory(c *gin.Context) {
	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid story ID"),
		})
		return
	}

	query := "SELECT hard_delete_story($1)"
	var success bool
	err = h.db.QueryRow(query, storyID).Scan(&success)
	if err != nil || !success {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to delete story"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Story permanently deleted",
	})
}

// Helper functions
func (h *StoriesHandler) checkUserLikedStory(storyID, userID uuid.UUID) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM story_likes WHERE story_id = $1 AND user_id = $2)"
	h.db.QueryRow(query, storyID, userID).Scan(&exists)
	return exists
}

func (h *StoriesHandler) checkUserViewedStory(storyID, userID uuid.UUID) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM story_views WHERE story_id = $1 AND user_id = $2)"
	h.db.QueryRow(query, storyID, userID).Scan(&exists)
	return exists
}
