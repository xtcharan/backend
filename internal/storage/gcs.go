package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"mime/multipart"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// GCSStorage implements StorageService for Google Cloud Storage
type GCSStorage struct {
	client     *storage.Client
	bucketName string
	cdnURL     string // Optional CDN URL prefix (e.g., "https://images.yourdomain.com")
}

// NewGCSStorage creates a new GCS storage service
// - client: authenticated GCS client
// - bucketName: the GCS bucket name
// - cdnURL: optional CDN URL (leave empty to use direct GCS URLs)
func NewGCSStorage(client *storage.Client, bucketName string, cdnURL string) *GCSStorage {
	// Normalize CDN URL - remove trailing slash
	cdnURL = strings.TrimSuffix(cdnURL, "/")

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
		cdnURL:     cdnURL,
	}
}

// UploadImage optimizes and uploads an image to GCS
func (s *GCSStorage) UploadImage(ctx context.Context, file multipart.File, filename string, folder string, imageType ImageType) (*UploadResult, error) {
	// 1. Decode the image (supports JPEG, PNG, GIF)
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	// 2. Get original dimensions
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	// 3. Resize if necessary (maintaining aspect ratio)
	maxDim := imageType.MaxDimension()
	finalWidth := originalWidth
	finalHeight := originalHeight

	if originalWidth > maxDim {
		// Resize maintaining aspect ratio
		img = imaging.Resize(img, maxDim, 0, imaging.Lanczos)
		finalWidth = img.Bounds().Dx()
		finalHeight = img.Bounds().Dy()
	}

	// 4. Compress to JPEG (quality 80 - good balance of size and quality)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to compress image: %w", err)
	}

	// 5. Generate unique object path
	uniqueFilename := fmt.Sprintf("%s.jpg", uuid.New().String())
	objectPath := fmt.Sprintf("%s/%s", folder, uniqueFilename)

	// 6. Upload to GCS
	wc := s.client.Bucket(s.bucketName).Object(objectPath).NewWriter(ctx)
	wc.ContentType = "image/jpeg"
	wc.CacheControl = "public, max-age=31536000" // Cache for 1 year (immutable content)

	written, err := io.Copy(wc, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write to bucket: %w", err)
	}
	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close GCS writer: %w", err)
	}

	// 7. Build the public URL
	var publicURL string
	if s.cdnURL != "" {
		// Use CDN URL if configured
		publicURL = fmt.Sprintf("%s/%s", s.cdnURL, objectPath)
	} else {
		// Use direct GCS URL
		publicURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectPath)
	}

	return &UploadResult{
		URL:       publicURL,
		Path:      objectPath,
		SizeBytes: written,
		Width:     finalWidth,
		Height:    finalHeight,
	}, nil
}

// Delete removes a file from GCS
func (s *GCSStorage) Delete(ctx context.Context, path string) error {
	obj := s.client.Bucket(s.bucketName).Object(path)
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			// Already deleted, not an error
			return nil
		}
		return fmt.Errorf("failed to delete object %s: %w", path, err)
	}
	return nil
}
