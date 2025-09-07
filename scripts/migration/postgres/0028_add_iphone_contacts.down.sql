-- Rollback: Remove iPhone contacts table
-- This migration removes the p_user_phone_contacts table and all its data

-- Drop the table (this will automatically drop indexes and constraints)
DROP TABLE IF EXISTS p_user_phone_contacts CASCADE;
