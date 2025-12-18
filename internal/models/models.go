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
	// Payment fields
	IsPaidEvent bool       `json:"is_paid_event" db:"is_paid_event"`
	EventAmount *float64   `json:"event_amount,omitempty" db:"event_amount"`
	Currency    *string    `json:"currency,omitempty" db:"currency"`
	ClubID      *uuid.UUID `json:"club_id,omitempty" db:"club_id"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
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
	// Payment fields
	IsPaidEvent bool     `json:"is_paid_event"`
	EventAmount *float64 `json:"event_amount"`
	Currency    *string  `json:"currency"`
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
		"2006-01-02T15:04:05.000Z",      // Flutter format with milliseconds
		"2006-01-02T15:04:05Z",          // ISO8601 without milliseconds
		"2006-01-02T15:04:05.000Z07:00", // RFC3339 with milliseconds
		"2006-01-02T15:04:05Z07:00",     // RFC3339
		"2006-01-02T15:04:05",           // Date and time only
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
// PAYMENTS
// ============================================================================

// EventPayment represents a payment transaction for event registration
type EventPayment struct {
	ID                uuid.UUID `json:"id" db:"id"`
	EventID           uuid.UUID `json:"event_id" db:"event_id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	RazorpayOrderID   string    `json:"razorpay_order_id" db:"razorpay_order_id"`
	RazorpayPaymentID *string   `json:"razorpay_payment_id,omitempty" db:"razorpay_payment_id"`
	RazorpaySignature *string   `json:"razorpay_signature,omitempty" db:"razorpay_signature"`
	Amount            float64   `json:"amount" db:"amount"`
	Currency          string    `json:"currency" db:"currency"`
	Status            string    `json:"status" db:"status"` // pending, paid, failed, refunded
	FailureReason     *string   `json:"failure_reason,omitempty" db:"failure_reason"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrderRequest represents request to create a Razorpay order
type CreateOrderRequest struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
}

// CreateOrderResponse represents response after creating a Razorpay order
type CreateOrderResponse struct {
	OrderID  string `json:"order_id"`
	Amount   int    `json:"amount"` // Amount in paise
	Currency string `json:"currency"`
	KeyID    string `json:"key_id"`
	EventID  string `json:"event_id"`
}

// VerifyPaymentRequest represents request to verify a payment
type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string `json:"razorpay_signature" binding:"required"`
	EventID           string `json:"event_id" binding:"required"`
}

// PaymentStatusResponse represents payment status for an event
type PaymentStatusResponse struct {
	HasPaid   bool    `json:"has_paid"`
	PaymentID *string `json:"payment_id,omitempty"`
	Status    *string `json:"status,omitempty"`
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

// ============================================================================
// SCHEDULES
// ============================================================================

// TimeString is a custom type for time-only values that serializes to "HH:MM" format
type TimeString time.Time

func (t TimeString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format("15:04"))), nil
}

func (t *TimeString) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = str[1 : len(str)-1] // remove quotes
	parsed, err := time.Parse("15:04", str)
	if err != nil {
		return err
	}
	*t = TimeString(parsed)
	return nil
}

func (t *TimeString) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = TimeString(v)
	case string:
		parsed, err := time.Parse("15:04:05", v)
		if err != nil {
			return err
		}
		*t = TimeString(parsed)
	}
	return nil
}

// Schedule represents a daily schedule item (official or personal)
type Schedule struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	Title        string      `json:"title" db:"title"`
	Description  *string     `json:"description,omitempty" db:"description"`
	ScheduleDate time.Time   `json:"schedule_date" db:"schedule_date"`
	StartTime    TimeString  `json:"start_time" db:"start_time"`
	EndTime      *TimeString `json:"end_time,omitempty" db:"end_time"`
	Location     *string     `json:"location,omitempty" db:"location"`
	ScheduleType string      `json:"schedule_type" db:"schedule_type"` // 'official' or 'personal'
	CreatedBy    uuid.UUID   `json:"created_by" db:"created_by"`
	UserID       *uuid.UUID  `json:"user_id,omitempty" db:"user_id"` // null for official schedules
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// CreateScheduleRequest represents schedule creation data
type CreateScheduleRequest struct {
	Title        string  `json:"title" binding:"required,max=255"`
	Description  *string `json:"description"`
	ScheduleDate string  `json:"schedule_date" binding:"required"` // YYYY-MM-DD format
	StartTime    string  `json:"start_time" binding:"required"`    // HH:MM format
	EndTime      *string `json:"end_time"`
	Location     *string `json:"location"`
	ScheduleType string  `json:"schedule_type"` // 'official' or 'personal', defaults to 'personal'
}

// UpdateScheduleRequest represents schedule update data
type UpdateScheduleRequest struct {
	Title        *string `json:"title" binding:"omitempty,max=255"`
	Description  *string `json:"description"`
	ScheduleDate *string `json:"schedule_date"` // YYYY-MM-DD format
	StartTime    *string `json:"start_time"`    // HH:MM format
	EndTime      *string `json:"end_time"`
	Location     *string `json:"location"`
}

// ============================================================================
// HOUSES
// ============================================================================

// House represents a house in the school house system
type House struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Color       *string    `json:"color,omitempty" db:"color"`
	Description *string    `json:"description,omitempty" db:"description"`
	LogoURL     *string    `json:"logo_url,omitempty" db:"logo_url"`
	Points      int        `json:"points" db:"points"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
	// Computed fields (not in DB)
	Roles []HouseRole `json:"roles,omitempty"`
}

// CreateHouseRequest represents house creation data
type CreateHouseRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

// UpdateHouseRequest represents house update data
type UpdateHouseRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
	Points      *int    `json:"points"`
}

// ============================================================================
// HOUSE ROLES
// ============================================================================

// HouseRole represents a role/position in a house (admin-defined)
type HouseRole struct {
	ID           uuid.UUID `json:"id" db:"id"`
	HouseID      uuid.UUID `json:"house_id" db:"house_id"`
	MemberName   string    `json:"member_name" db:"member_name"`
	RoleTitle    string    `json:"role_title" db:"role_title"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// CreateHouseRoleRequest represents house role creation data
type CreateHouseRoleRequest struct {
	MemberName   string `json:"member_name" binding:"required,max=255"`
	RoleTitle    string `json:"role_title" binding:"required,max=255"`
	DisplayOrder *int   `json:"display_order"`
}

// ============================================================================
// HOUSE ANNOUNCEMENTS
// ============================================================================

// HouseAnnouncement represents an announcement in a house
type HouseAnnouncement struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	HouseID   uuid.UUID  `json:"house_id" db:"house_id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
	// Computed fields
	AuthorName   string `json:"author_name,omitempty"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	IsLikedByMe  bool   `json:"is_liked_by_me"`
}

// CreateHouseAnnouncementRequest represents announcement creation data
type CreateHouseAnnouncementRequest struct {
	Title   string `json:"title" binding:"required,max=500"`
	Content string `json:"content" binding:"required"`
}

// UpdateHouseAnnouncementRequest represents announcement update data
type UpdateHouseAnnouncementRequest struct {
	Title   *string `json:"title" binding:"omitempty,max=500"`
	Content *string `json:"content"`
}

// ============================================================================
// ANNOUNCEMENT LIKES & COMMENTS
// ============================================================================

// AnnouncementLike represents a like on an announcement
type AnnouncementLike struct {
	ID             uuid.UUID `json:"id" db:"id"`
	AnnouncementID uuid.UUID `json:"announcement_id" db:"announcement_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// AnnouncementComment represents a comment on an announcement
type AnnouncementComment struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	AnnouncementID uuid.UUID  `json:"announcement_id" db:"announcement_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Content        string     `json:"content" db:"content"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"-" db:"deleted_at"`
	// Computed fields
	UserName  string `json:"user_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// CreateCommentRequest represents comment creation data
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
}

// ============================================================================
// HOUSE EVENTS
// ============================================================================

// HouseEvent represents an event specific to a house
type HouseEvent struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	HouseID              uuid.UUID  `json:"house_id" db:"house_id"`
	Title                string     `json:"title" db:"title"`
	Description          *string    `json:"description,omitempty" db:"description"`
	EventDate            time.Time  `json:"event_date" db:"event_date"`
	StartTime            *string    `json:"start_time,omitempty" db:"start_time"`
	EndTime              *string    `json:"end_time,omitempty" db:"end_time"`
	Venue                *string    `json:"venue,omitempty" db:"venue"`
	MaxParticipants      *int       `json:"max_participants,omitempty" db:"max_participants"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty" db:"registration_deadline"`
	Status               string     `json:"status" db:"status"` // open, closed, completed
	CreatedBy            *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt            *time.Time `json:"-" db:"deleted_at"`
	// Computed fields
	EnrollmentCount int  `json:"enrollment_count"`
	IsEnrolled      bool `json:"is_enrolled"`
}

// CreateHouseEventRequest represents house event creation data
type CreateHouseEventRequest struct {
	Title                string  `json:"title" binding:"required,max=255"`
	Description          *string `json:"description"`
	EventDate            string  `json:"event_date" binding:"required"` // YYYY-MM-DD format
	StartTime            *string `json:"start_time"`                    // HH:MM format
	EndTime              *string `json:"end_time"`
	Venue                *string `json:"venue"`
	MaxParticipants      *int    `json:"max_participants"`
	RegistrationDeadline *string `json:"registration_deadline"` // YYYY-MM-DD format
}

// UpdateHouseEventRequest represents house event update data
type UpdateHouseEventRequest struct {
	Title                *string `json:"title" binding:"omitempty,max=255"`
	Description          *string `json:"description"`
	EventDate            *string `json:"event_date"` // YYYY-MM-DD format
	StartTime            *string `json:"start_time"`
	EndTime              *string `json:"end_time"`
	Venue                *string `json:"venue"`
	MaxParticipants      *int    `json:"max_participants"`
	RegistrationDeadline *string `json:"registration_deadline"`
	Status               *string `json:"status"`
}

// HouseEventEnrollment represents a user enrollment in a house event
type HouseEventEnrollment struct {
	ID         uuid.UUID `json:"id" db:"id"`
	EventID    uuid.UUID `json:"event_id" db:"event_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	EnrolledAt time.Time `json:"enrolled_at" db:"enrolled_at"`
}
