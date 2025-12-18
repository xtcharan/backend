-- Migration 008: Event Payments
-- Add payment capability to events and track payment transactions

-- ============================================================================
-- ADD PAYMENT COLUMNS TO EVENTS TABLE
-- ============================================================================
ALTER TABLE events ADD COLUMN IF NOT EXISTS is_paid_event BOOLEAN DEFAULT false;
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_amount DECIMAL(10,2) DEFAULT 0.00;
ALTER TABLE events ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'INR';

-- Create index for filtering paid events
CREATE INDEX IF NOT EXISTS idx_events_is_paid ON events(is_paid_event) WHERE deleted_at IS NULL;

-- ============================================================================
-- EVENT PAYMENTS TABLE
-- Tracks all payment transactions for event registrations
-- ============================================================================
CREATE TABLE IF NOT EXISTS event_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    razorpay_order_id VARCHAR(50) NOT NULL,
    razorpay_payment_id VARCHAR(50),
    razorpay_signature VARCHAR(255),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    status VARCHAR(20) DEFAULT 'pending',  -- pending, paid, failed, refunded
    failure_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, user_id, razorpay_order_id)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_event_payments_event ON event_payments(event_id);
CREATE INDEX IF NOT EXISTS idx_event_payments_user ON event_payments(user_id);
CREATE INDEX IF NOT EXISTS idx_event_payments_order ON event_payments(razorpay_order_id);
CREATE INDEX IF NOT EXISTS idx_event_payments_status ON event_payments(status);

-- ============================================================================
-- TRIGGER: Update updated_at on event_payments
-- ============================================================================
DROP TRIGGER IF EXISTS update_event_payments_updated_at ON event_payments;
CREATE TRIGGER update_event_payments_updated_at
    BEFORE UPDATE ON event_payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
