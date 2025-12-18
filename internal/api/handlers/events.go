package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/college-event-backend/internal/models"
	"github.com/yourusername/college-event-backend/pkg/database"
)

type EventHandler struct {
	db *database.DB
}

func NewEventHandler(db *database.DB) *EventHandler {
	return &EventHandler{db: db}
}

// ListEvents returns all events
func (h *EventHandler) ListEvents(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, title, description, banner_url, start_date, end_date, location, category, 
		       status, max_participants, current_participants, registration_deadline, is_featured,
		       is_paid_event, event_amount, currency,
		       club_id, created_by, created_at, updated_at
		FROM events
		WHERE deleted_at IS NULL AND end_date >= $1
		ORDER BY start_date ASC
	`, time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch events"),
		})
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.Title, &event.Description, &event.BannerURL,
			&event.StartDate, &event.EndDate, &event.Location, &event.Category,
			&event.Status, &event.MaxParticipants, &event.CurrentParticipants,
			&event.RegistrationDeadline, &event.IsFeatured,
			&event.IsPaidEvent, &event.EventAmount, &event.Currency,
			&event.ClubID, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    events,
	})
}

// GetEvent returns a single event by ID
func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid event ID"),
		})
		return
	}

	var event models.Event
	err = h.db.QueryRow(`
		SELECT id, title, description, banner_url, start_date, end_date, location, category,
		       status, max_participants, current_participants, registration_deadline, is_featured,
		       is_paid_event, event_amount, currency,
		       club_id, created_by, created_at, updated_at
		FROM events
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&event.ID, &event.Title, &event.Description, &event.BannerURL,
		&event.StartDate, &event.EndDate, &event.Location, &event.Category,
		&event.Status, &event.MaxParticipants, &event.CurrentParticipants,
		&event.RegistrationDeadline, &event.IsFeatured,
		&event.IsPaidEvent, &event.EventAmount, &event.Currency,
		&event.ClubID, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("event not found"),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch event"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    event,
	})
}

// CreateEvent creates a new event (admin only)
func (h *EventHandler) CreateEvent(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log the error for debugging
		fmt.Printf("CreateEvent validation error: %v\n", err.Error())
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr(fmt.Sprintf("invalid request body: %s", err.Error())),
		})
		return
	}

	// Convert JSONTime to time.Time
	startTime := req.StartDate.Time()
	endTime := req.EndDate.Time()

	// Validate that end time is after start time
	if !endTime.After(startTime) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("end_date must be after start_date"),
		})
		return
	}

	// Use BannerURL if provided, otherwise use ImageURL for backward compatibility
	bannerURL := req.BannerURL
	if bannerURL == nil && req.ImageURL != nil {
		bannerURL = req.ImageURL
	}

	// Default currency to INR if not provided
	currency := req.Currency
	if currency == nil {
		defaultCurrency := "INR"
		currency = &defaultCurrency
	}

	var event models.Event
	err := h.db.QueryRow(`
		INSERT INTO events (title, description, banner_url, start_date, end_date, location, category, max_participants, is_paid_event, event_amount, currency, club_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, title, description, banner_url, start_date, end_date, location, category,
		          status, max_participants, current_participants, registration_deadline, is_featured,
		          is_paid_event, event_amount, currency,
		          club_id, created_by, created_at, updated_at
	`, req.Title, req.Description, bannerURL, startTime, endTime, req.Location, req.Category, req.MaxCapacity, req.IsPaidEvent, req.EventAmount, currency, req.ClubID, userID.(uuid.UUID)).Scan(
		&event.ID, &event.Title, &event.Description, &event.BannerURL,
		&event.StartDate, &event.EndDate, &event.Location, &event.Category,
		&event.Status, &event.MaxParticipants, &event.CurrentParticipants,
		&event.RegistrationDeadline, &event.IsFeatured,
		&event.IsPaidEvent, &event.EventAmount, &event.Currency,
		&event.ClubID, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		fmt.Printf("CreateEvent database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to create event"),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "event created successfully",
		Data:    event,
	})
}

// UpdateEvent updates an existing event (admin only)
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid event ID"),
		})
		return
	}

	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("UpdateEvent validation error: %v\n", err.Error())
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr(fmt.Sprintf("invalid request body: %s", err.Error())),
		})
		return
	}

	// Convert JSONTime to time.Time
	startTime := req.StartDate.Time()
	endTime := req.EndDate.Time()

	// Validate that end time is after start time
	if !endTime.After(startTime) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("end_date must be after start_date"),
		})
		return
	}

	// Use BannerURL if provided, otherwise use ImageURL for backward compatibility
	bannerURL := req.BannerURL
	if bannerURL == nil && req.ImageURL != nil {
		bannerURL = req.ImageURL
	}

	// Default currency to INR if not provided
	currency := req.Currency
	if currency == nil {
		defaultCurrency := "INR"
		currency = &defaultCurrency
	}

	var event models.Event
	err = h.db.QueryRow(`
		UPDATE events
		SET title = $1, description = $2, banner_url = $3, start_date = $4, end_date = $5, 
		    location = $6, category = $7, max_participants = $8, 
		    is_paid_event = $9, event_amount = $10, currency = $11,
		    club_id = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $13 AND deleted_at IS NULL
		RETURNING id, title, description, banner_url, start_date, end_date, location, category,
		          status, max_participants, current_participants, registration_deadline, is_featured,
		          is_paid_event, event_amount, currency,
		          club_id, created_by, created_at, updated_at
	`, req.Title, req.Description, bannerURL, startTime, endTime, req.Location, req.Category, req.MaxCapacity, req.IsPaidEvent, req.EventAmount, currency, req.ClubID, id).Scan(
		&event.ID, &event.Title, &event.Description, &event.BannerURL,
		&event.StartDate, &event.EndDate, &event.Location, &event.Category,
		&event.Status, &event.MaxParticipants, &event.CurrentParticipants,
		&event.RegistrationDeadline, &event.IsFeatured,
		&event.IsPaidEvent, &event.EventAmount, &event.Currency,
		&event.ClubID, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("event not found"),
		})
		return
	}

	if err != nil {
		fmt.Printf("UpdateEvent database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to update event"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "event updated successfully",
		Data:    event,
	})
}

// DeleteEvent deletes an event (admin only)
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid event ID"),
		})
		return
	}

	result, err := h.db.Exec(`
		UPDATE events
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to delete event"),
		})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("event not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "event deleted successfully",
	})
}
