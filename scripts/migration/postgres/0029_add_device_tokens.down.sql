-- Rollback: Remove device tokens table
-- This migration removes the p_user_device_tokens table and all its data

-- Drop the table (this will automatically drop indexes and constraints)
DROP TABLE IF EXISTS p_user_device_tokens CASCADE;
