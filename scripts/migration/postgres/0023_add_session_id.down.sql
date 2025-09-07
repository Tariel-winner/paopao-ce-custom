-- Drop index
DROP INDEX IF EXISTS idx_post_session_id;
 
-- Remove session_id from post table
ALTER TABLE p_post DROP COLUMN IF EXISTS session_id; 