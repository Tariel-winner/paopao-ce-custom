-- Add location fields to p_post table for Instagram-style location tagging
-- Migration: 0031_add_post_location_fields.up.sql

-- Add location fields to p_post table
ALTER TABLE p_post ADD COLUMN location_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE p_post ADD COLUMN location_lat DECIMAL(10, 8) NOT NULL DEFAULT 0;
ALTER TABLE p_post ADD COLUMN location_lng DECIMAL(11, 8) NOT NULL DEFAULT 0;
ALTER TABLE p_post ADD COLUMN location_address TEXT NOT NULL DEFAULT '';
ALTER TABLE p_post ADD COLUMN location_city VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE p_post ADD COLUMN location_state VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE p_post ADD COLUMN location_country VARCHAR(100) NOT NULL DEFAULT '';

-- Add indexes for better performance on location queries
CREATE INDEX idx_post_location_lat_lng ON p_post (location_lat, location_lng);
CREATE INDEX idx_post_location_name ON p_post (location_name);
CREATE INDEX idx_post_location_city ON p_post (location_city);
CREATE INDEX idx_post_location_country ON p_post (location_country);

-- Add composite index for location-based searches
CREATE INDEX idx_post_location_city_country ON p_post (location_city, location_country);

-- Add comments for documentation
COMMENT ON COLUMN p_post.location_name IS 'Location name (e.g., "Central Park", "Times Square")';
COMMENT ON COLUMN p_post.location_lat IS 'Latitude coordinate (decimal degrees)';
COMMENT ON COLUMN p_post.location_lng IS 'Longitude coordinate (decimal degrees)';
COMMENT ON COLUMN p_post.location_address IS 'Full address string (e.g., "123 Main St, New York, NY 10001")';
COMMENT ON COLUMN p_post.location_city IS 'City name (e.g., "New York")';
COMMENT ON COLUMN p_post.location_state IS 'State/Province name (e.g., "NY", "California")';
COMMENT ON COLUMN p_post.location_country IS 'Country name (e.g., "United States", "Canada")';
