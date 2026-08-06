CREATE OR REPLACE FUNCTION notify_mentions()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    post_id BIGINT;
    post_title TEXT;
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'posts' THEN post_id := NEW.id;
        WHEN 'comments' THEN post_id := NEW.id_post;
        WHEN 'comments_on_comments' THEN
            SELECT id_post INTO post_id FROM comments WHERE id = NEW.id_comment;
    END CASE;

    SELECT title INTO post_title FROM posts WHERE id = post_id;
    IF post_id IS NULL OR post_title IS NULL THEN
        RETURN NEW;
    END IF;

    INSERT INTO notifications (id_post, id_user, text_show)
    SELECT post_id, mentioned.id,
           author.username || ' mencionou você no post "' || post_title || '"'
    FROM (
        SELECT DISTINCT (regexp_matches(NEW.content, '@([A-Za-z0-9_]+([.][A-Za-z0-9_]+)*)', 'g'))[1] AS username
    ) AS mentions
    JOIN users AS mentioned ON mentioned.username = mentions.username
    JOIN users AS author ON author.id = NEW.id_user
    WHERE mentioned.id <> NEW.id_user;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_notify_mentions_post ON posts;
CREATE TRIGGER trg_notify_mentions_post
AFTER INSERT ON posts
FOR EACH ROW EXECUTE FUNCTION notify_mentions();

DROP TRIGGER IF EXISTS trg_notify_mentions_comment ON comments;
CREATE TRIGGER trg_notify_mentions_comment
AFTER INSERT ON comments
FOR EACH ROW EXECUTE FUNCTION notify_mentions();

DROP TRIGGER IF EXISTS trg_notify_mentions_comment_on_comment ON comments_on_comments;
CREATE TRIGGER trg_notify_mentions_comment_on_comment
AFTER INSERT ON comments_on_comments
FOR EACH ROW EXECUTE FUNCTION notify_mentions();
