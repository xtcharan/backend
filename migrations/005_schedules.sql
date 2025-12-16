-- Schedules table for daily timetable items
-- Supports both official (admin-created) and personal (student-created) schedules

CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    schedule_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME,
    location VARCHAR(255),
    schedule_type VARCHAR(50) NOT NULL DEFAULT 'personal', -- 'official' or 'personal'
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for official schedules, user_id for personal
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_schedules_date ON schedules(schedule_date);
CREATE INDEX idx_schedules_user ON schedules(user_id);
CREATE INDEX idx_schedules_type ON schedules(schedule_type);
CREATE INDEX idx_schedules_date_type ON schedules(schedule_date, schedule_type);

-- Trigger to auto-update updated_at
CREATE TRIGGER update_schedules_updated_at BEFORE UPDATE ON schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
