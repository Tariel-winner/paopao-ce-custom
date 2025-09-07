-- Add session_id to post table for better session tracking
ALTER TABLE p_post ADD COLUMN session_id VARCHAR(255);
 
-- Create index for faster lookups
CREATE INDEX idx_post_session_id ON p_post(session_id); 