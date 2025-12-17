package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/college-event-backend/internal/models"
	"github.com/yourusername/college-event-backend/internal/storage"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	storage storage.StorageService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(s storage.StorageService) *UploadHandler {
	return &UploadHandler{storage: s}
}

// UploadImage handles image upload with optimization
// POST /api/v1/admin/upload
// Form fields:
//   - file: the image file (required, max 10MB)
//   - folder: storage folder - "events", "clubs", "profiles" (optional, default: "misc")
//   - type: image type - "thumbnail", "banner", "original" (optional, default: "banner")
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 1. Limit request body size to 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	// 2. Parse the multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, models.APIResponse{
				Success: false,
				Error:   strPtr("File too large. Maximum size is 10MB"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("No file provided or invalid form data"),
		})
		return
	}
	defer file.Close()

	// 3. Validate content type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("Invalid file type. Allowed: JPEG, PNG, GIF, WebP"),
		})
		return
	}

	// 4. Get folder (default to "misc")
	folder := c.PostForm("folder")
	if folder == "" {
		folder = "misc"
	}
	// Sanitize folder name
	folder = sanitizeFolderName(folder)

	// 5. Get image type for sizing
	imageTypeStr := c.PostForm("type")
	imageType := storage.ImageTypeBanner // Default to banner size
	switch imageTypeStr {
	case "thumbnail":
		imageType = storage.ImageTypeThumbnail
	case "original":
		imageType = storage.ImageTypeOriginal
	case "banner", "":
		imageType = storage.ImageTypeBanner
	}

	// 6. Upload via storage service
	result, err := h.storage.UploadImage(c.Request.Context(), file, header.Filename, folder, imageType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("Failed to upload image: " + err.Error()),
		})
		return
	}

	// 7. Return success response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Image uploaded successfully",
		Data: gin.H{
			"url":        result.URL,
			"path":       result.Path,
			"size_bytes": result.SizeBytes,
			"width":      result.Width,
			"height":     result.Height,
		},
	})
}

// isValidImageType checks if the content type is an allowed image format
func isValidImageType(contentType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	for _, t := range allowed {
		if contentType == t {
			return true
		}
	}
	return false
}

// sanitizeFolderName ensures folder names are safe
func sanitizeFolderName(folder string) string {
	// Only allow alphanumeric, underscore, and hyphen
	allowed := "abcdefghijklmnopqrstuvwxyz0123456789_-"
	folder = strings.ToLower(folder)

	var result strings.Builder
	for _, char := range folder {
		if strings.ContainsRune(allowed, char) {
			result.WriteRune(char)
		}
	}

	if result.Len() == 0 {
		return "misc"
	}
	return result.String()
}
