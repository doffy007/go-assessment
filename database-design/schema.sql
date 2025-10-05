

-- 1. Users
CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    bio TEXT,
    profile_image_url TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. User Relationships
CREATE TABLE user_relationships (
    follower_id BIGINT NOT NULL REFERENCES users(user_id),
    following_id BIGINT NOT NULL REFERENCES users(user_id),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);

-- 3. Posts
CREATE TABLE posts (
    post_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    content TEXT,
    media_url TEXT[],
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE
);

-- 4. Comments
CREATE TABLE comments (
    comment_id BIGINT PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES posts(post_id),
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    content TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE
);

-- 5. Reactions
CREATE TABLE reactions (
    reaction_id BIGINT PRIMARY KEY,
    post_id BIGINT REFERENCES posts(post_id),
    comment_id BIGINT REFERENCES comments(comment_id),
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    reaction_type VARCHAR(20) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(post_id, comment_id, user_id, reaction_type)
);

-- 6. Messages
CREATE TABLE messages (
    message_id BIGINT PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES users(user_id),
    receiver_id BIGINT NOT NULL REFERENCES users(user_id),
    content TEXT NOT NULL,
    metadata JSONB,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 7. Activity Feeds
CREATE TABLE feeds (
    feed_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    actor_id BIGINT NOT NULL REFERENCES users(user_id),
    action_type VARCHAR(50) NOT NULL,
    target_post_id BIGINT REFERENCES posts(post_id),
    target_comment_id BIGINT REFERENCES comments(comment_id),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE INDEX idx_posts_user_created ON posts(user_id, created_at);
CREATE INDEX idx_comments_post_created ON comments(post_id, created_at);
CREATE INDEX idx_messages_receiver_created ON messages(receiver_id, created_at);
CREATE INDEX idx_feeds_user_created ON feeds(user_id, created_at);
CREATE INDEX idx_relationship_follower ON user_relationships(follower_id);
CREATE INDEX idx_relationship_following ON user_relationships(following_id);



-- 1. User Feed View
CREATE VIEW user_feed_view AS
SELECT
    a.post_id,
    a.user_id AS author_id,
    b.username AS author_username,
    b.display_name AS author_name,
    a.content,
    a.media_url,
    a.metadata AS post_metadata,
    COUNT(DISTINCT c.comment_id) AS comment_count,
    COUNT(DISTINCT d.reaction_id) AS reaction_count,
    a.created_at
FROM posts a
JOIN users b ON b.user_id = a.user_id
LEFT JOIN comments c ON c.post_id = a.post_id AND c.is_deleted = FALSE
LEFT JOIN reactions d ON d.post_id = a.post_id
WHERE a.is_deleted = FALSE
GROUP BY a.post_id, a.user_id, b.username, b.display_name, a.content, a.media_url, a.metadata, a.created_at;

-- 2. Post Details View
CREATE VIEW post_detail_view AS
SELECT
    a.post_id,
    a.user_id AS author_id,
    a.content AS post_content,
    a.media_url,
    c.comment_id,
    c.user_id AS commenter_id,
    c.content AS comment_content,
    d.reaction_id,
    d.user_id AS reactor_id,
    d.reaction_type,
    a.metadata AS post_metadata,
    c.metadata AS comment_metadata,
    d.metadata AS reaction_metadata,
    a.created_at AS post_created_at,
    c.created_at AS comment_created_at,
    d.created_at AS reaction_created_at
FROM posts a
LEFT JOIN comments c ON c.post_id = a.post_id AND c.is_deleted = FALSE
LEFT JOIN reactions d ON d.post_id = a.post_id
WHERE a.is_deleted = FALSE;

-- 3. Messages Inbox View
CREATE VIEW inbox_view AS
SELECT
    a.receiver_id,
    a.sender_id,
    b.username AS sender_username,
    MAX(a.created_at) AS latest_message_time,
    COUNT(*) FILTER (WHERE a.is_read = FALSE) AS unread_count
FROM messages a
JOIN users b ON b.user_id = a.sender_id
GROUP BY a.receiver_id, a.sender_id, b.username;



-- 1. Materialized Feed View
CREATE MATERIALIZED VIEW user_feed_mv AS
SELECT
    e.follower_id AS user_id,
    a.post_id,
    a.user_id AS author_id,
    b.username AS author_username,
    a.content,
    a.media_url,
    COUNT(DISTINCT c.comment_id) AS comment_count,
    COUNT(DISTINCT d.reaction_id) AS reaction_count,
    a.created_at
FROM posts a
JOIN users b ON b.user_id = a.user_id
JOIN user_relationships e ON e.following_id = a.user_id
LEFT JOIN comments c ON c.post_id = a.post_id AND c.is_deleted = FALSE
LEFT JOIN reactions d ON d.post_id = a.post_id
WHERE a.is_deleted = FALSE
GROUP BY e.follower_id, a.post_id, a.user_id, b.username, a.content, a.media_url, a.created_at;

-- 2. Materialized Reaction Count per Post
CREATE MATERIALIZED VIEW post_reaction_count_mv AS
SELECT
    a.post_id,
    COUNT(*) AS total_reactions,
    COUNT(*) FILTER (WHERE a.reaction_type = 'like') AS like_count,
    COUNT(*) FILTER (WHERE a.reaction_type = 'love') AS love_count,
    COUNT(*) FILTER (WHERE a.reaction_type = 'haha') AS haha_count
FROM reactions a
GROUP BY a.post_id;
