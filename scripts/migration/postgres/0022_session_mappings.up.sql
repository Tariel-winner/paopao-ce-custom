-- Create session_mappings table for peer_id to user_id mapping
CREATE TABLE session_mappings (
    id SERIAL PRIMARY KEY,
    room_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    peer_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(room_id, session_id, peer_id)
);

-- Create indexes for efficient querying
CREATE INDEX idx_session_mappings_room_session ON session_mappings(room_id, session_id);
CREATE INDEX idx_session_mappings_peer_id ON session_mappings(peer_id);
CREATE INDEX idx_session_mappings_user_id ON session_mappings(user_id);
CREATE INDEX idx_session_mappings_created_at ON session_mappings(created_at); 