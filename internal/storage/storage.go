package storage

import (
	"context"
	"mime/multipart"
)

// ImageType defines the type of image for size optimization
type ImageType string

const (
	ImageTypeThumbnail ImageType = "thumbnail" // 600px max width (event cards, previews)
	ImageTypeBanner    ImageType = "banner"    // 1080px max width (event banners, headers)
	ImageTypeOriginal  ImageType = "original"  // No resize, just compress
)

// UploadResult contains information about an uploaded image
type UploadResult struct {
	URL       string `json:"url"`        // Public URL (via CDN if configured)
	Path      string `json:"path"`       // Storage path (for deletion)
	SizeBytes int64  `json:"size_bytes"` // Final optimized file size
	Width     int    `json:"width"`      // Final image width
	Height    int    `json:"height"`     // Final image height
}

// StorageService is the interface for cloud storage operations
// Supports GCS, S3, R2, or local filesystem
type StorageService interface {
	// UploadImage optimizes and uploads an image
	// - file: the multipart file from the request
	// - filename: desired filename (will be made unique)
	// - folder: storage folder (e.g., "events", "clubs", "profiles")
	// - imageType: determines resize dimensions
	UploadImage(ctx context.Context, file multipart.File, filename string, folder string, imageType ImageType) (*UploadResult, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, path string) error
}

// MaxDimension returns the maximum width for each image type
func (t ImageType) MaxDimension() int {
	switch t {
	case ImageTypeThumbnail:
		return 600
	case ImageTypeBanner:
		return 1080
	default:
		return 1920 // Reasonable max for "original"
	}
}
