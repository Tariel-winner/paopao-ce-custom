-- Add categories field directly to p_room table for faster integer array queries
ALTER TABLE p_room ADD COLUMN categories INTEGER[] DEFAULT '{}';

-- Create index for efficient category-based queries
CREATE INDEX idx_room_categories ON p_room USING GIN (categories);

-- Add comment explaining the field
COMMENT ON COLUMN p_room.categories IS 'Array of category IDs that this room belongs to, for prioritization in room listing'; 