-- Remove audio metadata fields from post_content table
ALTER TABLE p_post_content
DROP COLUMN duration,
DROP COLUMN size; 