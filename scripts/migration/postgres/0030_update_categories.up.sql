-- Update categories to match the new static frontend categories (35 total)
-- This migration safely updates existing categories and adds new ones
-- without breaking foreign key references

-- First, clear existing categories (this is safe because we'll repopulate them)
DELETE FROM p_categories;

-- Reset the sequence to start from 1
SELECT setval('p_categories_id_seq', 1, false);

-- Insert the new 35 categories matching the frontend static list
INSERT INTO p_categories (id, name, description, icon, color, created_on, modified_on) VALUES

-- ENTERTAINMENT & VIRAL (High trending potential)
(1, 'Comedy', 'Funny and entertaining content', '😂', '#FFEAA7', 1750794780, 1750794780),
(2, 'Music', 'Music and audio content', '🎵', '#FF6B6B', 1750794780, 1750794780),
(3, 'Movies', 'Movie reviews and discussions', '🎬', '#6C5CE7', 1750794780, 1750794780),
(4, 'TV Shows', 'TV show content and discussions', '📺', '#A29BFE', 1750794780, 1750794780),
(5, 'Gaming', 'Video games and gaming content', '🎮', '#4ECDC4', 1750794780, 1750794780),
(6, 'Viral', 'Viral content and memes', '🔥', '#FF4757', 1750794780, 1750794780),
(7, 'Celebrities', 'Celebrity news and gossip', '⭐', '#FFA502', 1750794780, 1750794780),

-- NEWS & CURRENT EVENTS (Always trending)
(8, 'News', 'Current events and breaking news', '📰', '#2D3436', 1750794780, 1750794780),
(9, 'Politics', 'Political discussions and content', '🗳️', '#636E72', 1750794780, 1750794780),
(10, 'Weather', 'Weather events and natural disasters', '🌦️', '#74B9FF', 1750794780, 1750794780),

-- TECHNOLOGY & INNOVATION (High trending)
(11, 'Technology', 'Tech news and product launches', '💻', '#45B7D1', 1750794780, 1750794780),
(12, 'AI', 'Artificial intelligence and automation', '🤖', '#6C5CE7', 1750794780, 1750794780),
(13, 'Social Media', 'Platform updates and influencer content', '📱', '#FF6B9D', 1750794780, 1750794780),

-- FINANCE & CRYPTO (Market trending)
(14, 'Finance', 'Market news and financial advice', '💰', '#00CEC9', 1750794780, 1750794780),
(15, 'Crypto', 'Cryptocurrency and blockchain', '₿', '#F39C12', 1750794780, 1750794780),

-- SPORTS & EVENTS (Regular trending)
(16, 'Sports', 'Sports events and athlete news', '⚽', '#96CEB4', 1750794780, 1750794780),
(17, 'Esports', 'Competitive gaming and tournaments', '🏆', '#4ECDC4', 1750794780, 1750794780),

-- LIFESTYLE & CULTURE (Trending topics)
(18, 'Fashion', 'Fashion trends and style', '👗', '#A29BFE', 1750794780, 1750794780),
(19, 'Beauty', 'Beauty trends and tutorials', '💄', '#FF69B4', 1750794780, 1750794780),
(20, 'Food', 'Food trends and viral recipes', '🍕', '#FF8C42', 1750794780, 1750794780),
(21, 'Travel', 'Travel destinations and experiences', '✈️', '#74B9FF', 1750794780, 1750794780),

-- CREATIVE & ARTS (Trending content)
(22, 'Creative Arts', 'Art, photography, and design', '🎨', '#E84393', 1750794780, 1750794780),
(23, 'Dance', 'Dance trends and performances', '💃', '#FF6B9D', 1750794780, 1750794780),
(24, 'Music Production', 'Music creation and production', '🎧', '#FF6B6B', 1750794780, 1750794780),

-- HEALTH & FITNESS (Popular topics)
(25, 'Health & Wellness', 'Health trends and wellness tips', '🏥', '#E17055', 1750794780, 1750794780),
(26, 'Fitness', 'Workout trends and fitness content', '💪', '#00B894', 1750794780, 1750794780),

-- BUSINESS & CAREER (Professional trending)
(27, 'Business', 'Business news and entrepreneurship', '💼', '#00B894', 1750794780, 1750794780),
(28, 'Career', 'Career advice and job market trends', '📈', '#2D3436', 1750794780, 1750794780),

-- SCIENCE & EDUCATION (Knowledge trending)
(29, 'Science', 'Scientific discoveries and breakthroughs', '🔬', '#6C5CE7', 1750794780, 1750794780),
(30, 'Learning', 'Educational content and tutorials', '📚', '#DDA0DD', 1750794780, 1750794780),

-- LIFESTYLE & PERSONAL (Relatable content)
(31, 'Relationships', 'Dating and relationship advice', '💕', '#FD79A8', 1750794780, 1750794780),
(32, 'Family', 'Parenting and family content', '👨‍👩‍👧‍👦', '#FF7675', 1750794780, 1750794780),
(33, 'Pets', 'Pet videos and animal content', '🐕', '#FDCB6E', 1750794780, 1750794780),

-- AUTOMOTIVE & TRANSPORTATION (Enthusiast content)
(34, 'Automotive', 'Car reviews and automotive news', '🚗', '#E17055', 1750794780, 1750794780),

-- ENVIRONMENT & SUSTAINABILITY (Growing trend)
(35, 'Environment', 'Climate change and sustainability', '🌍', '#00B894', 1750794780, 1750794780);

-- Set the sequence to continue from 36 for future auto-generated IDs
SELECT setval('p_categories_id_seq', 35, true);
