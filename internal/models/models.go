package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserRole defines user roles in the system
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleStudent UserRole = "student"
	RoleFaculty UserRole = "faculty"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	Role         UserRole   `json:"role" db:"role"`
	AvatarURL    *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Department   *string    `json:"department,omitempty" db:"department"`
	Year         *int       `json:"year,omitempty" db:"year"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email      string  `json:"email" binding:"required,email"`
	Password   string  `json:"password" binding:"required,min=8"`
	FullName   string  `json:"full_name" binding:"required"`
	Department *string `json:"department"`
	Year       *int    `json:"year"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Event represents an event in the system
type Event struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	ImageURL    *string    `json:"image_url,omitempty" db:"image_url"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     time.Time  `json:"end_date" db:"end_date"`
	Location    *string    `json:"location,omitempty" db:"location"`
	Category    *string    `json:"category,omitempty" db:"category"`
	MaxCapacity *int       `json:"max_capacity,omitempty" db:"max_capacity"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

// CreateEventRequest represents event creation data
type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description *string   `json:"description"`
	ImageURL    *string   `json:"image_url"`
	StartDate   JSONTime  `json:"start_date" binding:"required"`
	EndDate     JSONTime  `json:"end_date" binding:"required"`
	Location    *string   `json:"location"`
	Category    *string   `json:"category"`
	MaxCapacity *int      `json:"max_capacity"`
}

// JSONTime is a custom time type that handles multiple datetime formats
type JSONTime time.Time

// UnmarshalJSON parses datetime strings in multiple formats
func (jt *JSONTime) UnmarshalJSON(data []byte) error {
	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return fmt.Errorf("failed to unmarshal datetime: %w", err)
	}

	// Try multiple formats to handle Flutter and standard formats
	formats := []string{
		"2006-01-02T15:04:05.000Z",        // Flutter format with milliseconds
		"2006-01-02T15:04:05Z",             // ISO8601 without milliseconds
		"2006-01-02T15:04:05.000Z07:00",   // RFC3339 with milliseconds
		"2006-01-02T15:04:05Z07:00",       // RFC3339
		"2006-01-02T15:04:05",              // Date and time only
	}

	var parsedTime time.Time
	var err error

	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			*jt = JSONTime(parsedTime)
			return nil
		}
	}

	return fmt.Errorf("unable to parse datetime '%s' in any supported format: %w", dateStr, err)
}

// MarshalJSON converts JSONTime to RFC3339 format
func (jt JSONTime) MarshalJSON() ([]byte, error) {
	t := time.Time(jt)
	return json.Marshal(t.Format(time.RFC3339Nano))
}

// Time converts JSONTime to time.Time
func (jt JSONTime) Time() time.Time {
	return time.Time(jt)
}

// EventRegistration represents a user's registration for an event
type EventRegistration struct {
	ID           uuid.UUID `json:"id" db:"id"`
	EventID      uuid.UUID `json:"event_id" db:"event_id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
}

// Club represents a college club
type Club struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	Department  *string    `json:"department,omitempty" db:"department"`
	ImageURL    *string    `json:"image_url,omitempty" db:"image_url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

// Announcement represents a college announcement
type Announcement struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	Priority  string     `json:"priority" db:"priority"` // low, medium, high, urgent
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FromUser  uuid.UUID `json:"from_user" db:"from_user"`
	ToUser    uuid.UUID `json:"to_user" db:"to_user"`
	Message   string    `json:"message" db:"message"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}
