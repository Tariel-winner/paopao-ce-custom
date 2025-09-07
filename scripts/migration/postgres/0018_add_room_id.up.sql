-- Add room_id to post table
ALTER TABLE p_post ADD COLUMN room_id VARCHAR(255);

-- Add room_id to post_content table
ALTER TABLE p_post_content ADD COLUMN room_id VARCHAR(255);

-- Create index for faster lookups
CREATE INDEX idx_post_room_id ON p_post(room_id);
CREATE INDEX idx_post_content_room_id ON p_post_content(room_id); 