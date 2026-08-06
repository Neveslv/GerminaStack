CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_user_post_unique
    ON reactions (id_user, id_post) WHERE id_post IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_user_comment_unique
    ON reactions (id_user, id_comment) WHERE id_comment IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_user_comment_on_comment_unique
    ON reactions (id_user, id_comment_on_comment) WHERE id_comment_on_comment IS NOT NULL;

CREATE OR REPLACE FUNCTION create_message(
    p_message_type TEXT,
    p_id_user BIGINT,
    p_id_parent BIGINT,
    p_content TEXT,
    p_title TEXT DEFAULT NULL,
    p_image_url TEXT DEFAULT NULL,
    p_image_description TEXT DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE v_id BIGINT;
BEGIN
    IF p_message_type = 'post' THEN
        IF p_id_parent IS NULL THEN RAISE EXCEPTION 'post requires an id_subject parent'; END IF;
        INSERT INTO posts (id_user, id_subject, title, image_url, image_description, content)
        VALUES (p_id_user, p_id_parent, p_title, p_image_url, p_image_description, p_content)
        RETURNING id INTO v_id;
    ELSIF p_message_type = 'comment' THEN
        IF p_id_parent IS NULL THEN RAISE EXCEPTION 'comment requires an id_post parent'; END IF;
        INSERT INTO comments (id_post, id_user, content)
        VALUES (p_id_parent, p_id_user, p_content) RETURNING id INTO v_id;
    ELSIF p_message_type = 'comment_on_comment' THEN
        IF p_id_parent IS NULL THEN RAISE EXCEPTION 'comment_on_comment requires an id_comment parent'; END IF;
        INSERT INTO comments_on_comments (id_comment, id_user, content)
        VALUES (p_id_parent, p_id_user, p_content) RETURNING id INTO v_id;
    ELSE
        RAISE EXCEPTION 'invalid message type: %', p_message_type;
    END IF;
    RETURN v_id;
END;
$$;

CREATE OR REPLACE FUNCTION reaction(
    p_id_user BIGINT,
    p_id_message BIGINT,
    p_message_type TEXT,
    p_reaction_type TEXT
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE v_reaction TEXT;
BEGIN
    IF p_reaction_type NOT IN ('like', 'dislike') THEN
        RAISE EXCEPTION 'invalid reaction type: %', p_reaction_type;
    END IF;
    IF p_message_type = 'post' THEN
        SELECT reaction_type INTO v_reaction FROM reactions WHERE id_user = p_id_user AND id_post = p_id_message;
        IF v_reaction IS NULL THEN
            INSERT INTO reactions (id_user, id_post, reaction_type) VALUES (p_id_user, p_id_message, p_reaction_type);
        ELSIF v_reaction = p_reaction_type THEN
            DELETE FROM reactions WHERE id_user = p_id_user AND id_post = p_id_message;
        ELSE
            UPDATE reactions SET reaction_type = p_reaction_type WHERE id_user = p_id_user AND id_post = p_id_message;
        END IF;
    ELSIF p_message_type = 'comment' THEN
        SELECT reaction_type INTO v_reaction FROM reactions WHERE id_user = p_id_user AND id_comment = p_id_message;
        IF v_reaction IS NULL THEN
            INSERT INTO reactions (id_user, id_comment, reaction_type) VALUES (p_id_user, p_id_message, p_reaction_type);
        ELSIF v_reaction = p_reaction_type THEN
            DELETE FROM reactions WHERE id_user = p_id_user AND id_comment = p_id_message;
        ELSE
            UPDATE reactions SET reaction_type = p_reaction_type WHERE id_user = p_id_user AND id_comment = p_id_message;
        END IF;
    ELSIF p_message_type = 'comment_on_comment' THEN
        SELECT reaction_type INTO v_reaction FROM reactions WHERE id_user = p_id_user AND id_comment_on_comment = p_id_message;
        IF v_reaction IS NULL THEN
            INSERT INTO reactions (id_user, id_comment_on_comment, reaction_type) VALUES (p_id_user, p_id_message, p_reaction_type);
        ELSIF v_reaction = p_reaction_type THEN
            DELETE FROM reactions WHERE id_user = p_id_user AND id_comment_on_comment = p_id_message;
        ELSE
            UPDATE reactions SET reaction_type = p_reaction_type WHERE id_user = p_id_user AND id_comment_on_comment = p_id_message;
        END IF;
    ELSE
        RAISE EXCEPTION 'invalid message type: %', p_message_type;
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION mark_notifications_as_read(p_id_user BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE notifications SET is_read = TRUE WHERE id_user = p_id_user AND is_read = FALSE;
END;
$$;

CREATE OR REPLACE FUNCTION update_reaction_count()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE v_post_id BIGINT; v_comment_id BIGINT; v_reply_id BIGINT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        v_post_id := OLD.id_post;
        v_comment_id := OLD.id_comment;
        v_reply_id := OLD.id_comment_on_comment;
    ELSE
        v_post_id := NEW.id_post;
        v_comment_id := NEW.id_comment;
        v_reply_id := NEW.id_comment_on_comment;
    END IF;
    IF v_post_id IS NOT NULL THEN
        UPDATE posts SET likes = GREATEST(0, (SELECT COUNT(*) FROM reactions WHERE id_post = v_post_id AND reaction_type = 'like')),
                         dislikes = GREATEST(0, (SELECT COUNT(*) FROM reactions WHERE id_post = v_post_id AND reaction_type = 'dislike'))
        WHERE id = v_post_id;
    ELSIF v_comment_id IS NOT NULL THEN
        UPDATE comments SET likes = GREATEST(0, (SELECT COUNT(*) FROM reactions WHERE id_comment = v_comment_id AND reaction_type = 'like')),
                            dislikes = GREATEST(0, (SELECT COUNT(*) FROM reactions WHERE id_comment = v_comment_id AND reaction_type = 'dislike'))
        WHERE id = v_comment_id;
    ELSIF v_reply_id IS NOT NULL THEN
        UPDATE comments_on_comments SET likes = GREATEST(0, (SELECT COUNT(*) FROM reactions WHERE id_comment_on_comment = v_reply_id AND reaction_type = 'like')),
                                        dislikes = GREATEST(0, (SELECT COUNT(*) FROM reactions WHERE id_comment_on_comment = v_reply_id AND reaction_type = 'dislike'))
        WHERE id = v_reply_id;
    END IF;
    RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS trg_update_reaction_count ON reactions;
CREATE TRIGGER trg_update_reaction_count
AFTER INSERT OR UPDATE OR DELETE ON reactions
FOR EACH ROW EXECUTE FUNCTION update_reaction_count();

CREATE OR REPLACE FUNCTION notify_mentions()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE v_username TEXT; v_id_user_mentioned BIGINT; v_post_id BIGINT; v_post_title TEXT; v_author_username TEXT;
BEGIN
    IF TG_TABLE_NAME = 'posts' THEN
        v_post_id := NEW.id; v_post_title := NEW.title;
    ELSIF TG_TABLE_NAME = 'comments' THEN
        v_post_id := NEW.id_post;
        SELECT title INTO v_post_title FROM posts WHERE id = v_post_id;
    ELSIF TG_TABLE_NAME = 'comments_on_comments' THEN
        SELECT posts.id, posts.title INTO v_post_id, v_post_title
        FROM comments JOIN posts ON posts.id = comments.id_post WHERE comments.id = NEW.id_comment;
    ELSE
        RETURN NEW;
    END IF;
    SELECT username INTO v_author_username FROM users WHERE id = NEW.id_user;
    FOR v_username IN SELECT DISTINCT matches.parts[1]
        FROM regexp_matches(NEW.content, '@([a-zA-Z0-9_]+)', 'g') AS matches(parts)
    LOOP
        SELECT id INTO v_id_user_mentioned FROM users WHERE username = v_username;
        IF v_id_user_mentioned IS NOT NULL AND v_id_user_mentioned <> NEW.id_user THEN
            INSERT INTO notifications (id_post, id_user, text_show)
            VALUES (v_post_id, v_id_user_mentioned, v_author_username || ' mentioned you in post "' || v_post_title || '"');
        END IF;
    END LOOP;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_notify_mentions_post ON posts;
CREATE TRIGGER trg_notify_mentions_post AFTER INSERT ON posts
FOR EACH ROW EXECUTE FUNCTION notify_mentions();
DROP TRIGGER IF EXISTS trg_notify_mentions_comment ON comments;
CREATE TRIGGER trg_notify_mentions_comment AFTER INSERT ON comments
FOR EACH ROW EXECUTE FUNCTION notify_mentions();
DROP TRIGGER IF EXISTS trg_notify_mentions_comment_on_comment ON comments_on_comments;
CREATE TRIGGER trg_notify_mentions_comment_on_comment AFTER INSERT ON comments_on_comments
FOR EACH ROW EXECUTE FUNCTION notify_mentions();
