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
	Token        string `json:"token"`
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ============================================================================
// DEPARTMENTS
// ============================================================================

// Department represents an academic department
type Department struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Code         string    `json:"code" db:"code"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description,omitempty" db:"description"`
	LogoURL      *string   `json:"logo_url,omitempty" db:"logo_url"`
	IconName     *string   `json:"icon_name,omitempty" db:"icon_name"`
	ColorHex     string    `json:"color_hex" db:"color_hex"`
	TotalMembers int       `json:"total_members" db:"total_members"`
	TotalClubs   int       `json:"total_clubs" db:"total_clubs"`
	TotalEvents  int       `json:"total_events" db:"total_events"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDepartmentRequest represents department creation data
type CreateDepartmentRequest struct {
	Code        string  `json:"code" binding:"required,max=10"`
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
	IconName    *string `json:"icon_name"`
	ColorHex    *string `json:"color_hex"`
}

// UpdateDepartmentRequest represents department update data
type UpdateDepartmentRequest struct {
	Code        *string `json:"code" binding:"omitempty,max=10"`
	Name        *string `json:"name" binding:"omitempty,max=255"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
	IconName    *string `json:"icon_name"`
	ColorHex    *string `json:"color_hex"`
}

// ============================================================================
// CLUBS
// ============================================================================

// Club represents a student club
type Club struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	DepartmentID   *uuid.UUID      `json:"department_id,omitempty" db:"department_id"`
	Name           string          `json:"name" db:"name"`
	Tagline        *string         `json:"tagline,omitempty" db:"tagline"`
	Description    *string         `json:"description,omitempty" db:"description"`
	LogoURL        *string         `json:"logo_url,omitempty" db:"logo_url"`
	PrimaryColor   string          `json:"primary_color" db:"primary_color"`
	SecondaryColor string          `json:"secondary_color" db:"secondary_color"`
	MemberCount    int             `json:"member_count" db:"member_count"`
	EventCount     int             `json:"event_count" db:"event_count"`
	AwardsCount    int             `json:"awards_count" db:"awards_count"`
	Rating         float64         `json:"rating" db:"rating"`
	Email          *string         `json:"email,omitempty" db:"email"`
	Phone          *string         `json:"phone,omitempty" db:"phone"`
	Website        *string         `json:"website,omitempty" db:"website"`
	SocialLinks    json.RawMessage `json:"social_links,omitempty" db:"social_links"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateClubRequest represents club creation data
type CreateClubRequest struct {
	DepartmentID   *uuid.UUID      `json:"department_id"`
	Name           string          `json:"name" binding:"required,max=255"`
	Tagline        *string         `json:"tagline"`
	Description    *string         `json:"description"`
	LogoURL        *string         `json:"logo_url"`
	PrimaryColor   *string         `json:"primary_color"`
	SecondaryColor *string         `json:"secondary_color"`
	Email          *string         `json:"email"`
	Phone          *string         `json:"phone"`
	Website        *string         `json:"website"`
	SocialLinks    json.RawMessage `json:"social_links"`
}

// UpdateClubRequest represents club update data
type UpdateClubRequest struct {
	DepartmentID   *uuid.UUID      `json:"department_id"`
	Name           *string         `json:"name" binding:"omitempty,max=255"`
	Tagline        *string         `json:"tagline"`
	Description    *string         `json:"description"`
	LogoURL        *string         `json:"logo_url"`
	PrimaryColor   *string         `json:"primary_color"`
	SecondaryColor *string         `json:"secondary_color"`
	Email          *string         `json:"email"`
	Phone          *string         `json:"phone"`
	Website        *string         `json:"website"`
	SocialLinks    json.RawMessage `json:"social_links"`
}

// ============================================================================
// CLUB MEMBERS
// ============================================================================

// ClubMember represents a member of a club
type ClubMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ClubID    uuid.UUID `json:"club_id" db:"club_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	Position  *string   `json:"position,omitempty" db:"position"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ClubMemberWithUser represents a club member with user details
type ClubMemberWithUser struct {
	ClubMember
	User User `json:"user"`
}

// AddClubMemberRequest represents add member data
type AddClubMemberRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	Role     *string   `json:"role"`
	Position *string   `json:"position"`
}

// UpdateClubMemberRequest represents update member data
type UpdateClubMemberRequest struct {
	Role     *string `json:"role"`
	Position *string `json:"position"`
}

// ============================================================================
// CLUB ANNOUNCEMENTS
// ============================================================================

// ClubAnnouncement represents a club announcement
type ClubAnnouncement struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ClubID    uuid.UUID  `json:"club_id" db:"club_id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	Priority  string     `json:"priority" db:"priority"`
	IsPinned  bool       `json:"is_pinned" db:"is_pinned"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateAnnouncementRequest represents announcement creation data
type CreateAnnouncementRequest struct {
	Title    string  `json:"title" binding:"required,max=500"`
	Content  string  `json:"content" binding:"required"`
	Priority *string `json:"priority"`
	IsPinned *bool   `json:"is_pinned"`
}

// UpdateAnnouncementRequest represents announcement update data
type UpdateAnnouncementRequest struct {
	Title    *string `json:"title" binding:"omitempty,max=500"`
	Content  *string `json:"content"`
	Priority *string `json:"priority"`
	IsPinned *bool   `json:"is_pinned"`
}

// ============================================================================
// CLUB AWARDS
// ============================================================================

// ClubAward represents an award won by a club
type ClubAward struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ClubID         uuid.UUID  `json:"club_id" db:"club_id"`
	AwardName      string     `json:"award_name" db:"award_name"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Position       *string    `json:"position,omitempty" db:"position"`
	PrizeAmount    *float64   `json:"prize_amount,omitempty" db:"prize_amount"`
	EventName      *string    `json:"event_name,omitempty" db:"event_name"`
	AwardedDate    *time.Time `json:"awarded_date,omitempty" db:"awarded_date"`
	CertificateURL *string    `json:"certificate_url,omitempty" db:"certificate_url"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// CreateAwardRequest represents award creation data
type CreateAwardRequest struct {
	AwardName      string     `json:"award_name" binding:"required,max=500"`
	Description    *string    `json:"description"`
	Position       *string    `json:"position"`
	PrizeAmount    *float64   `json:"prize_amount"`
	EventName      *string    `json:"event_name"`
	AwardedDate    *time.Time `json:"awarded_date"`
	CertificateURL *string    `json:"certificate_url"`
}

// ============================================================================
// EVENTS
// ============================================================================

// Event represents an event in the system
type Event struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	Title                string     `json:"title" db:"title"`
	Description          *string    `json:"description,omitempty" db:"description"`
	StartDate            time.Time  `json:"start_date" db:"start_date"`
	EndDate              time.Time  `json:"end_date" db:"end_date"`
	Location             *string    `json:"location,omitempty" db:"location"`
	BannerURL            *string    `json:"banner_url,omitempty" db:"banner_url"`
	Category             *string    `json:"category,omitempty" db:"category"`
	Status               *string    `json:"status,omitempty" db:"status"`
	MaxParticipants      *int       `json:"max_participants,omitempty" db:"max_participants"`
	CurrentParticipants  int        `json:"current_participants" db:"current_participants"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty" db:"registration_deadline"`
	IsFeatured           bool       `json:"is_featured" db:"is_featured"`
	ClubID               *uuid.UUID `json:"club_id,omitempty" db:"club_id"`
	CreatedBy            *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt            *time.Time `json:"-" db:"deleted_at"`
}

// CreateEventRequest represents event creation data
type CreateEventRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	ImageURL    *string    `json:"image_url"`
	BannerURL   *string    `json:"banner_url"`
	StartDate   JSONTime   `json:"start_date" binding:"required"`
	EndDate     JSONTime   `json:"end_date" binding:"required"`
	Location    *string    `json:"location"`
	Category    *string    `json:"category"`
	MaxCapacity *int       `json:"max_capacity"`
	ClubID      *uuid.UUID `json:"club_id"`
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

// ============================================================================
// ANNOUNCEMENTS
// ============================================================================

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
