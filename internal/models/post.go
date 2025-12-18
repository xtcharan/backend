package models

import (
	"time"

	"github.com/google/uuid"
)

// ContentType defines the type of content in a post or story
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
)

// StorageClass defines the GCS storage class for media files
type StorageClass string

const (
	StorageClassStandard StorageClass = "STANDARD"
	StorageClassNearline StorageClass = "NEARLINE"
	StorageClassColdline StorageClass = "COLDLINE"
	StorageClassArchive  StorageClass = "ARCHIVE"
)

// Post represents an announcement post
type Post struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	ClubID    *uuid.UUID `json:"club_id,omitempty" db:"club_id"`
	HouseID   *uuid.UUID `json:"house_id,omitempty" db:"house_id"`

	// Content
	ContentType  ContentType `json:"content_type" db:"content_type"`
	ImageURL     *string     `json:"image_url,omitempty" db:"image_url"`
	VideoURL     *string     `json:"video_url,omitempty" db:"video_url"`
	ThumbnailURL *string     `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	DurationSecs *int        `json:"duration_seconds,omitempty" db:"duration_seconds"`
	Description  string      `json:"description" db:"description"`
	Hashtags     []string    `json:"hashtags" db:"hashtags"`

	// Metadata
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Storage lifecycle
	ArchivedAt   *time.Time   `json:"archived_at,omitempty" db:"archived_at"`
	StorageClass StorageClass `json:"storage_class" db:"storage_class"`

	// Metrics
	LikeCount    int `json:"like_count" db:"like_count"`
	CommentCount int `json:"comment_count" db:"comment_count"`
	ShareCount   int `json:"share_count" db:"share_count"`
	ViewCount    int `json:"view_count" db:"view_count"`
}

// PostResponse is the response DTO with additional user data
type PostResponse struct {
	Post
	Creator      UserSummary `json:"creator"`
	IsLikedByMe  bool        `json:"is_liked_by_me"`
	IsSharedByMe bool        `json:"is_shared_by_me"`
}

// CreatePostRequest is the request to create a new post
type CreatePostRequest struct {
	ClubID       *uuid.UUID  `json:"club_id,omitempty"`
	HouseID      *uuid.UUID  `json:"house_id,omitempty"`
	ContentType  ContentType `json:"content_type" binding:"required,oneof=text image video"`
	ImageURL     *string     `json:"image_url,omitempty"`
	VideoURL     *string     `json:"video_url,omitempty"`
	ThumbnailURL *string     `json:"thumbnail_url,omitempty"`
	DurationSecs *int        `json:"duration_seconds,omitempty"`
	Description  string      `json:"description" binding:"required,min=1,max=2000"`
	Hashtags     []string    `json:"hashtags,omitempty"`
}

// UpdatePostRequest is the request to update an existing post
type UpdatePostRequest struct {
	Description *string   `json:"description,omitempty" binding:"omitempty,min=1,max=2000"`
	Hashtags    *[]string `json:"hashtags,omitempty"`
}

// PostLike represents a user's like on a post
type PostLike struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PostID    uuid.UUID `json:"post_id" db:"post_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PostComment represents a comment on a post
type PostComment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	PostID          uuid.UUID  `json:"post_id" db:"post_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	Content         string     `json:"content" db:"content"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty" db:"parent_comment_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PostCommentResponse is the response DTO with user data
type PostCommentResponse struct {
	PostComment
	User    UserSummary           `json:"user"`
	Replies []PostCommentResponse `json:"replies,omitempty"`
}

// PostShare represents a user sharing a post
type PostShare struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PostID      uuid.UUID `json:"post_id" db:"post_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	SharedAt    time.Time `json:"shared_at" db:"shared_at"`
	ShareMethod *string   `json:"share_method,omitempty" db:"share_method"`
}

// CreateShareRequest is the request to track a share
type CreateShareRequest struct {
	ShareMethod *string `json:"share_method,omitempty" binding:"omitempty,oneof=whatsapp instagram copy_link download"`
}

// PostView represents a view on a post
type PostView struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	PostID              uuid.UUID  `json:"post_id" db:"post_id"`
	UserID              *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ViewedAt            time.Time  `json:"viewed_at" db:"viewed_at"`
	DurationWatchedSecs *int       `json:"duration_watched_seconds,omitempty" db:"duration_watched_seconds"`
}

// Story represents a 24-hour ephemeral story
type Story struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	ClubID    *uuid.UUID `json:"club_id,omitempty" db:"club_id"`
	HouseID   *uuid.UUID `json:"house_id,omitempty" db:"house_id"`

	// Content
	ContentType  ContentType `json:"content_type" db:"content_type"`
	ImageURL     *string     `json:"image_url,omitempty" db:"image_url"`
	VideoURL     *string     `json:"video_url,omitempty" db:"video_url"`
	ThumbnailURL string      `json:"thumbnail_url" db:"thumbnail_url"`
	DurationSecs *int        `json:"duration_seconds,omitempty" db:"duration_seconds"`
	Description  *string     `json:"description,omitempty" db:"description"`
	Hashtags     []string    `json:"hashtags" db:"hashtags"`

	// Metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`

	// Metrics
	ViewCount int `json:"view_count" db:"view_count"`
	LikeCount int `json:"like_count" db:"like_count"`
}

// StoryResponse is the response DTO with additional data
type StoryResponse struct {
	Story
	Creator       UserSummary `json:"creator"`
	IsLikedByMe   bool        `json:"is_liked_by_me"`
	IsViewedByMe  bool        `json:"is_viewed_by_me"`
	TimeRemaining int         `json:"time_remaining_seconds"` // Seconds until expiry
}

// CreateStoryRequest is the request to create a new story
type CreateStoryRequest struct {
	ClubID       *uuid.UUID  `json:"club_id,omitempty"`
	HouseID      *uuid.UUID  `json:"house_id,omitempty"`
	ContentType  ContentType `json:"content_type" binding:"required,oneof=image video"`
	ImageURL     *string     `json:"image_url,omitempty"`
	VideoURL     *string     `json:"video_url,omitempty"`
	ThumbnailURL string      `json:"thumbnail_url" binding:"required"`
	DurationSecs *int        `json:"duration_seconds,omitempty"`
	Description  *string     `json:"description,omitempty" binding:"omitempty,max=500"`
	Hashtags     []string    `json:"hashtags,omitempty"`
}

// StoryLike represents a user's like on a story
type StoryLike struct {
	ID        uuid.UUID `json:"id" db:"id"`
	StoryID   uuid.UUID `json:"story_id" db:"story_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// StoryView represents a user viewing a story
type StoryView struct {
	ID       uuid.UUID `json:"id" db:"id"`
	StoryID  uuid.UUID `json:"story_id" db:"story_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	ViewedAt time.Time `json:"viewed_at" db:"viewed_at"`
}

// UserSummary is a minimal user representation for responses
type UserSummary struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      UserRole  `json:"role"`
}

// ListPostsQuery contains query parameters for listing posts
type ListPostsQuery struct {
	Page     int        `form:"page" binding:"omitempty,min=1"`
	PageSize int        `form:"page_size" binding:"omitempty,min=1,max=100"`
	Hashtag  *string    `form:"hashtag" binding:"omitempty"`
	ClubID   *uuid.UUID `form:"club_id" binding:"omitempty"`
	HouseID  *uuid.UUID `form:"house_id" binding:"omitempty"`
	Search   *string    `form:"q" binding:"omitempty,min=1"`
}

// PostsListResponse is the paginated response for posts
type PostsListResponse struct {
	Posts      []PostResponse `json:"posts"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalCount int            `json:"total_count"`
	TotalPages int            `json:"total_pages"`
}
