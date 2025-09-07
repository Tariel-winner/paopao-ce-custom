-- Create reactions master table
CREATE TABLE p_reactions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(10),
    color VARCHAR(7), -- hex color
    is_positive BOOLEAN DEFAULT true,
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0,
    is_del SMALLINT NOT NULL DEFAULT 0
);

-- Create user-to-user reactions table (NEW - this is what you need)
CREATE TABLE p_user_reactions (
    id BIGSERIAL PRIMARY KEY,
    reactor_user_id BIGINT NOT NULL DEFAULT 0,  -- User who is reacting
    target_user_id BIGINT NOT NULL DEFAULT 0,   -- User being reacted to
    reaction_type_id BIGINT NOT NULL DEFAULT 1, -- Type of reaction
    created_on BIGINT NOT NULL DEFAULT 0,
    modified_on BIGINT NOT NULL DEFAULT 0,
    deleted_on BIGINT NOT NULL DEFAULT 0,
    is_del SMALLINT NOT NULL DEFAULT 0
);

-- Add unique constraint: one reaction per reactor-target pair
ALTER TABLE p_user_reactions ADD CONSTRAINT uk_user_reactions_unique 
    UNIQUE (reactor_user_id, target_user_id);

-- Create indexes for user reactions
CREATE INDEX idx_user_reactions_reactor ON p_user_reactions (reactor_user_id);
CREATE INDEX idx_user_reactions_target ON p_user_reactions (target_user_id);
CREATE INDEX idx_user_reactions_type ON p_user_reactions (reaction_type_id);

-- Create foreign key constraints for user reactions
ALTER TABLE p_user_reactions ADD CONSTRAINT fk_user_reactions_reaction 
    FOREIGN KEY (reaction_type_id) REFERENCES p_reactions(id) ON DELETE SET DEFAULT;

-- Add foreign key constraints to user table
ALTER TABLE p_user_reactions ADD CONSTRAINT fk_user_reactions_reactor_user 
    FOREIGN KEY (reactor_user_id) REFERENCES p_user(id) ON DELETE CASCADE;

ALTER TABLE p_user_reactions ADD CONSTRAINT fk_user_reactions_target_user 
    FOREIGN KEY (target_user_id) REFERENCES p_user(id) ON DELETE CASCADE;

-- Create index on reactions table
CREATE INDEX idx_reactions_name ON p_reactions(name);

-- Insert reaction types (user personality descriptors)
INSERT INTO p_reactions (id, name, description, icon, color, is_positive, created_on, modified_on) VALUES
-- Positive reactions (is_positive = true)
(1, 'like', 'Basic approval, neutral positive', '👍', '#4ECDC4', true, 1750794780, 1750794780),
(2, 'love', 'Strong emotional connection, affection', '❤️', '#FF6B6B', true, 1750794780, 1750794780),
(3, 'hot', 'Attractive, good-looking', '🔥', '#FF8C42', true, 1750794780, 1750794780),
(4, 'smart', 'Intelligent, clever', '🧠', '#6C5CE7', true, 1750794780, 1750794780),
(5, 'funny', 'Humorous, entertaining', '😂', '#FFEAA7', true, 1750794780, 1750794780),
(6, 'kind', 'Compassionate, helpful', '🤗', '#00B894', true, 1750794780, 1750794780),
(7, 'brave', 'Courageous, bold', '💪', '#F39C12', true, 1750794780, 1750794780),
(8, 'cool', 'Awesome, impressive', '😎', '#74B9FF', true, 1750794780, 1750794780),
(9, 'sweet', 'Nice, pleasant', '🍯', '#FFD93D', true, 1750794780, 1750794780),
(10, 'strong', 'Resilient, powerful', '💪', '#2D3436', true, 1750794780, 1750794780),
(11, 'friendly', 'Approachable, sociable', '😊', '#A29BFE', true, 1750794780, 1750794780),
(12, 'honest', 'Truthful, trustworthy', '🤝', '#00CEC9', true, 1750794780, 1750794780),
(13, 'generous', 'Giving, selfless', '🎁', '#FD79A8', true, 1750794780, 1750794780),
(14, 'fit', 'Athletic, in good shape', '🏃', '#00B894', true, 1750794780, 1750794780),
(15, 'creative', 'Artistic, innovative', '🎨', '#E84393', true, 1750794780, 1750794780),

-- Negative reactions (is_positive = false)
(16, 'stupid', 'Not smart, poor thinking', '🤦', '#E17055', false, 1750794780, 1750794780),
(17, 'mean', 'Unkind, cruel', '��', '#FF7675', false, 1750794780, 1750794780),
(18, 'fake', 'Dishonest, inauthentic', '🎭', '#636E72', false, 1750794780, 1750794780),
(19, 'lazy', 'Not hardworking', '😴', '#B2BEC3', false, 1750794780, 1750794780);

-- Add comments explaining the reaction system
COMMENT ON TABLE p_reactions IS 'Master table for user personality reaction types';
COMMENT ON TABLE p_user_reactions IS 'User-to-user reactions - one user reacts to another user''s personality';
COMMENT ON COLUMN p_user_reactions.reactor_user_id IS 'User who is giving the reaction';
COMMENT ON COLUMN p_user_reactions.target_user_id IS 'User who is receiving the reaction';
COMMENT ON COLUMN p_user_reactions.reaction_type_id IS 'Reference to p_reactions.id - describes the reactor''s perception of the target user'; 