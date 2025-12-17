package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/college-event-backend/internal/models"
)

// HouseHandler handles house-related requests
type HouseHandler struct {
	DB *sql.DB
}

// NewHouseHandler creates a new HouseHandler
func NewHouseHandler(db *sql.DB) *HouseHandler {
	return &HouseHandler{DB: db}
}

// ============================================================================
// HOUSES CRUD
// ============================================================================

// GetHouses returns all houses
func (h *HouseHandler) GetHouses(c *gin.Context) {
	query := `
		SELECT id, name, color, description, logo_url, points, created_at, updated_at
		FROM houses
		WHERE deleted_at IS NULL
		ORDER BY points DESC
	`
	rows, err := h.DB.QueryContext(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch houses"),
		})
		return
	}
	defer rows.Close()

	houses := []models.House{}
	for rows.Next() {
		var house models.House
		if err := rows.Scan(&house.ID, &house.Name, &house.Color, &house.Description, &house.LogoURL, &house.Points, &house.CreatedAt, &house.UpdatedAt); err != nil {
			continue
		}
		houses = append(houses, house)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    houses,
	})
}

// GetHouse returns a single house with its roles
func (h *HouseHandler) GetHouse(c *gin.Context) {
	houseID := c.Param("id")

	var house models.House
	query := `
		SELECT id, name, color, description, logo_url, points, created_at, updated_at
		FROM houses
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := h.DB.QueryRowContext(c.Request.Context(), query, houseID).Scan(
		&house.ID, &house.Name, &house.Color, &house.Description, &house.LogoURL, &house.Points, &house.CreatedAt, &house.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error:   strPtr("House not found"),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch house"),
		})
		return
	}

	// Fetch roles
	roleQuery := `
		SELECT id, house_id, member_name, role_title, display_order, created_at
		FROM house_roles
		WHERE house_id = $1
		ORDER BY display_order ASC, created_at ASC
	`
	roleRows, err := h.DB.QueryContext(c.Request.Context(), roleQuery, houseID)
	if err == nil {
		defer roleRows.Close()
		for roleRows.Next() {
			var role models.HouseRole
			if err := roleRows.Scan(&role.ID, &role.HouseID, &role.MemberName, &role.RoleTitle, &role.DisplayOrder, &role.CreatedAt); err == nil {
				house.Roles = append(house.Roles, role)
			}
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    house,
	})
}

// CreateHouse creates a new house (admin only)
func (h *HouseHandler) CreateHouse(c *gin.Context) {
	var req models.CreateHouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	var house models.House
	query := `
		INSERT INTO houses (name, color, description, logo_url, points)
		VALUES ($1, $2, $3, $4, 0)
		RETURNING id, name, color, description, logo_url, points, created_at
	`
	err := h.DB.QueryRowContext(
		c.Request.Context(),
		query,
		req.Name, req.Color, req.Description, req.LogoURL,
	).Scan(&house.ID, &house.Name, &house.Color, &house.Description, &house.LogoURL, &house.Points, &house.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to create house: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "House created successfully",
		Data:    house,
	})
}

// UpdateHouse updates an existing house (admin only)
func (h *HouseHandler) UpdateHouse(c *gin.Context) {
	houseID := c.Param("id")

	var req models.UpdateHouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	query := `
		UPDATE houses SET
			name = COALESCE($2, name),
			color = COALESCE($3, color),
			description = COALESCE($4, description),
			logo_url = COALESCE($5, logo_url),
			points = COALESCE($6, points),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, color, description, logo_url, points, created_at, updated_at
	`

	var house models.House
	err := h.DB.QueryRowContext(
		c.Request.Context(),
		query,
		houseID, req.Name, req.Color, req.Description, req.LogoURL, req.Points,
	).Scan(&house.ID, &house.Name, &house.Color, &house.Description, &house.LogoURL, &house.Points, &house.CreatedAt, &house.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error:   strPtr("House not found"),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to update house"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "House updated successfully",
		Data:    house,
	})
}

// DeleteHouse permanently deletes a house (admin only)
// Note: Related data (roles, announcements, events) will be cascade deleted
func (h *HouseHandler) DeleteHouse(c *gin.Context) {
	houseID := c.Param("id")

	query := `DELETE FROM houses WHERE id = $1`
	result, err := h.DB.ExecContext(c.Request.Context(), query, houseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to delete house"),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("House not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "House deleted successfully",
	})
}

// ============================================================================
// HOUSE ROLES
// ============================================================================

// AddHouseRole adds a role to a house
func (h *HouseHandler) AddHouseRole(c *gin.Context) {
	houseID := c.Param("id")

	var req models.CreateHouseRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	displayOrder := 0
	if req.DisplayOrder != nil {
		displayOrder = *req.DisplayOrder
	}

	var role models.HouseRole
	query := `
		INSERT INTO house_roles (house_id, member_name, role_title, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, house_id, member_name, role_title, display_order, created_at
	`
	err := h.DB.QueryRowContext(
		c.Request.Context(),
		query,
		houseID, req.MemberName, req.RoleTitle, displayOrder,
	).Scan(&role.ID, &role.HouseID, &role.MemberName, &role.RoleTitle, &role.DisplayOrder, &role.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to add role: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Role added successfully",
		Data:    role,
	})
}

// RemoveHouseRole removes a role from a house
func (h *HouseHandler) RemoveHouseRole(c *gin.Context) {
	roleID := c.Param("role_id")

	query := `DELETE FROM house_roles WHERE id = $1`
	result, err := h.DB.ExecContext(c.Request.Context(), query, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to remove role"),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("Role not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Role removed successfully",
	})
}

// ============================================================================
// HOUSE ANNOUNCEMENTS
// ============================================================================

// GetAnnouncements returns all announcements for a house
func (h *HouseHandler) GetAnnouncements(c *gin.Context) {
	houseID := c.Param("id")
	userID, _ := c.Get("user_id")

	query := `
		SELECT 
			ha.id, ha.house_id, ha.title, ha.content, ha.created_by, ha.created_at, ha.updated_at,
			COALESCE(u.full_name, 'Unknown') as author_name,
			(SELECT COUNT(*) FROM announcement_likes WHERE announcement_id = ha.id) as like_count,
			(SELECT COUNT(*) FROM announcement_comments WHERE announcement_id = ha.id AND deleted_at IS NULL) as comment_count
		FROM house_announcements ha
		LEFT JOIN users u ON ha.created_by = u.id
		WHERE ha.house_id = $1 AND ha.deleted_at IS NULL
		ORDER BY ha.created_at DESC
	`
	rows, err := h.DB.QueryContext(c.Request.Context(), query, houseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch announcements"),
		})
		return
	}
	defer rows.Close()

	announcements := []models.HouseAnnouncement{}
	for rows.Next() {
		var a models.HouseAnnouncement
		if err := rows.Scan(&a.ID, &a.HouseID, &a.Title, &a.Content, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt, &a.AuthorName, &a.LikeCount, &a.CommentCount); err == nil {
			announcements = append(announcements, a)
		}
	}

	// Check if current user liked each announcement
	if userID != nil {
		for i := range announcements {
			var count int
			likeQuery := `SELECT COUNT(*) FROM announcement_likes WHERE announcement_id = $1 AND user_id = $2`
			h.DB.QueryRowContext(c.Request.Context(), likeQuery, announcements[i].ID, userID).Scan(&count)
			announcements[i].IsLikedByMe = count > 0
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    announcements,
	})
}

// CreateAnnouncement creates a new announcement (admin only)
func (h *HouseHandler) CreateAnnouncement(c *gin.Context) {
	houseID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req models.CreateHouseAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	var announcement models.HouseAnnouncement
	query := `
		INSERT INTO house_announcements (house_id, title, content, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, house_id, title, content, created_by, created_at, updated_at
	`
	err := h.DB.QueryRowContext(
		c.Request.Context(),
		query,
		houseID, req.Title, req.Content, userID,
	).Scan(&announcement.ID, &announcement.HouseID, &announcement.Title, &announcement.Content, &announcement.CreatedBy, &announcement.CreatedAt, &announcement.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to create announcement: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Announcement created successfully",
		Data:    announcement,
	})
}

// LikeAnnouncement toggles like on an announcement
func (h *HouseHandler) LikeAnnouncement(c *gin.Context) {
	announcementID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("User not authenticated"),
		})
		return
	}

	// Check if already liked
	var count int
	checkQuery := `SELECT COUNT(*) FROM announcement_likes WHERE announcement_id = $1 AND user_id = $2`
	h.DB.QueryRowContext(c.Request.Context(), checkQuery, announcementID, userID).Scan(&count)

	if count > 0 {
		// Unlike
		deleteQuery := `DELETE FROM announcement_likes WHERE announcement_id = $1 AND user_id = $2`
		h.DB.ExecContext(c.Request.Context(), deleteQuery, announcementID, userID)
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Unliked",
			Data:    map[string]bool{"liked": false},
		})
	} else {
		// Like
		insertQuery := `INSERT INTO announcement_likes (announcement_id, user_id) VALUES ($1, $2)`
		h.DB.ExecContext(c.Request.Context(), insertQuery, announcementID, userID)
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Liked",
			Data:    map[string]bool{"liked": true},
		})
	}
}

// GetComments returns comments for an announcement
func (h *HouseHandler) GetComments(c *gin.Context) {
	announcementID := c.Param("id")

	query := `
		SELECT 
			ac.id, ac.announcement_id, ac.user_id, ac.content, ac.created_at, ac.updated_at,
			COALESCE(u.full_name, 'Unknown') as user_name,
			COALESCE(u.avatar_url, '') as avatar_url
		FROM announcement_comments ac
		LEFT JOIN users u ON ac.user_id = u.id
		WHERE ac.announcement_id = $1 AND ac.deleted_at IS NULL
		ORDER BY ac.created_at ASC
	`
	rows, err := h.DB.QueryContext(c.Request.Context(), query, announcementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch comments"),
		})
		return
	}
	defer rows.Close()

	comments := []models.AnnouncementComment{}
	for rows.Next() {
		var cm models.AnnouncementComment
		if err := rows.Scan(&cm.ID, &cm.AnnouncementID, &cm.UserID, &cm.Content, &cm.CreatedAt, &cm.UpdatedAt, &cm.UserName, &cm.AvatarURL); err == nil {
			comments = append(comments, cm)
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    comments,
	})
}

// AddComment adds a comment to an announcement
func (h *HouseHandler) AddComment(c *gin.Context) {
	announcementID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("User not authenticated"),
		})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	var comment models.AnnouncementComment
	query := `
		INSERT INTO announcement_comments (announcement_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, announcement_id, user_id, content, created_at, updated_at
	`
	err := h.DB.QueryRowContext(
		c.Request.Context(),
		query,
		announcementID, userID, req.Content,
	).Scan(&comment.ID, &comment.AnnouncementID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to add comment: " + err.Error()),
		})
		return
	}

	// Get user info
	userQuery := `SELECT full_name, COALESCE(avatar_url, '') FROM users WHERE id = $1`
	h.DB.QueryRowContext(c.Request.Context(), userQuery, userID).Scan(&comment.UserName, &comment.AvatarURL)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Comment added successfully",
		Data:    comment,
	})
}

// ============================================================================
// HOUSE EVENTS
// ============================================================================

// GetHouseEvents returns all events for a house
func (h *HouseHandler) GetHouseEvents(c *gin.Context) {
	houseID := c.Param("id")
	userID, _ := c.Get("user_id")

	query := `
		SELECT 
			he.id, he.house_id, he.title, he.description, he.event_date, 
			he.start_time::text, he.end_time::text, he.venue, he.max_participants,
			he.registration_deadline, he.status, he.created_by, he.created_at, he.updated_at,
			(SELECT COUNT(*) FROM house_event_enrollments WHERE event_id = he.id) as enrollment_count
		FROM house_events he
		WHERE he.house_id = $1 AND he.deleted_at IS NULL
		ORDER BY he.event_date ASC
	`
	rows, err := h.DB.QueryContext(c.Request.Context(), query, houseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to fetch events: " + err.Error()),
		})
		return
	}
	defer rows.Close()

	events := []models.HouseEvent{}
	for rows.Next() {
		var e models.HouseEvent
		if err := rows.Scan(&e.ID, &e.HouseID, &e.Title, &e.Description, &e.EventDate, &e.StartTime, &e.EndTime, &e.Venue, &e.MaxParticipants, &e.RegistrationDeadline, &e.Status, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt, &e.EnrollmentCount); err == nil {
			events = append(events, e)
		}
	}

	// Check if current user is enrolled in each event
	if userID != nil {
		for i := range events {
			var count int
			enrollQuery := `SELECT COUNT(*) FROM house_event_enrollments WHERE event_id = $1 AND user_id = $2`
			h.DB.QueryRowContext(c.Request.Context(), enrollQuery, events[i].ID, userID).Scan(&count)
			events[i].IsEnrolled = count > 0
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    events,
	})
}

// CreateHouseEvent creates a new house event (admin only)
func (h *HouseHandler) CreateHouseEvent(c *gin.Context) {
	houseID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req models.CreateHouseEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid request: " + err.Error()),
		})
		return
	}

	// Parse event date
	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid event date format"),
		})
		return
	}

	// Parse registration deadline if provided
	var regDeadline *time.Time
	if req.RegistrationDeadline != nil {
		parsed, err := time.Parse("2006-01-02", *req.RegistrationDeadline)
		if err == nil {
			regDeadline = &parsed
		}
	}

	var event models.HouseEvent
	query := `
		INSERT INTO house_events (house_id, title, description, event_date, start_time, end_time, venue, max_participants, registration_deadline, created_by)
		VALUES ($1, $2, $3, $4, $5::time, $6::time, $7, $8, $9, $10)
		RETURNING id, house_id, title, description, event_date, start_time::text, end_time::text, venue, max_participants, registration_deadline, status, created_by, created_at, updated_at
	`
	err = h.DB.QueryRowContext(
		c.Request.Context(),
		query,
		houseID, req.Title, req.Description, eventDate, req.StartTime, req.EndTime, req.Venue, req.MaxParticipants, regDeadline, userID,
	).Scan(&event.ID, &event.HouseID, &event.Title, &event.Description, &event.EventDate, &event.StartTime, &event.EndTime, &event.Venue, &event.MaxParticipants, &event.RegistrationDeadline, &event.Status, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to create event: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Event created successfully",
		Data:    event,
	})
}

// EnrollInEvent enrolls user in a house event
func (h *HouseHandler) EnrollInEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("User not authenticated"),
		})
		return
	}

	// Check if already enrolled
	var count int
	checkQuery := `SELECT COUNT(*) FROM house_event_enrollments WHERE event_id = $1 AND user_id = $2`
	h.DB.QueryRowContext(c.Request.Context(), checkQuery, eventID, userID).Scan(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Already enrolled in this event"),
		})
		return
	}

	// Check event capacity
	var maxParticipants sql.NullInt64
	var enrollmentCount int
	capacityQuery := `SELECT max_participants, (SELECT COUNT(*) FROM house_event_enrollments WHERE event_id = $1) FROM house_events WHERE id = $1`
	h.DB.QueryRowContext(c.Request.Context(), capacityQuery, eventID).Scan(&maxParticipants, &enrollmentCount)

	if maxParticipants.Valid && enrollmentCount >= int(maxParticipants.Int64) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Event is full"),
		})
		return
	}

	// Enroll
	insertQuery := `INSERT INTO house_event_enrollments (event_id, user_id) VALUES ($1, $2)`
	_, err := h.DB.ExecContext(c.Request.Context(), insertQuery, eventID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to enroll: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Successfully enrolled in event",
	})
}

// UnenrollFromEvent removes user enrollment from a house event
func (h *HouseHandler) UnenrollFromEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("User not authenticated"),
		})
		return
	}

	deleteQuery := `DELETE FROM house_event_enrollments WHERE event_id = $1 AND user_id = $2`
	result, err := h.DB.ExecContext(c.Request.Context(), deleteQuery, eventID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to unenroll"),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("Not enrolled in this event"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Successfully unenrolled from event",
	})
}
