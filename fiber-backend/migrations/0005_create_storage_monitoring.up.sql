-- Create storage_alerts table
CREATE TABLE IF NOT EXISTS storage_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(20) NOT NULL,
    component VARCHAR(50) NOT NULL,
    path TEXT NOT NULL,
    used_percent DECIMAL(5,2) NOT NULL,
    free_bytes BIGINT NOT NULL,
    total_bytes BIGINT NOT NULL,
    message TEXT,
    email_sent BOOLEAN DEFAULT false,
    email_sent_at TIMESTAMP,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by VARCHAR(100),
    acknowledged_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create cleanup_actions table
CREATE TABLE IF NOT EXISTS cleanup_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES storage_alerts(id),
    action_type VARCHAR(50) NOT NULL,
    details JSONB,
    success BOOLEAN,
    error_message TEXT,
    freed_bytes BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Index for faster alert lookups
CREATE INDEX IF NOT EXISTS idx_storage_alerts_component ON storage_alerts(component);
CREATE INDEX IF NOT EXISTS idx_storage_alerts_created_at ON storage_alerts(created_at);
