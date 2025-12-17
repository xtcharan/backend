-- House system extended schema
-- Extends the basic houses table with full feature set

-- ============================================================================
-- ENHANCE HOUSES TABLE
-- ============================================================================

-- Add new columns to existing houses table
ALTER TABLE houses ADD COLUMN IF NOT EXISTS logo_url VARCHAR(500);
ALTER TABLE houses ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Create trigger for updated_at
CREATE TRIGGER IF NOT EXISTS update_houses_updated_at BEFORE UPDATE ON houses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- HOUSE ROLES (Flexible, admin-defined)
-- ============================================================================

CREATE TABLE IF NOT EXISTS house_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    house_id UUID NOT NULL REFERENCES houses(id) ON DELETE CASCADE,
    member_name VARCHAR(255) NOT NULL,
    role_title VARCHAR(255) NOT NULL,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_house_roles_house ON house_roles(house_id);

-- ============================================================================
-- HOUSE ANNOUNCEMENTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS house_announcements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    house_id UUID NOT NULL REFERENCES houses(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_house_announcements_house ON house_announcements(house_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_house_announcements_created_at ON house_announcements(created_at DESC) WHERE deleted_at IS NULL;

CREATE TRIGGER update_house_announcements_updated_at BEFORE UPDATE ON house_announcements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ANNOUNCEMENT LIKES
-- ============================================================================

CREATE TABLE IF NOT EXISTS announcement_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    announcement_id UUID NOT NULL REFERENCES house_announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(announcement_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_likes_announcement ON announcement_likes(announcement_id);
CREATE INDEX IF NOT EXISTS idx_announcement_likes_user ON announcement_likes(user_id);

-- ============================================================================
-- ANNOUNCEMENT COMMENTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS announcement_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    announcement_id UUID NOT NULL REFERENCES house_announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement ON announcement_comments(announcement_id) WHERE deleted_at IS NULL;

CREATE TRIGGER update_announcement_comments_updated_at BEFORE UPDATE ON announcement_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- HOUSE EVENTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS house_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    house_id UUID NOT NULL REFERENCES houses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    venue VARCHAR(255),
    max_participants INTEGER,
    registration_deadline DATE,
    status VARCHAR(50) DEFAULT 'open',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_house_events_house ON house_events(house_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_house_events_date ON house_events(event_date) WHERE deleted_at IS NULL;

CREATE TRIGGER update_house_events_updated_at BEFORE UPDATE ON house_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- HOUSE EVENT ENROLLMENTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS house_event_enrollments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES house_events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_house_event_enrollments_event ON house_event_enrollments(event_id);
CREATE INDEX IF NOT EXISTS idx_house_event_enrollments_user ON house_event_enrollments(user_id);
