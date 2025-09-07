-- Note: SQLite doesn't support DROP COLUMN directly
-- We need to create a new table without these columns and copy data
CREATE TABLE p_post_content_new (
    id BIGINT PRIMARY KEY AUTOINCREMENT,
    post_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL DEFAULT 0,
    content varchar(4000) NOT NULL DEFAULT '',
    type tinyint NOT NULL DEFAULT 2,
    sort int NOT NULL DEFAULT 100,
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0,
    is_del tinyint NOT NULL DEFAULT 0
);

-- Copy data from old table to new table
INSERT INTO p_post_content_new 
SELECT id, post_id, user_id, content, type, sort, created_on, modified_on, deleted_on, is_del 
FROM p_post_content;

-- Drop old table
DROP TABLE p_post_content;

-- Rename new table to original name
ALTER TABLE p_post_content_new RENAME TO p_post_content; 