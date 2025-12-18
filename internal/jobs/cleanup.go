package jobs

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/yourusername/college-event-backend/internal/storage"
)

// CleanupService handles automated cleanup and archiving jobs
type CleanupService struct {
	db      *sql.DB
	storage storage.StorageService
	cron    *cron.Cron
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(db *sql.DB, storageService storage.StorageService) *CleanupService {
	return &CleanupService{
		db:      db,
		storage: storageService,
		cron:    cron.New(),
	}
}

// Start starts the cron jobs
func (s *CleanupService) Start() {
	// Story cleanup - every hour at minute 0
	s.cron.AddFunc("0 * * * *", func() {
		if err := s.CleanupExpiredStories(); err != nil {
			log.Printf("[CRON] Story cleanup failed: %v", err)
		} else {
			log.Println("[CRON] Story cleanup completed successfully")
		}
	})

	// Post archiving - daily at 2 AM
	s.cron.AddFunc("0 2 * * *", func() {
		if err := s.ArchiveOldPosts(); err != nil {
			log.Printf("[CRON] Post archiving failed: %v", err)
		} else {
			log.Println("[CRON] Post archiving completed successfully")
		}
	})

	s.cron.Start()
	log.Println("[CRON] Cleanup service started")
}

// Stop stops the cron jobs
func (s *CleanupService) Stop() {
	s.cron.Stop()
	log.Println("[CRON] Cleanup service stopped")
}

// CleanupExpiredStories deletes expired stories and their media files
func (s *CleanupService) CleanupExpiredStories() error {
	ctx := context.Background()
	startTime := time.Now()

	log.Println("[CLEANUP] Starting expired stories cleanup...")

	// Get expired stories
	query := `SELECT * FROM get_expired_stories()`
	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	deletedCount := 0
	failedCount := 0

	for rows.Next() {
		var storyID uuid.UUID
		var imageURL, videoURL, thumbnailURL sql.NullString
		var hoursExpired float64

		err := rows.Scan(&storyID, &imageURL, &videoURL, &thumbnailURL, &hoursExpired)
		if err != nil {
			log.Printf("[CLEANUP] Failed to scan story: %v", err)
			failedCount++
			continue
		}

		// Delete media files from storage
		if imageURL.Valid && imageURL.String != "" {
			path := extractPathFromURL(imageURL.String)
			if err := s.storage.Delete(ctx, path); err != nil {
				log.Printf("[CLEANUP] Failed to delete image %s: %v", path, err)
			}
		}

		if videoURL.Valid && videoURL.String != "" {
			path := extractPathFromURL(videoURL.String)
			if err := s.storage.Delete(ctx, path); err != nil {
				log.Printf("[CLEANUP] Failed to delete video %s: %v", path, err)
			}
		}

		if thumbnailURL.Valid && thumbnailURL.String != "" {
			path := extractPathFromURL(thumbnailURL.String)
			if err := s.storage.Delete(ctx, path); err != nil {
				log.Printf("[CLEANUP] Failed to delete thumbnail %s: %v", path, err)
			}
		}

		// Hard delete story from database
		var success bool
		err = s.db.QueryRow(`SELECT hard_delete_story($1)`, storyID).Scan(&success)
		if err != nil || !success {
			log.Printf("[CLEANUP] Failed to delete story %s from database: %v", storyID, err)
			failedCount++
			continue
		}

		deletedCount++
		log.Printf("[CLEANUP] Deleted story %s (expired %.1f hours ago)", storyID, hoursExpired)
	}

	duration := time.Since(startTime)
	log.Printf("[CLEANUP] Story cleanup complete: %d deleted, %d failed in %.2fs",
		deletedCount, failedCount, duration.Seconds())

	return nil
}

// ArchiveOldPosts moves old posts to Archive storage class
func (s *CleanupService) ArchiveOldPosts() error {
	ctx := context.Background()
	startTime := time.Now()

	log.Println("[ARCHIVE] Starting post archiving (posts older than 60 days)...")

	// Check if storage service supports archiving
	gcsStorage, ok := s.storage.(interface {
		MoveToArchive(ctx context.Context, path string) error
	})
	if !ok {
		log.Println("[ARCHIVE] Storage service doesn't support archiving, skipping")
		return nil
	}

	// Get posts eligible for archiving
	query := `SELECT * FROM get_posts_for_archive()`
	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	archivedCount := 0
	failedCount := 0

	for rows.Next() {
		var postID uuid.UUID
		var imageURL, videoURL, thumbnailURL sql.NullString
		var ageDays int

		err := rows.Scan(&postID, &imageURL, &videoURL, &thumbnailURL, &ageDays)
		if err != nil {
			log.Printf("[ARCHIVE] Failed to scan post: %v", err)
			failedCount++
			continue
		}

		// Move media files to Archive storage class
		mediaMoved := false

		if imageURL.Valid && imageURL.String != "" {
			path := extractPathFromURL(imageURL.String)
			if err := gcsStorage.MoveToArchive(ctx, path); err != nil {
				log.Printf("[ARCHIVE] Failed to archive image %s: %v", path, err)
				failedCount++
				continue
			}
			mediaMoved = true
		}

		if videoURL.Valid && videoURL.String != "" {
			path := extractPathFromURL(videoURL.String)
			if err := gcsStorage.MoveToArchive(ctx, path); err != nil {
				log.Printf("[ARCHIVE] Failed to archive video %s: %v", path, err)
				failedCount++
				continue
			}
			mediaMoved = true
		}

		if thumbnailURL.Valid && thumbnailURL.String != "" {
			path := extractPathFromURL(thumbnailURL.String)
			if err := gcsStorage.MoveToArchive(ctx, path); err != nil {
				log.Printf("[ARCHIVE] Failed to archive thumbnail %s: %v", path, err)
				failedCount++
				continue
			}
			mediaMoved = true
		}

		// Mark post as archived in database
		if mediaMoved {
			var success bool
			err = s.db.QueryRow(`SELECT mark_post_as_archived($1)`, postID).Scan(&success)
			if err != nil || !success {
				log.Printf("[ARCHIVE] Failed to mark post %s as archived: %v", postID, err)
				failedCount++
				continue
			}

			archivedCount++
			log.Printf("[ARCHIVE] Archived post %s (%d days old)", postID, ageDays)
		}
	}

	duration := time.Since(startTime)
	log.Printf("[ARCHIVE] Post archiving complete: %d archived, %d failed in %.2fs",
		archivedCount, failedCount, duration.Seconds())

	return nil
}

// extractPathFromURL extracts the GCS object path from a full URL
// Example: https://storage.googleapis.com/bucket/posts/abc.jpg -> posts/abc.jpg
func extractPathFromURL(url string) string {
	// Handle different URL formats
	// GCS: https://storage.googleapis.com/bucket-name/path/to/file.jpg
	// CDN: https://cdn.example.com/path/to/file.jpg

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return url
	}

	// Find bucket name and path
	for i, part := range parts {
		if part == "storage.googleapis.com" && i+2 < len(parts) {
			// Join everything after bucket name
			return strings.Join(parts[i+2:], "/")
		}
	}

	// Fallback: assume last parts form the path
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}

	return url
}
