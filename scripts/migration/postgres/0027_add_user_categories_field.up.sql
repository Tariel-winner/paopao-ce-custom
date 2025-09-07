-- Add categories field to user table for storing user's category preferences
-- This will store category IDs as an INTEGER array for better performance with integer operations
ALTER TABLE p_user ADD COLUMN categories INTEGER[] DEFAULT '{}';

-- Create index for categories field for better query performance
CREATE INDEX idx_user_categories ON p_user USING GIN (categories); 