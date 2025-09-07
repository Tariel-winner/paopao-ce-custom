-- Modify p_post.user_id to store array of int64
ALTER TABLE p_post ALTER COLUMN user_id DROP DEFAULT;
ALTER TABLE p_post ALTER COLUMN user_id TYPE jsonb USING to_jsonb(ARRAY[user_id, 0]::bigint[]);
ALTER TABLE p_post ALTER COLUMN user_id SET DEFAULT to_jsonb(ARRAY[0, 0]::bigint[]);

-- Modify p_post_content.user_id to store array of int64
ALTER TABLE p_post_content ALTER COLUMN user_id DROP DEFAULT;
ALTER TABLE p_post_content ALTER COLUMN user_id TYPE jsonb USING to_jsonb(ARRAY[user_id, 0]::bigint[]);
ALTER TABLE p_post_content ALTER COLUMN user_id SET DEFAULT to_jsonb(ARRAY[0, 0]::bigint[]);

-- Drop existing indexes
DROP INDEX IF EXISTS idx_post_user_id;
DROP INDEX IF EXISTS idx_post_content_user_id;

-- Create new indexes using the first element of the array
CREATE INDEX idx_post_user_id ON p_post (CAST(user_id->0 AS bigint));
CREATE INDEX idx_post_content_user_id ON p_post_content (CAST(user_id->0 AS bigint)); 