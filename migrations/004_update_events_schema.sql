-- Migration: Update events table to match API requirements
-- Adds missing columns that the events handler expects

-- Rename image_url to banner_url
ALTER TABLE events RENAME COLUMN image_url TO banner_url;

-- Rename max_capacity to max_participants
ALTER TABLE events RENAME COLUMN max_capacity TO max_participants;

-- Add missing columns
ALTER TABLE events ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'upcoming';
ALTER TABLE events ADD COLUMN IF NOT EXISTS current_participants INTEGER DEFAULT 0;
ALTER TABLE events ADD COLUMN IF NOT EXISTS registration_deadline TIMESTAMP;
ALTER TABLE events ADD COLUMN IF NOT EXISTS is_featured BOOLEAN DEFAULT FALSE;
ALTER TABLE events ADD COLUMN IF NOT EXISTS club_id UUID REFERENCES clubs(id) ON DELETE SET NULL;

-- Create index on club_id for better query performance
CREATE INDEX IF NOT EXISTS idx_events_club_id ON events(club_id) WHERE deleted_at IS NULL;

-- Create index on status for filtering
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status) WHERE deleted_at IS NULL;

-- Create index on is_featured for featured events queries
CREATE INDEX IF NOT EXISTS idx_events_is_featured ON events(is_featured) WHERE deleted_at IS NULL AND is_featured = TRUE;
