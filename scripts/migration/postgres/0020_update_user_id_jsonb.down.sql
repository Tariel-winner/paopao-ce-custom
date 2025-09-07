-- Revert p_post_star table
ALTER TABLE p_post_star ALTER COLUMN user_id TYPE bigint USING CAST(user_id->0 AS bigint);
ALTER TABLE p_post_star ALTER COLUMN user_id SET DEFAULT 0;
DROP INDEX IF EXISTS idx_post_star_user_id;
CREATE INDEX idx_post_star_user_id ON p_post_star (user_id);

-- Revert p_post_collection table
ALTER TABLE p_post_collection ALTER COLUMN user_id TYPE bigint USING CAST(user_id->0 AS bigint);
ALTER TABLE p_post_collection ALTER COLUMN user_id SET DEFAULT 0;
DROP INDEX IF EXISTS idx_post_collection_user_id;
CREATE INDEX idx_post_collection_user_id ON p_post_collection (user_id);

-- Revert p_post_attachment_bill table
ALTER TABLE p_post_attachment_bill ALTER COLUMN user_id TYPE bigint USING CAST(user_id->0 AS bigint);
ALTER TABLE p_post_attachment_bill ALTER COLUMN user_id SET DEFAULT 0;
DROP INDEX IF EXISTS idx_post_attachment_bill_user_id;
CREATE INDEX idx_post_attachment_bill_user_id ON p_post_attachment_bill (user_id);

-- Recreate original views
DROP VIEW IF EXISTS p_post_by_media;
CREATE VIEW p_post_by_media AS 
SELECT post.* 
FROM
    ( SELECT DISTINCT post_id FROM p_post_content WHERE ( TYPE = 3 OR TYPE = 4 OR TYPE = 7 OR TYPE = 8 ) AND is_del = 0 ) media
    JOIN p_post post ON media.post_id = post.ID 
WHERE
    post.is_del = 0;

DROP VIEW IF EXISTS p_post_by_comment;
CREATE VIEW p_post_by_comment AS 
SELECT P.*, C.user_id comment_user_id
FROM
    (
    SELECT
        post_id,
        user_id
    FROM
        p_comment 
    WHERE
        is_del = 0 UNION
    SELECT
        post_id,
        reply.user_id user_id
    FROM
        p_comment_reply reply
        JOIN p_comment COMMENT ON reply.comment_id = COMMENT.ID 
    WHERE
        reply.is_del = 0 
        AND COMMENT.is_del = 0 
    )
    C JOIN p_post P ON C.post_id = P.ID 
WHERE
    P.is_del = 0; 