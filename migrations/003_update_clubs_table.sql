-- Update existing clubs table to match new schema

-- Drop old department column (string) if exists
ALTER TABLE clubs DROP COLUMN IF EXISTS department;

-- Add new columns if they don't exist
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS department_id UUID REFERENCES departments(id) ON DELETE CASCADE;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS tagline VARCHAR(500);
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS logo_url TEXT;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS primary_color VARCHAR(7) DEFAULT '#4F46E5';
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS secondary_color VARCHAR(7) DEFAULT '#818CF8';
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS member_count INT DEFAULT 0;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS event_count INT DEFAULT 0;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS awards_count INT DEFAULT 0;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS rating DECIMAL(3,2) DEFAULT 0.0;
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS website VARCHAR(255);
ALTER TABLE clubs ADD COLUMN IF NOT EXISTS social_links JSONB;

-- Drop old image_url if exists and we have logo_url
ALTER TABLE clubs DROP COLUMN IF EXISTS image_url;

-- Recreate index for department_id
DROP INDEX IF EXISTS idx_clubs_department;
CREATE INDEX idx_clubs_department ON clubs(department_id);
