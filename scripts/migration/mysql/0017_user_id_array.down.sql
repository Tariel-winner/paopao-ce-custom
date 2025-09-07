-- Drop new indexes
DROP INDEX idx_post_user_id ON p_post;
DROP INDEX idx_post_content_user_id ON p_post_content;

-- Modify p_post.user_id back to bigint
ALTER TABLE p_post MODIFY COLUMN user_id BIGINT NOT NULL DEFAULT 0 COMMENT 'Host user ID';

-- Modify p_post_content.user_id back to bigint
ALTER TABLE p_post_content MODIFY COLUMN user_id BIGINT NOT NULL DEFAULT 0 COMMENT 'Host user ID';

-- Create original indexes
CREATE INDEX idx_post_user_id ON p_post (user_id);
CREATE INDEX idx_post_content_user_id ON p_post_content (user_id); 