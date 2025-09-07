CREATE TABLE p_room (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL DEFAULT 0,
    hms_room_id VARCHAR(255) DEFAULT NULL,
    speaker_ids JSONB DEFAULT NULL,
    start_time BIGINT DEFAULT NULL,
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0,
    is_del SMALLINT NOT NULL DEFAULT 0,
    queue JSONB DEFAULT NULL,
    is_blocked_from_space SMALLINT NOT NULL DEFAULT 0,
    topics JSONB DEFAULT NULL,
    CONSTRAINT fk_room_host FOREIGN KEY (host_id) REFERENCES p_user(id) ON DELETE CASCADE
);

CREATE INDEX idx_room_host ON p_room (host_id);
CREATE INDEX idx_room_hms ON p_room (hms_room_id);
CREATE INDEX idx_room_created ON p_room (created_on); 