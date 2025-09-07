-- Create new tables with array user_id
CREATE TABLE p_post_new (
    id BIGINT PRIMARY KEY,
    user_id JSON NOT NULL DEFAULT '[0,0]',
    comment_count BIGINT NOT NULL DEFAULT 0,
    collection_count BIGINT NOT NULL DEFAULT 0,
    share_count BIGINT NOT NULL DEFAULT 0,
    upvote_count BIGINT NOT NULL DEFAULT 0,
    visibility INTEGER NOT NULL DEFAULT 0,
    is_top INTEGER NOT NULL DEFAULT 0,
    is_essence INTEGER NOT NULL DEFAULT 0,
    is_lock INTEGER NOT NULL DEFAULT 0,
    latest_replied_on BIGINT NOT NULL DEFAULT 0,
    tags TEXT NOT NULL DEFAULT '',
    attachment_price BIGINT NOT NULL DEFAULT 0,
    ip TEXT NOT NULL DEFAULT '',
    ip_loc TEXT NOT NULL DEFAULT '',
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE p_post_content_new (
    id BIGINT PRIMARY KEY,
    post_id BIGINT NOT NULL DEFAULT 0,
    user_id JSON NOT NULL DEFAULT '[0,0]',
    content TEXT NOT NULL DEFAULT '',
    type INTEGER NOT NULL DEFAULT 2,
    sort INTEGER NOT NULL DEFAULT 100,
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0
);

-- Copy data from old tables to new tables
INSERT INTO p_post_new 
SELECT 
    id,
    json_array(user_id, 0),
    comment_count,
    collection_count,
    share_count,
    upvote_count,
    visibility,
    is_top,
    is_essence,
    is_lock,
    latest_replied_on,
    tags,
    attachment_price,
    ip,
    ip_loc,
    created_on,
    modified_on,
    deleted_on
FROM p_post;

INSERT INTO p_post_content_new
SELECT 
    id,
    post_id,
    json_array(user_id, 0),
    content,
    type,
    sort,
    created_on,
    modified_on,
    deleted_on
FROM p_post_content;

-- Drop old tables and rename new tables
DROP TABLE p_post;
DROP TABLE p_post_content;
ALTER TABLE p_post_new RENAME TO p_post;
ALTER TABLE p_post_content_new RENAME TO p_post_content;

-- Create indexes
CREATE INDEX idx_post_user_id ON p_post (json_extract(user_id, '$[0]'));
CREATE INDEX idx_post_visibility ON p_post (visibility);
CREATE INDEX idx_post_content_post_id ON p_post_content (post_id);
CREATE INDEX idx_post_content_user_id ON p_post_content (json_extract(user_id, '$[0]')); 