-- Drop indexes
DROP INDEX IF EXISTS idx_post_room_id;
DROP INDEX IF EXISTS idx_post_content_room_id;

-- Remove room_id from post_content table
ALTER TABLE p_post_content DROP COLUMN IF EXISTS room_id;

-- Remove room_id from post table
ALTER TABLE p_post DROP COLUMN IF EXISTS room_id; 