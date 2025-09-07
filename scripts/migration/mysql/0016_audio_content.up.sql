-- Add audio metadata fields to post_content table
ALTER TABLE p_post_content
ADD COLUMN duration varchar(32) DEFAULT NULL COMMENT 'Audio duration',
ADD COLUMN size varchar(32) DEFAULT NULL COMMENT 'Audio file size'; 