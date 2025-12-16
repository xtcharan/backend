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

type ScheduleHandler struct {
	db *database.DB
}

func NewScheduleHandler(db *database.DB) *ScheduleHandler {
	return &ScheduleHandler{db: db}
}

// ListSchedules returns schedules for a specific date
// Official schedules + personal schedules for the authenticated user
func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	// Get date from query parameter
	dateStr := c.Query("date")
	if dateStr == "" {
		// Default to today
		dateStr = time.Now().Format("2006-01-02")
	}

	// Parse the date
	scheduleDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid date format, use YYYY-MM-DD"),
		})
		return
	}

	// Check if user is authenticated (optional for viewing official schedules)
	userID, userExists := c.Get("user_id")

	var rows *sql.Rows
	if userExists {
		// Get official schedules + personal schedules for this user
		rows, err = h.db.Query(`
			SELECT id, title, description, schedule_date, start_time, end_time, location, 
			       schedule_type, created_by, user_id, created_at, updated_at
			FROM schedules
			WHERE schedule_date = $1 
			  AND (schedule_type = 'official' OR (schedule_type = 'personal' AND user_id = $2))
			ORDER BY start_time ASC
		`, scheduleDate, userID.(uuid.UUID))
	} else {
		// Get only official schedules
		rows, err = h.db.Query(`
			SELECT id, title, description, schedule_date, start_time, end_time, location,
			       schedule_type, created_by, user_id, created_at, updated_at
			FROM schedules
			WHERE schedule_date = $1 AND schedule_type = 'official'
			ORDER BY start_time ASC
		`, scheduleDate)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch schedules"),
		})
		return
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		err := rows.Scan(
			&schedule.ID, &schedule.Title, &schedule.Description, &schedule.ScheduleDate,
			&schedule.StartTime, &schedule.EndTime, &schedule.Location, &schedule.ScheduleType,
			&schedule.CreatedBy, &schedule.UserID, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			continue
		}
		schedules = append(schedules, schedule)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    schedules,
	})
}

// GetSchedule returns a single schedule by ID
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid schedule ID"),
		})
		return
	}

	var schedule models.Schedule
	err = h.db.QueryRow(`
		SELECT id, title, description, schedule_date, start_time, end_time, location,
		       schedule_type, created_by, user_id, created_at, updated_at
		FROM schedules
		WHERE id = $1
	`, id).Scan(
		&schedule.ID, &schedule.Title, &schedule.Description, &schedule.ScheduleDate,
		&schedule.StartTime, &schedule.EndTime, &schedule.Location, &schedule.ScheduleType,
		&schedule.CreatedBy, &schedule.UserID, &schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("schedule not found"),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch schedule"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    schedule,
	})
}

// CreateSchedule creates a new schedule item
// Admin can create 'official' schedules, students can only create 'personal' schedules
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("CreateSchedule validation error: %v\n", err.Error())
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr(fmt.Sprintf("invalid request body: %s", err.Error())),
		})
		return
	}

	// Parse schedule date
	scheduleDate, err := time.Parse("2006-01-02", req.ScheduleDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid schedule_date format, use YYYY-MM-DD"),
		})
		return
	}

	// Determine schedule type and validate permissions
	scheduleType := "personal"
	var targetUserID *uuid.UUID

	if req.ScheduleType == "official" {
		// Only admin can create official schedules
		roleVal, _ := userRole.(models.UserRole)
		if roleVal != models.RoleAdmin {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("only admin can create official schedules"),
			})
			return
		}
		scheduleType = "official"
		// Official schedules have no specific user_id
		targetUserID = nil
	} else {
		// Personal schedule - associate with the current user
		scheduleType = "personal"
		uid := userID.(uuid.UUID)
		targetUserID = &uid
	}

	var schedule models.Schedule
	err = h.db.QueryRow(`
		INSERT INTO schedules (title, description, schedule_date, start_time, end_time, location, schedule_type, created_by, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, title, description, schedule_date, start_time, end_time, location, schedule_type, created_by, user_id, created_at, updated_at
	`, req.Title, req.Description, scheduleDate, req.StartTime, req.EndTime, req.Location, scheduleType, userID.(uuid.UUID), targetUserID).Scan(
		&schedule.ID, &schedule.Title, &schedule.Description, &schedule.ScheduleDate,
		&schedule.StartTime, &schedule.EndTime, &schedule.Location, &schedule.ScheduleType,
		&schedule.CreatedBy, &schedule.UserID, &schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err != nil {
		fmt.Printf("CreateSchedule database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to create schedule"),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "schedule created successfully",
		Data:    schedule,
	})
}

// UpdateSchedule updates an existing schedule
// Admin can update any schedule, users can only update their own personal schedules
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid schedule ID"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	// First, fetch the existing schedule to check permissions
	var existingSchedule models.Schedule
	err = h.db.QueryRow(`
		SELECT id, title, description, schedule_date, start_time, end_time, location,
		       schedule_type, created_by, user_id, created_at, updated_at
		FROM schedules WHERE id = $1
	`, id).Scan(
		&existingSchedule.ID, &existingSchedule.Title, &existingSchedule.Description,
		&existingSchedule.ScheduleDate, &existingSchedule.StartTime, &existingSchedule.EndTime,
		&existingSchedule.Location, &existingSchedule.ScheduleType, &existingSchedule.CreatedBy,
		&existingSchedule.UserID, &existingSchedule.CreatedAt, &existingSchedule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("schedule not found"),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch schedule"),
		})
		return
	}

	// Check permissions
	// Admin can update any schedule
	// Non-admin can only update their own personal schedules
	roleVal, _ := userRole.(models.UserRole)
	if roleVal != models.RoleAdmin {
		if existingSchedule.ScheduleType == "official" {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("cannot edit official schedules"),
			})
			return
		}
		if existingSchedule.UserID == nil || *existingSchedule.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("cannot edit other users' schedules"),
			})
			return
		}
	}

	var req models.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr(fmt.Sprintf("invalid request body: %s", err.Error())),
		})
		return
	}

	// Build update query dynamically
	title := existingSchedule.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := existingSchedule.Description
	if req.Description != nil {
		description = req.Description
	}
	scheduleDate := existingSchedule.ScheduleDate
	if req.ScheduleDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.ScheduleDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   strPtr("invalid schedule_date format"),
			})
			return
		}
		scheduleDate = parsedDate
	}
	startTime := existingSchedule.StartTime
	if req.StartTime != nil {
		// Parse the string time
		parsedTime, err := time.Parse("15:04", *req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   strPtr("invalid start_time format, use HH:MM"),
			})
			return
		}
		startTime = models.TimeString(parsedTime)
	}
	endTime := existingSchedule.EndTime
	if req.EndTime != nil {
		parsedTime, err := time.Parse("15:04", *req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   strPtr("invalid end_time format, use HH:MM"),
			})
			return
		}
		ts := models.TimeString(parsedTime)
		endTime = &ts
	}
	location := existingSchedule.Location
	if req.Location != nil {
		location = req.Location
	}

	var schedule models.Schedule
	err = h.db.QueryRow(`
		UPDATE schedules
		SET title = $1, description = $2, schedule_date = $3, start_time = $4, end_time = $5, location = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING id, title, description, schedule_date, start_time, end_time, location, schedule_type, created_by, user_id, created_at, updated_at
	`, title, description, scheduleDate, startTime, endTime, location, id).Scan(
		&schedule.ID, &schedule.Title, &schedule.Description, &schedule.ScheduleDate,
		&schedule.StartTime, &schedule.EndTime, &schedule.Location, &schedule.ScheduleType,
		&schedule.CreatedBy, &schedule.UserID, &schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err != nil {
		fmt.Printf("UpdateSchedule database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to update schedule"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "schedule updated successfully",
		Data:    schedule,
	})
}

// DeleteSchedule deletes a schedule
// Admin can delete any schedule, users can only delete their own personal schedules
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid schedule ID"),
		})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	// First, fetch the existing schedule to check permissions
	var existingSchedule models.Schedule
	err = h.db.QueryRow(`
		SELECT id, schedule_type, user_id FROM schedules WHERE id = $1
	`, id).Scan(&existingSchedule.ID, &existingSchedule.ScheduleType, &existingSchedule.UserID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("schedule not found"),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch schedule"),
		})
		return
	}

	// Check permissions
	roleVal, _ := userRole.(models.UserRole)
	if roleVal != models.RoleAdmin {
		if existingSchedule.ScheduleType == "official" {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("cannot delete official schedules"),
			})
			return
		}
		if existingSchedule.UserID == nil || *existingSchedule.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("cannot delete other users' schedules"),
			})
			return
		}
	}

	_, err = h.db.Exec(`DELETE FROM schedules WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to delete schedule"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "schedule deleted successfully",
	})
}
