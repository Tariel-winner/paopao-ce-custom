-- Migration: Add iPhone contacts table for storing user's phone contacts
-- This allows matching iPhone contacts with app users for push notifications

CREATE TABLE p_user_phone_contacts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,                    -- who uploaded these contacts
    contact_name VARCHAR(100) NOT NULL,         -- "John Doe" from iPhone
    contact_phone VARCHAR(20) NOT NULL,         -- "+1234567890" from iPhone
    contact_email VARCHAR(100) DEFAULT '',      -- "john@email.com" from iPhone
    is_matched BOOLEAN DEFAULT false,           -- found matching app user?
    matched_user_id BIGINT DEFAULT NULL,        -- if matched, link to p_user.id
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0,
    is_del SMALLINT NOT NULL DEFAULT 0,
    
    -- Foreign key to p_user (who uploaded contacts)
    CONSTRAINT fk_user_phone_contacts_user 
        FOREIGN KEY (user_id) REFERENCES p_user(id),
    
    -- Foreign key to matched app user (if found)
    CONSTRAINT fk_user_phone_contacts_matched_user 
        FOREIGN KEY (matched_user_id) REFERENCES p_user(id)
);

-- Indexes for performance
CREATE INDEX idx_user_phone_contacts_user_id ON p_user_phone_contacts(user_id);
CREATE INDEX idx_user_phone_contacts_phone ON p_user_phone_contacts(contact_phone);
CREATE INDEX idx_user_phone_contacts_matched ON p_user_phone_contacts(is_matched, matched_user_id);
CREATE INDEX idx_user_phone_contacts_is_del ON p_user_phone_contacts(is_del);

-- Add comment for documentation
COMMENT ON TABLE p_user_phone_contacts IS 'Stores iPhone contacts uploaded by users for contact discovery and push notifications';
COMMENT ON COLUMN p_user_phone_contacts.user_id IS 'App user who uploaded these contacts';
COMMENT ON COLUMN p_user_phone_contacts.contact_name IS 'Contact name from iPhone address book';
COMMENT ON COLUMN p_user_phone_contacts.contact_phone IS 'Contact phone number from iPhone address book';
COMMENT ON COLUMN p_user_phone_contacts.contact_email IS 'Contact email from iPhone address book';
COMMENT ON COLUMN p_user_phone_contacts.is_matched IS 'Whether this contact matches an existing app user';
COMMENT ON COLUMN p_user_phone_contacts.matched_user_id IS 'If matched, the app user ID this contact corresponds to';
