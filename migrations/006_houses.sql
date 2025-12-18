-- Migration 006: Houses system with full feature set
-- Houses, roles, announcements, events, and enrollments

-- ============================================================================
-- HOUSES TABLE (Full schema)
-- ============================================================================
CREATE TABLE IF NOT EXISTS houses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(50),
    description TEXT,
    logo_url VARCHAR(500),
    points INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- ============================================================================
-- HOUSE MEMBERS
-- ============================================================================
CREATE TABLE IF NOT EXISTS house_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    house_id UUID NOT NULL REFERENCES houses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_house_members_house ON house_members(house_id);

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

CREATE INDEX IF NOT EXISTS idx_house_announcements_house ON house_announcements(house_id);
CREATE INDEX IF NOT EXISTS idx_house_announcements_created_at ON house_announcements(created_at DESC);

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

CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement ON announcement_comments(announcement_id);

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

CREATE INDEX IF NOT EXISTS idx_house_events_house ON house_events(house_id);
CREATE INDEX IF NOT EXISTS idx_house_events_date ON house_events(event_date);

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

-- ============================================================================
-- TRIGGERS
-- ============================================================================
DROP TRIGGER IF EXISTS update_houses_updated_at ON houses;
CREATE TRIGGER update_houses_updated_at BEFORE UPDATE ON houses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_house_announcements_updated_at ON house_announcements;
CREATE TRIGGER update_house_announcements_updated_at BEFORE UPDATE ON house_announcements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_announcement_comments_updated_at ON announcement_comments;
CREATE TRIGGER update_announcement_comments_updated_at BEFORE UPDATE ON announcement_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_house_events_updated_at ON house_events;
CREATE TRIGGER update_house_events_updated_at BEFORE UPDATE ON house_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
