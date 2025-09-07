-- Migration: Add device tokens table for storing iOS/Android push notification tokens
-- This allows users to receive push notifications on multiple devices

CREATE TABLE p_user_device_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,                    -- which user owns this device
    device_token VARCHAR(255) NOT NULL,         -- push notification token from device
    platform VARCHAR(10) NOT NULL,              -- 'ios' or 'android'
    device_id VARCHAR(100) NOT NULL,            -- unique device identifier
    device_name VARCHAR(100) DEFAULT '',        -- user-friendly device name (e.g., "iPhone 15")
    is_active BOOLEAN DEFAULT true,             -- whether this device should receive notifications
    last_used_on BIGINT DEFAULT 0,             -- when this device was last used
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0,
    is_del SMALLINT NOT NULL DEFAULT 0,
    
    -- Foreign key to p_user
    CONSTRAINT fk_user_device_tokens_user 
        FOREIGN KEY (user_id) REFERENCES p_user(id),
    
    -- Ensure platform is valid
    CONSTRAINT chk_platform_valid 
        CHECK (platform IN ('ios', 'android')),
    
    -- Ensure device token is unique per user per device
    CONSTRAINT uk_user_device_token 
        UNIQUE (user_id, device_token)
);

-- Indexes for performance
CREATE INDEX idx_user_device_tokens_user_id ON p_user_device_tokens(user_id);
CREATE INDEX idx_user_device_tokens_platform ON p_user_device_tokens(platform);
CREATE INDEX idx_user_device_tokens_active ON p_user_device_tokens(is_active, is_del);
CREATE INDEX idx_user_device_tokens_device_id ON p_user_device_tokens(device_id);
CREATE INDEX idx_user_device_tokens_token ON p_user_device_tokens(device_token);

-- Add comment for documentation
COMMENT ON TABLE p_user_device_tokens IS 'Stores device tokens for push notifications (iOS/Android)';
COMMENT ON COLUMN p_user_device_tokens.user_id IS 'App user who owns this device';
COMMENT ON COLUMN p_user_device_tokens.device_token IS 'Push notification token from device (APNS/FCM)';
COMMENT ON COLUMN p_user_device_tokens.platform IS 'Device platform: ios or android';
COMMENT ON COLUMN p_user_device_tokens.device_id IS 'Unique device identifier for tracking';
COMMENT ON COLUMN p_user_device_tokens.device_name IS 'User-friendly device name';
COMMENT ON COLUMN p_user_device_tokens.is_active IS 'Whether this device should receive notifications';
COMMENT ON COLUMN p_user_device_tokens.last_used_on IS 'Timestamp when device was last used';
