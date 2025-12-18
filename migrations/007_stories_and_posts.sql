-- Stories and Posts Feature (Professional Implementation)
-- Posts: Primary feature for announcements, updates, and engagement
-- Stories: Future feature (minimal implementation for 24hr announcements)
-- 
-- Key Features:
-- 1. Posts with image/video support (primary focus)
-- 2. Likes, comments, and shares on posts
-- 3. Stories with automatic 24-hour hard delete
-- 4. Automatic archive storage for old posts (2+ months)
-- 5. Professional cleanup and lifecycle management

-- ============================================================================
-- POSTS TABLE (Primary Feature)
-- ============================================================================

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Creator info (admin/faculty only)
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Optional: Link to club/house for group-specific posts
    club_id UUID REFERENCES clubs(id) ON DELETE SET NULL,
    house_id UUID REFERENCES houses(id) ON DELETE SET NULL,
    
    -- Content (supports text, image, or video)
    content_type VARCHAR(10) NOT NULL DEFAULT 'text', -- 'text', 'image', or 'video'
    image_url TEXT, -- URL to image in GCS (Standard storage)
    video_url TEXT, -- URL to video in GCS (Standard storage)
    thumbnail_url TEXT, -- Thumbnail for videos or image preview
    duration_seconds INTEGER, -- Video duration in seconds (null for images/text)
    description TEXT NOT NULL,
    hashtags TEXT[], -- Array of hashtags (without # prefix) for discovery
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP, -- Soft delete support
    
    -- Storage lifecycle tracking
    archived_at TIMESTAMP, -- When post was moved to Archive storage
    storage_class VARCHAR(20) DEFAULT 'STANDARD', -- 'STANDARD', 'NEARLINE', 'COLDLINE', 'ARCHIVE'
    
    -- Metrics (denormalized for performance)
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    
    -- Constraints
    CHECK (char_length(description) > 0 AND char_length(description) <= 2000),
    CHECK (content_type IN ('text', 'image', 'video')),
    CHECK (
        (content_type = 'text') OR
        (content_type = 'image' AND image_url IS NOT NULL AND video_url IS NULL) OR
        (content_type = 'video' AND video_url IS NOT NULL AND image_url IS NULL AND thumbnail_url IS NOT NULL AND duration_seconds > 0)
    ),
    CHECK (storage_class IN ('STANDARD', 'NEARLINE', 'COLDLINE', 'ARCHIVE'))
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_posts_created_by ON posts(created_by) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_posts_club_id ON posts(club_id) WHERE deleted_at IS NULL AND club_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_posts_house_id ON posts(house_id) WHERE deleted_at IS NULL AND house_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_posts_storage_class ON posts(storage_class, created_at) WHERE deleted_at IS NULL;

-- GIN index for hashtag search
CREATE INDEX IF NOT EXISTS idx_posts_hashtags ON posts USING GIN(hashtags) WHERE deleted_at IS NULL;

-- Index for archive lifecycle management
CREATE INDEX IF NOT EXISTS idx_posts_archive_eligible ON posts(created_at) 
    WHERE deleted_at IS NULL 
    AND archived_at IS NULL 
    AND storage_class = 'STANDARD'
    AND created_at < (CURRENT_TIMESTAMP - INTERVAL '2 months');

-- Trigger for updated_at
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- POST LIKES
-- ============================================================================

CREATE TABLE IF NOT EXISTS post_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- User can only like a post once
    UNIQUE(post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_likes_post ON post_likes(post_id);
CREATE INDEX IF NOT EXISTS idx_post_likes_user ON post_likes(user_id);

-- ============================================================================
-- POST COMMENTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS post_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Comment content
    content TEXT NOT NULL,
    
    -- Optional: Reply support (nested comments)
    parent_comment_id UUID REFERENCES post_comments(id) ON DELETE CASCADE,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Constraints
    CHECK (char_length(content) > 0 AND char_length(content) <= 500)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_post_comments_post ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_post_comments_user ON post_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_post_comments_parent ON post_comments(parent_comment_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_post_comments_created_at ON post_comments(created_at ASC) WHERE deleted_at IS NULL;

-- Trigger for updated_at
CREATE TRIGGER update_post_comments_updated_at BEFORE UPDATE ON post_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- POST SHARES (Track when users share posts)
-- ============================================================================

CREATE TABLE IF NOT EXISTS post_shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Optional: Track share method/platform
    share_method VARCHAR(50) -- 'whatsapp', 'instagram', 'copy_link', 'download', etc.
);

CREATE INDEX IF NOT EXISTS idx_post_shares_post ON post_shares(post_id);
CREATE INDEX IF NOT EXISTS idx_post_shares_user ON post_shares(user_id);

-- ============================================================================
-- POST VIEWS (Track engagement)
-- ============================================================================

CREATE TABLE IF NOT EXISTS post_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- Nullable for anonymous views
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Track view duration for video posts
    duration_watched_seconds INTEGER
);

CREATE INDEX IF NOT EXISTS idx_post_views_post ON post_views(post_id);
CREATE INDEX IF NOT EXISTS idx_post_views_user ON post_views(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_post_views_viewed_at ON post_views(viewed_at DESC);

-- ============================================================================
-- STORIES TABLE (Future Feature - Minimal Implementation)
-- ============================================================================

CREATE TABLE IF NOT EXISTS stories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Creator info
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Optional: Link to club/house
    club_id UUID REFERENCES clubs(id) ON DELETE SET NULL,
    house_id UUID REFERENCES houses(id) ON DELETE SET NULL,
    
    -- Content (image or video only)
    content_type VARCHAR(10) NOT NULL DEFAULT 'image', -- 'image' or 'video'
    image_url TEXT,
    video_url TEXT,
    thumbnail_url TEXT NOT NULL, -- For story circles
    duration_seconds INTEGER,
    description TEXT,
    hashtags TEXT[],
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours'),
    
    -- Metrics
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    
    -- Constraints
    CHECK (expires_at > created_at),
    CHECK (content_type IN ('image', 'video')),
    CHECK (
        (content_type = 'image' AND image_url IS NOT NULL AND video_url IS NULL AND duration_seconds IS NULL) OR
        (content_type = 'video' AND video_url IS NOT NULL AND image_url IS NULL AND duration_seconds IS NOT NULL AND duration_seconds > 0)
    )
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_stories_created_by ON stories(created_by);
CREATE INDEX IF NOT EXISTS idx_stories_expires_at ON stories(expires_at);
CREATE INDEX IF NOT EXISTS idx_stories_active ON stories(created_at DESC) WHERE expires_at > CURRENT_TIMESTAMP;

-- Simple likes for stories
CREATE TABLE IF NOT EXISTS story_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(story_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_story_likes_story ON story_likes(story_id);

-- Simple views for stories
CREATE TABLE IF NOT EXISTS story_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(story_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_story_views_story ON story_views(story_id);

-- ============================================================================
-- FUNCTIONS & TRIGGERS FOR AUTOMATIC COUNT UPDATES
-- ============================================================================

-- Update post like count
CREATE OR REPLACE FUNCTION update_post_like_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE posts SET like_count = like_count + 1 WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE posts SET like_count = like_count - 1 WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_post_like_count
    AFTER INSERT OR DELETE ON post_likes
    FOR EACH ROW EXECUTE FUNCTION update_post_like_count();

-- Update post comment count
CREATE OR REPLACE FUNCTION update_post_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE posts SET comment_count = comment_count - 1 WHERE id = OLD.post_id;
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE' AND OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
        -- Soft delete
        UPDATE posts SET comment_count = comment_count - 1 WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF (TG_OP = 'UPDATE' AND OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL) THEN
        -- Restore
        UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_post_comment_count
    AFTER INSERT OR DELETE OR UPDATE ON post_comments
    FOR EACH ROW EXECUTE FUNCTION update_post_comment_count();

-- Update post share count
CREATE OR REPLACE FUNCTION update_post_share_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE posts SET share_count = share_count + 1 WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE posts SET share_count = share_count - 1 WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_post_share_count
    AFTER INSERT OR DELETE ON post_shares
    FOR EACH ROW EXECUTE FUNCTION update_post_share_count();

-- Update post view count
CREATE OR REPLACE FUNCTION increment_post_view_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE posts SET view_count = view_count + 1 WHERE id = NEW.post_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_post_view_count
    AFTER INSERT ON post_views
    FOR EACH ROW EXECUTE FUNCTION increment_post_view_count();

-- Update story counts (simple)
CREATE OR REPLACE FUNCTION update_story_like_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE stories SET like_count = like_count + 1 WHERE id = NEW.story_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE stories SET like_count = like_count - 1 WHERE id = OLD.story_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_story_like_count
    AFTER INSERT OR DELETE ON story_likes
    FOR EACH ROW EXECUTE FUNCTION update_story_like_count();

CREATE OR REPLACE FUNCTION increment_story_view_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE stories SET view_count = view_count + 1 WHERE id = NEW.story_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_story_view_count
    AFTER INSERT ON story_views
    FOR EACH ROW EXECUTE FUNCTION increment_story_view_count();

-- ============================================================================
-- ADMIN FUNCTIONS FOR HARD DELETE
-- ============================================================================

-- Hard delete a post and its associated data
CREATE OR REPLACE FUNCTION hard_delete_post(post_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete all associated data (cascades will handle most)
    DELETE FROM posts WHERE id = post_uuid;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count > 0;
END;
$$ LANGUAGE plpgsql;

-- Hard delete a story and its associated data
CREATE OR REPLACE FUNCTION hard_delete_story(story_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete all associated data (cascades will handle most)
    DELETE FROM stories WHERE id = story_uuid;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count > 0;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- LIFECYCLE MANAGEMENT FUNCTIONS (For Cron Jobs)
-- ============================================================================

-- Get posts eligible for archive storage (older than 2 months, still in STANDARD storage)
CREATE OR REPLACE FUNCTION get_posts_for_archive()
RETURNS TABLE (
    post_id UUID,
    image_url TEXT,
    video_url TEXT,
    thumbnail_url TEXT,
    age_days INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p.id,
        p.image_url,
        p.video_url,
        p.thumbnail_url,
        EXTRACT(DAY FROM (CURRENT_TIMESTAMP - p.created_at))::INTEGER as age_days
    FROM posts p
    WHERE p.deleted_at IS NULL
        AND p.archived_at IS NULL
        AND p.storage_class = 'STANDARD'
        AND p.created_at < (CURRENT_TIMESTAMP - INTERVAL '2 months')
    ORDER BY p.created_at ASC;
END;
$$ LANGUAGE plpgsql;

-- Mark post as archived (called after moving files to Archive storage)
CREATE OR REPLACE FUNCTION mark_post_as_archived(post_uuid UUID)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE posts 
    SET 
        archived_at = CURRENT_TIMESTAMP,
        storage_class = 'ARCHIVE',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = post_uuid;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Get expired stories for hard deletion
CREATE OR REPLACE FUNCTION get_expired_stories()
RETURNS TABLE (
    story_id UUID,
    image_url TEXT,
    video_url TEXT,
    thumbnail_url TEXT,
    hours_expired NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.id,
        s.image_url,
        s.video_url,
        s.thumbnail_url,
        EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - s.expires_at)) / 3600 as hours_expired
    FROM stories s
    WHERE s.expires_at < CURRENT_TIMESTAMP
    ORDER BY s.expires_at ASC;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

-- Migration adds:
-- ✅ Posts table with image/video support (PRIMARY FEATURE)
-- ✅ Likes, comments, shares, and view tracking for posts
-- ✅ Storage lifecycle tracking (STANDARD → ARCHIVE for old posts)
-- ✅ Stories table (minimal implementation for future use)
-- ✅ Hard delete functions for admin use
-- ✅ Lifecycle management functions for cron jobs
-- ✅ Professional constraints and indexes

-- Cron Job Requirements:
-- 1. Story Cleanup (every hour):
--    - Find expired stories using get_expired_stories()
--    - Delete media files from GCS
--    - Hard delete stories using hard_delete_story()
--
-- 2. Post Archive (daily at 2 AM):
--    - Find old posts using get_posts_for_archive()
--    - Move media files to Archive storage class in GCS
--    - Mark posts as archived using mark_post_as_archived()

-- GCS Lifecycle Policy:
-- {
--   "lifecycle": {
--     "rule": [
--       {
--         "action": {"type": "SetStorageClass", "storageClass": "ARCHIVE"},
--         "condition": {"age": 60, "matchesPrefix": ["posts/"]}
--       },
--       {
--         "action": {"type": "Delete"},
--         "condition": {"age": 1, "matchesPrefix": ["stories/"]}
--       }
--     ]
--   }
-- }
