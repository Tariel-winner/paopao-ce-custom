-- Rollback migration: 0031_add_post_location_fields.down.sql

-- Drop indexes first
DROP INDEX IF EXISTS idx_post_location_city_country;
DROP INDEX IF EXISTS idx_post_location_country;
DROP INDEX IF EXISTS idx_post_location_city;
DROP INDEX IF EXISTS idx_post_location_name;
DROP INDEX IF EXISTS idx_post_location_lat_lng;

-- Remove location columns from p_post table
ALTER TABLE p_post DROP COLUMN IF EXISTS location_country;
ALTER TABLE p_post DROP COLUMN IF EXISTS location_state;
ALTER TABLE p_post DROP COLUMN IF EXISTS location_city;
ALTER TABLE p_post DROP COLUMN IF EXISTS location_address;
ALTER TABLE p_post DROP COLUMN IF EXISTS location_lng;
ALTER TABLE p_post DROP COLUMN IF EXISTS location_lat;
ALTER TABLE p_post DROP COLUMN IF EXISTS location_name;
