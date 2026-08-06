DO $$
BEGIN
    IF to_regclass('idx_posts_subject_created') IS NOT NULL
       AND pg_get_indexdef(to_regclass('idx_posts_subject_created')) NOT LIKE '%(id_subject, created_at DESC, id DESC)%' THEN
        DROP INDEX idx_posts_subject_created;
    END IF;

    IF to_regclass('idx_posts_user_created') IS NOT NULL
       AND pg_get_indexdef(to_regclass('idx_posts_user_created')) NOT LIKE '%(id_user, created_at DESC, id DESC)%' THEN
        DROP INDEX idx_posts_user_created;
    END IF;

    IF to_regclass('idx_comments_post_created') IS NOT NULL
       AND pg_get_indexdef(to_regclass('idx_comments_post_created')) NOT LIKE '%(id_post, created_at DESC, id DESC)%' THEN
        DROP INDEX idx_comments_post_created;
    END IF;

    IF to_regclass('idx_comments_on_comments_comment_created') IS NOT NULL
       AND pg_get_indexdef(to_regclass('idx_comments_on_comments_comment_created')) NOT LIKE '%(id_comment, created_at DESC, id DESC)%' THEN
        DROP INDEX idx_comments_on_comments_comment_created;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_posts_subject_created ON posts (id_subject, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_user_created ON posts (id_user, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments (id_post, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_on_comments_comment_created ON comments_on_comments (id_comment, created_at DESC, id DESC);
