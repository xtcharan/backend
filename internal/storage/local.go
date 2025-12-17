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
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// LocalStorage implements StorageService for local filesystem
// Useful for development and testing without cloud credentials
type LocalStorage struct {
	basePath string // Base directory for uploads (e.g., "./uploads")
	baseURL  string // Base URL for serving files (e.g., "http://localhost:8080/uploads")
}

// NewLocalStorage creates a new local filesystem storage service
func NewLocalStorage(basePath string, baseURL string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// UploadImage optimizes and saves an image to the local filesystem
func (s *LocalStorage) UploadImage(ctx context.Context, file multipart.File, filename string, folder string, imageType ImageType) (*UploadResult, error) {
	// 1. Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	// 2. Resize if necessary
	maxDim := imageType.MaxDimension()
	finalWidth := img.Bounds().Dx()
	finalHeight := img.Bounds().Dy()

	if finalWidth > maxDim {
		img = imaging.Resize(img, maxDim, 0, imaging.Lanczos)
		finalWidth = img.Bounds().Dx()
		finalHeight = img.Bounds().Dy()
	}

	// 3. Compress to JPEG
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to compress image: %w", err)
	}

	// 4. Create directory structure
	folderPath := filepath.Join(s.basePath, folder)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 5. Generate unique filename and save
	uniqueFilename := fmt.Sprintf("%s.jpg", uuid.New().String())
	filePath := filepath.Join(folderPath, uniqueFilename)

	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 6. Build the public URL
	relativePath := fmt.Sprintf("%s/%s", folder, uniqueFilename)
	publicURL := fmt.Sprintf("%s/%s", s.baseURL, relativePath)

	return &UploadResult{
		URL:       publicURL,
		Path:      relativePath,
		SizeBytes: written,
		Width:     finalWidth,
		Height:    finalHeight,
	}, nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			// Already deleted, not an error
			return nil
		}
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}
