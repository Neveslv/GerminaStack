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
DECLARE
    v_id BIGINT;
BEGIN
    IF p_message_type = 'post' THEN
        IF p_id_parent IS NULL THEN
            RAISE EXCEPTION 'post requires an id_subject parent';
        END IF;
        INSERT INTO posts (id_user, id_subject, title, image_url, image_description, content)
        VALUES (p_id_user, p_id_parent, p_title, p_image_url, p_image_description, p_content)
        RETURNING id INTO v_id;
    ELSIF p_message_type = 'comment' THEN
        IF p_id_parent IS NULL THEN
            RAISE EXCEPTION 'comment requires an id_post parent';
        END IF;
        INSERT INTO comments (id_post, id_user, content)
        VALUES (p_id_parent, p_id_user, p_content)
        RETURNING id INTO v_id;
    ELSIF p_message_type = 'comment_on_comment' THEN
        IF p_id_parent IS NULL THEN
            RAISE EXCEPTION 'comment_on_comment requires an id_comment parent';
        END IF;
        INSERT INTO comments_on_comments (id_comment, id_user, content)
        VALUES (p_id_parent, p_id_user, p_content)
        RETURNING id INTO v_id;
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
DECLARE
    v_reaction TEXT;
BEGIN
    IF p_reaction_type NOT IN ('like', 'dislike') THEN
        RAISE EXCEPTION 'invalid reaction type: %', p_reaction_type;
    END IF;

    IF p_message_type = 'post' THEN
        SELECT reaction_type INTO v_reaction
        FROM reactions
        WHERE id_user = p_id_user AND id_post = p_id_message;
        IF v_reaction IS NULL THEN
            INSERT INTO reactions (id_user, id_post, reaction_type)
            VALUES (p_id_user, p_id_message, p_reaction_type);
        ELSIF v_reaction = p_reaction_type THEN
            DELETE FROM reactions
            WHERE id_user = p_id_user AND id_post = p_id_message;
        ELSE
            UPDATE reactions
            SET reaction_type = p_reaction_type
            WHERE id_user = p_id_user AND id_post = p_id_message;
        END IF;
    ELSIF p_message_type = 'comment' THEN
        SELECT reaction_type INTO v_reaction
        FROM reactions
        WHERE id_user = p_id_user AND id_comment = p_id_message;
        IF v_reaction IS NULL THEN
            INSERT INTO reactions (id_user, id_comment, reaction_type)
            VALUES (p_id_user, p_id_message, p_reaction_type);
        ELSIF v_reaction = p_reaction_type THEN
            DELETE FROM reactions
            WHERE id_user = p_id_user AND id_comment = p_id_message;
        ELSE
            UPDATE reactions
            SET reaction_type = p_reaction_type
            WHERE id_user = p_id_user AND id_comment = p_id_message;
        END IF;
    ELSIF p_message_type = 'comment_on_comment' THEN
        SELECT reaction_type INTO v_reaction
        FROM reactions
        WHERE id_user = p_id_user AND id_comment_on_comment = p_id_message;
        IF v_reaction IS NULL THEN
            INSERT INTO reactions (id_user, id_comment_on_comment, reaction_type)
            VALUES (p_id_user, p_id_message, p_reaction_type);
        ELSIF v_reaction = p_reaction_type THEN
            DELETE FROM reactions
            WHERE id_user = p_id_user AND id_comment_on_comment = p_id_message;
        ELSE
            UPDATE reactions
            SET reaction_type = p_reaction_type
            WHERE id_user = p_id_user AND id_comment_on_comment = p_id_message;
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
    UPDATE notifications
    SET is_read = TRUE
    WHERE id_user = p_id_user AND is_read = FALSE;
END;
$$;

CREATE OR REPLACE FUNCTION update_reaction_count()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_post_id BIGINT;
    v_comment_id BIGINT;
    v_reply_id BIGINT;
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
        UPDATE posts
        SET likes = (SELECT COUNT(*) FROM reactions WHERE id_post = v_post_id AND reaction_type = 'like'),
            dislikes = (SELECT COUNT(*) FROM reactions WHERE id_post = v_post_id AND reaction_type = 'dislike')
        WHERE id = v_post_id;
    ELSIF v_comment_id IS NOT NULL THEN
        UPDATE comments
        SET likes = (SELECT COUNT(*) FROM reactions WHERE id_comment = v_comment_id AND reaction_type = 'like'),
            dislikes = (SELECT COUNT(*) FROM reactions WHERE id_comment = v_comment_id AND reaction_type = 'dislike')
        WHERE id = v_comment_id;
    ELSIF v_reply_id IS NOT NULL THEN
        UPDATE comments_on_comments
        SET likes = (SELECT COUNT(*) FROM reactions WHERE id_comment_on_comment = v_reply_id AND reaction_type = 'like'),
            dislikes = (SELECT COUNT(*) FROM reactions WHERE id_comment_on_comment = v_reply_id AND reaction_type = 'dislike')
        WHERE id = v_reply_id;
    END IF;
    RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS trg_update_reaction_count ON reactions;
CREATE TRIGGER trg_update_reaction_count
AFTER INSERT OR UPDATE OR DELETE ON reactions
FOR EACH ROW EXECUTE FUNCTION update_reaction_count();
