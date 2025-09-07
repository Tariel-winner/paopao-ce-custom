-- Add audio metadata fields to post_content table
ALTER TABLE p_post_content
ADD COLUMN duration varchar(32) DEFAULT NULL;
ALTER TABLE p_post_content
ADD COLUMN size varchar(32) DEFAULT NULL; 