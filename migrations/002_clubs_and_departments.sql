-- Migration: Clubs and Departments System
-- Description: Creates departments, clubs, and related tables with auto-calculation triggers

-- ============================================================================
-- DEPARTMENTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url TEXT,
    icon_name VARCHAR(50),
    color_hex VARCHAR(7) DEFAULT '#4F46E5',
    total_members INT DEFAULT 0,
    total_clubs INT DEFAULT 0,
    total_events INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_departments_code ON departments(code);

-- ============================================================================
-- CLUBS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS clubs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department_id UUID REFERENCES departments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    tagline VARCHAR(500),
    description TEXT,
    logo_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#4F46E5',
    secondary_color VARCHAR(7) DEFAULT '#818CF8',
    member_count INT DEFAULT 0,
    event_count INT DEFAULT 0,
    awards_count INT DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.0,
    email VARCHAR(255),
    phone VARCHAR(20),
    website VARCHAR(255),
    social_links JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_clubs_department ON clubs(department_id);
CREATE INDEX idx_clubs_name ON clubs(name);

-- ============================================================================
-- CLUB MEMBERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS club_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    position VARCHAR(100),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(club_id, user_id)
);

CREATE INDEX idx_club_members_club ON club_members(club_id);
CREATE INDEX idx_club_members_user ON club_members(user_id);
CREATE INDEX idx_club_members_role ON club_members(role);

-- ============================================================================
-- CLUB ANNOUNCEMENTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS club_announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID REFERENCES clubs(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    priority VARCHAR(20) DEFAULT 'normal',
    is_pinned BOOLEAN DEFAULT false,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_club_announcements_club ON club_announcements(club_id);
CREATE INDEX idx_club_announcements_priority ON club_announcements(priority);
CREATE INDEX idx_club_announcements_pinned ON club_announcements(is_pinned);

-- ============================================================================
-- CLUB AWARDS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS club_awards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID REFERENCES clubs(id) ON DELETE CASCADE,
    award_name VARCHAR(500) NOT NULL,
    description TEXT,
    position VARCHAR(100),
    prize_amount DECIMAL(10,2),
    event_name VARCHAR(255),
    awarded_date DATE,
    certificate_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_club_awards_club ON club_awards(club_id);
CREATE INDEX idx_club_awards_date ON club_awards(awarded_date);

-- ============================================================================
-- UPDATE EVENTS TABLE TO LINK WITH CLUBS
-- ============================================================================
ALTER TABLE events ADD COLUMN IF NOT EXISTS club_id UUID REFERENCES clubs(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_events_club ON events(club_id);

-- ============================================================================
-- TRIGGER FUNCTIONS FOR AUTO-CALCULATION
-- ============================================================================

-- Function: Update club member count
CREATE OR REPLACE FUNCTION update_club_member_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE clubs SET member_count = (
            SELECT COUNT(*) FROM club_members WHERE club_id = OLD.club_id
        ) WHERE id = OLD.club_id;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        UPDATE clubs SET member_count = (
            SELECT COUNT(*) FROM club_members WHERE club_id = NEW.club_id
        ) WHERE id = NEW.club_id;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function: Update club awards count
CREATE OR REPLACE FUNCTION update_club_awards_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE clubs SET awards_count = (
            SELECT COUNT(*) FROM club_awards WHERE club_id = OLD.club_id
        ) WHERE id = OLD.club_id;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        UPDATE clubs SET awards_count = (
            SELECT COUNT(*) FROM club_awards WHERE club_id = NEW.club_id
        ) WHERE id = NEW.club_id;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function: Update club event count
CREATE OR REPLACE FUNCTION update_club_event_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        IF OLD.club_id IS NOT NULL THEN
            UPDATE clubs SET event_count = (
                SELECT COUNT(*) FROM events WHERE club_id = OLD.club_id
            ) WHERE id = OLD.club_id;
        END IF;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        IF NEW.club_id IS NOT NULL THEN
            UPDATE clubs SET event_count = (
                SELECT COUNT(*) FROM events WHERE club_id = NEW.club_id
            ) WHERE id = NEW.club_id;
        END IF;
        RETURN NEW;
    ELSIF (TG_OP = 'UPDATE') THEN
        IF OLD.club_id IS DISTINCT FROM NEW.club_id THEN
            IF OLD.club_id IS NOT NULL THEN
                UPDATE clubs SET event_count = (
                    SELECT COUNT(*) FROM events WHERE club_id = OLD.club_id
                ) WHERE id = OLD.club_id;
            END IF;
            IF NEW.club_id IS NOT NULL THEN
                UPDATE clubs SET event_count = (
                    SELECT COUNT(*) FROM events WHERE club_id = NEW.club_id
                ) WHERE id = NEW.club_id;
            END IF;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function: Update department total clubs
CREATE OR REPLACE FUNCTION update_department_clubs_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        IF OLD.department_id IS NOT NULL THEN
            UPDATE departments SET total_clubs = (
                SELECT COUNT(*) FROM clubs WHERE department_id = OLD.department_id
            ) WHERE id = OLD.department_id;
        END IF;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        IF NEW.department_id IS NOT NULL THEN
            UPDATE departments SET total_clubs = (
                SELECT COUNT(*) FROM clubs WHERE department_id = NEW.department_id
            ) WHERE id = NEW.department_id;
        END IF;
        RETURN NEW;
    ELSIF (TG_OP = 'UPDATE') THEN
        IF OLD.department_id IS DISTINCT FROM NEW.department_id THEN
            IF OLD.department_id IS NOT NULL THEN
                UPDATE departments SET total_clubs = (
                    SELECT COUNT(*) FROM clubs WHERE department_id = OLD.department_id
                ) WHERE id = OLD.department_id;
            END IF;
            IF NEW.department_id IS NOT NULL THEN
                UPDATE departments SET total_clubs = (
                    SELECT COUNT(*) FROM clubs WHERE department_id = NEW.department_id
                ) WHERE id = NEW.department_id;
            END IF;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function: Update department total members
CREATE OR REPLACE FUNCTION update_department_members_count()
RETURNS TRIGGER AS $$
DECLARE
    dept_id UUID;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        SELECT department_id INTO dept_id FROM clubs WHERE id = OLD.club_id;
        IF dept_id IS NOT NULL THEN
            UPDATE departments SET total_members = (
                SELECT COUNT(DISTINCT cm.user_id)
                FROM club_members cm
                JOIN clubs c ON cm.club_id = c.id
                WHERE c.department_id = dept_id
            ) WHERE id = dept_id;
        END IF;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        SELECT department_id INTO dept_id FROM clubs WHERE id = NEW.club_id;
        IF dept_id IS NOT NULL THEN
            UPDATE departments SET total_members = (
                SELECT COUNT(DISTINCT cm.user_id)
                FROM club_members cm
                JOIN clubs c ON cm.club_id = c.id
                WHERE c.department_id = dept_id
            ) WHERE id = dept_id;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function: Update department total events
CREATE OR REPLACE FUNCTION update_department_events_count()
RETURNS TRIGGER AS $$
DECLARE
    dept_id UUID;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        IF OLD.club_id IS NOT NULL THEN
            SELECT department_id INTO dept_id FROM clubs WHERE id = OLD.club_id;
            IF dept_id IS NOT NULL THEN
                UPDATE departments SET total_events = (
                    SELECT COUNT(*)
                    FROM events e
                    JOIN clubs c ON e.club_id = c.id
                    WHERE c.department_id = dept_id
                ) WHERE id = dept_id;
            END IF;
        END IF;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        IF NEW.club_id IS NOT NULL THEN
            SELECT department_id INTO dept_id FROM clubs WHERE id = NEW.club_id;
            IF dept_id IS NOT NULL THEN
                UPDATE departments SET total_events = (
                    SELECT COUNT(*)
                    FROM events e
                    JOIN clubs c ON e.club_id = c.id
                    WHERE c.department_id = dept_id
                ) WHERE id = dept_id;
            END IF;
        END IF;
        RETURN NEW;
    ELSIF (TG_OP = 'UPDATE') THEN
        IF OLD.club_id IS DISTINCT FROM NEW.club_id THEN
            IF OLD.club_id IS NOT NULL THEN
                SELECT department_id INTO dept_id FROM clubs WHERE id = OLD.club_id;
                IF dept_id IS NOT NULL THEN
                    UPDATE departments SET total_events = (
                        SELECT COUNT(*)
                        FROM events e
                        JOIN clubs c ON e.club_id = c.id
                        WHERE c.department_id = dept_id
                    ) WHERE id = dept_id;
                END IF;
            END IF;
            IF NEW.club_id IS NOT NULL THEN
                SELECT department_id INTO dept_id FROM clubs WHERE id = NEW.club_id;
                IF dept_id IS NOT NULL THEN
                    UPDATE departments SET total_events = (
                        SELECT COUNT(*)
                        FROM events e
                        JOIN clubs c ON e.club_id = c.id
                        WHERE c.department_id = dept_id
                    ) WHERE id = dept_id;
                END IF;
            END IF;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- CREATE TRIGGERS
-- ============================================================================

-- Club member count triggers
DROP TRIGGER IF EXISTS trigger_update_club_member_count ON club_members;
CREATE TRIGGER trigger_update_club_member_count
    AFTER INSERT OR DELETE ON club_members
    FOR EACH ROW EXECUTE FUNCTION update_club_member_count();

-- Club awards count triggers
DROP TRIGGER IF EXISTS trigger_update_club_awards_count ON club_awards;
CREATE TRIGGER trigger_update_club_awards_count
    AFTER INSERT OR DELETE ON club_awards
    FOR EACH ROW EXECUTE FUNCTION update_club_awards_count();

-- Club event count triggers
DROP TRIGGER IF EXISTS trigger_update_club_event_count ON events;
CREATE TRIGGER trigger_update_club_event_count
    AFTER INSERT OR UPDATE OR DELETE ON events
    FOR EACH ROW EXECUTE FUNCTION update_club_event_count();

-- Department clubs count triggers
DROP TRIGGER IF EXISTS trigger_update_department_clubs_count ON clubs;
CREATE TRIGGER trigger_update_department_clubs_count
    AFTER INSERT OR UPDATE OR DELETE ON clubs
    FOR EACH ROW EXECUTE FUNCTION update_department_clubs_count();

-- Department members count triggers
DROP TRIGGER IF EXISTS trigger_update_department_members_count ON club_members;
CREATE TRIGGER trigger_update_department_members_count
    AFTER INSERT OR DELETE ON club_members
    FOR EACH ROW EXECUTE FUNCTION update_department_members_count();

-- Department events count triggers
DROP TRIGGER IF EXISTS trigger_update_department_events_count ON events;
CREATE TRIGGER trigger_update_department_events_count
    AFTER INSERT OR UPDATE OR DELETE ON events
    FOR EACH ROW EXECUTE FUNCTION update_department_events_count();

-- ============================================================================
-- UPDATED_AT TRIGGERS
-- ============================================================================

DROP TRIGGER IF EXISTS update_departments_updated_at ON departments;
CREATE TRIGGER update_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_clubs_updated_at ON clubs;
CREATE TRIGGER update_clubs_updated_at
    BEFORE UPDATE ON clubs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_club_announcements_updated_at ON club_announcements;
CREATE TRIGGER update_club_announcements_updated_at
    BEFORE UPDATE ON club_announcements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
