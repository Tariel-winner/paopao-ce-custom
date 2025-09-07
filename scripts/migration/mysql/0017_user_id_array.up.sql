-- Modify p_post.user_id to store array of int64
ALTER TABLE p_post MODIFY COLUMN user_id JSON NOT NULL DEFAULT '[0,0]' COMMENT 'Array of user IDs [host_id, visitor_id]';

-- Modify p_post_content.user_id to store array of int64
ALTER TABLE p_post_content MODIFY COLUMN user_id JSON NOT NULL DEFAULT '[0,0]' COMMENT 'Array of user IDs [host_id, visitor_id]';

-- Drop existing indexes
DROP INDEX idx_post_user_id ON p_post;
DROP INDEX idx_post_content_user_id ON p_post_content;

-- Create new indexes using the first element of the array
CREATE INDEX idx_post_user_id ON p_post ((CAST(JSON_EXTRACT(user_id, '$[0]') AS SIGNED)));
CREATE INDEX idx_post_content_user_id ON p_post_content ((CAST(JSON_EXTRACT(user_id, '$[0]') AS SIGNED))); 