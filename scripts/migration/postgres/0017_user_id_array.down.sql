-- Drop new indexes
DROP INDEX IF EXISTS idx_post_user_id;
DROP INDEX IF EXISTS idx_post_content_user_id;

-- Revert p_post.user_id to single bigint
ALTER TABLE p_post ALTER COLUMN user_id TYPE bigint USING (user_id->0)::bigint;
ALTER TABLE p_post ALTER COLUMN user_id SET DEFAULT 0;

-- Revert p_post_content.user_id to single bigint
ALTER TABLE p_post_content ALTER COLUMN user_id TYPE bigint USING (user_id->0)::bigint;
ALTER TABLE p_post_content ALTER COLUMN user_id SET DEFAULT 0;

-- Recreate original indexes
CREATE INDEX idx_post_user_id ON p_post (user_id);
CREATE INDEX idx_post_content_user_id ON p_post_content (user_id); 