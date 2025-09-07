-- Rollback: Restore the original categories from migration 0024
-- This restores the old categories if needed

-- Clear the updated categories
DELETE FROM p_categories;

-- Reset sequence
SELECT setval('p_categories_id_seq', 1, false);

-- Restore original categories from migration 0024
INSERT INTO p_categories (name, description, icon, color, created_on, modified_on) VALUES

-- ENTERTAINMENT (Core entertainment categories)
('Comedy', 'Funny and entertaining content', '😂', '#FFEAA7', 1750794780, 1750794780),
('Dance', 'Dance performances and tutorials', '💃', '#FF6B9D', 1750794780, 1750794780),
('Music', 'Music and audio content', '🎵', '#FF6B6B', 1750794780, 1750794780),
('Movies', 'Movie reviews and discussions', '🎬', '#6C5CE7', 1750794780, 1750794780),
('TV Shows', 'TV show content and discussions', '📺', '#A29BFE', 1750794780, 1750794780),
('Anime', 'Anime and manga content', '🌸', '#FD79A8', 1750794780, 1750794780),
('Gaming', 'Video games and gaming content', '🎮', '#4ECDC4', 1750794780, 1750794780),

-- LIFESTYLE & FASHION (Personal interests)
('Fashion', 'Fashion and style content', '👗', '#A29BFE', 1750794780, 1750794780),
('Beauty', 'Beauty tips and makeup tutorials', '💄', '#FF69B4', 1750794780, 1750794780),
('Lifestyle', 'Daily life and lifestyle content', '🌟', '#FFD93D', 1750794780, 1750794780),
('Food', 'Cooking and food content', '🍕', '#FF8C42', 1750794780, 1750794780),
('Travel', 'Travel and adventure content', '✈️', '#74B9FF', 1750794780, 1750794780),
('Home', 'Home improvement and decor', '🏠', '#FF7675', 1750794780, 1750794780),

-- SPORTS & FITNESS (Physical activities)
('Sports', 'Sports and fitness content', '⚽', '#96CEB4', 1750794780, 1750794780),
('Fitness', 'Workout and fitness routines', '💪', '#00B894', 1750794780, 1750794780),
('Outdoor', 'Outdoor activities and adventures', '🏔️', '#74B9FF', 1750794780, 1750794780),

-- TECHNOLOGY & SCIENCE (Intellectual interests)
('Technology', 'Tech news and discussions', '💻', '#45B7D1', 1750794780, 1750794780),
('Science', 'Scientific discoveries and explanations', '🔬', '#6C5CE7', 1750794780, 1750794780),
('Programming', 'Coding and software development', '💻', '#2D3436', 1750794780, 1750794780),

-- EDUCATION & LEARNING (Knowledge)
('Education', 'Learning and educational content', '📚', '#DDA0DD', 1750794780, 1750794780),
('Books', 'Book reviews and literature', '📖', '#DDA0DD', 1750794780, 1750794780),
('Language', 'Language learning content', '🗣️', '#6C5CE7', 1750794780, 1750794780),
('History', 'Historical content and discussions', '🏛️', '#E17055', 1750794780, 1750794780),

-- CREATIVE ARTS (Creative expression)
('Art', 'Art and creative content', '🎨', '#E84393', 1750794780, 1750794780),
('Photography', 'Photography and visual arts', '📸', '#6C5CE7', 1750794780, 1750794780),
('DIY', 'Do-it-yourself projects and crafts', '🔧', '#F39C12', 1750794780, 1750794780),
('Design', 'Graphic design and visual content', '🎨', '#E84393', 1750794780, 1750794780),

-- BUSINESS & PROFESSIONAL (Career interests)
('Business', 'Business and entrepreneurship', '💼', '#00B894', 1750794780, 1750794780),
('Finance', 'Financial advice and tips', '💰', '#00CEC9', 1750794780, 1750794780),
('Career', 'Career advice and professional development', '📈', '#2D3436', 1750794780, 1750794780),

-- HEALTH & WELLNESS (Wellbeing)
('Health', 'Health and wellness content', '🏥', '#E17055', 1750794780, 1750794780),
('Mental Health', 'Mental health and wellness', '🧠', '#6C5CE7', 1750794780, 1750794780),
('Yoga', 'Yoga and meditation content', '🧘', '#00B894', 1750794780, 1750794780),

-- FAMILY & RELATIONSHIPS (Social connections)
('Family', 'Family and parenting content', '👨‍👩‍👧‍👦', '#FF7675', 1750794780, 1750794780),
('Relationships', 'Relationship advice and content', '💕', '#FD79A8', 1750794780, 1750794780),
('Pets', 'Pet and animal content', '🐕', '#FDCB6E', 1750794780, 1750794780),

-- AUTOMOTIVE & TRANSPORTATION (Vehicles)
('Cars', 'Automotive content and reviews', '🚗', '#E17055', 1750794780, 1750794780),
('Motorcycles', 'Motorcycle content and reviews', '🏍️', '#F39C12', 1750794780, 1750794780),

-- NEWS & CURRENT EVENTS (Information)
('News', 'Current events and news', '📰', '#2D3436', 1750794780, 1750794780),
('Politics', 'Political discussions and content', '🗳️', '#636E72', 1750794780, 1750794780),
('Environment', 'Environmental awareness and sustainability', '🌍', '#00B894', 1750794780, 1750794780);
