-- Remove categories field from user table
DROP INDEX IF EXISTS idx_user_categories;
ALTER TABLE p_user DROP COLUMN IF EXISTS categories; 