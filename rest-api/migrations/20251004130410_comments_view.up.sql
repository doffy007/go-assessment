CREATE OR REPLACE VIEW comment_view AS
SELECT 
    a.id        AS comment_id,
    a.post_id   AS post_id,
    b.title     AS post_title,
    b.slug      AS post_slug,
    a.user_id,
    c.username  AS author_username,
    c.name      AS author_name,
    a.content,
    a.status,
    a.metadata,
    a.created_at,
    a.updated_at,
    a.deleted_at
FROM comments a
JOIN posts b ON a.post_id = b.id
JOIN users c ON a.user_id = c.id
WHERE a.deleted_at IS NULL;
