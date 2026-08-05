CREATE TABLE IF NOT EXISTS years (
    id BIGSERIAL PRIMARY KEY,
    year TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS subjects (
    id BIGSERIAL PRIMARY KEY,
    id_year BIGINT REFERENCES years(id) ON DELETE SET NULL,
    subject TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (id_year, subject)
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    id_year BIGINT NOT NULL REFERENCES years(id),
    name TEXT NOT NULL,
    profile_image_url TEXT,
    profile_image_description TEXT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    CHECK ((profile_image_url IS NULL) = (profile_image_description IS NULL))
);

CREATE TABLE IF NOT EXISTS preferences (
    id BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    contrast_theme TEXT,
    font_family TEXT,
    font_spacing TEXT,
    font_size TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    CHECK (contrast_theme IS NULL OR contrast_theme IN ('normal', 'dark', 'high_contrast', 'black_yellow', 'yellow_black')),
    CHECK (font_family IS NULL OR font_family IN ('normal', 'arial', 'verdana', 'lexend', 'atkinson_hyperlegible', 'open_dyslexic')),
    CHECK (font_spacing IS NULL OR font_spacing IN ('normal', 'pequeno', 'grande')),
    CHECK (font_size IS NULL OR font_size IN ('normal', 'pequeno', 'grande'))
);

CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id_subject BIGINT NOT NULL REFERENCES subjects(id),
    title TEXT NOT NULL,
    image_url TEXT,
    image_description TEXT,
    content TEXT NOT NULL,
    likes BIGINT NOT NULL DEFAULT 0 CHECK (likes >= 0),
    dislikes BIGINT NOT NULL DEFAULT 0 CHECK (dislikes >= 0),
    created_at TIMESTAMPTZ DEFAULT now(),
    CHECK ((image_url IS NULL) = (image_description IS NULL))
);

CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    id_post BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    id_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    likes BIGINT NOT NULL DEFAULT 0 CHECK (likes >= 0),
    dislikes BIGINT NOT NULL DEFAULT 0 CHECK (dislikes >= 0),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS comments_on_comments (
    id BIGSERIAL PRIMARY KEY,
    id_comment BIGINT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    id_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    likes BIGINT NOT NULL DEFAULT 0 CHECK (likes >= 0),
    dislikes BIGINT NOT NULL DEFAULT 0 CHECK (dislikes >= 0),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reactions (
    id BIGSERIAL PRIMARY KEY,
    id_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id_post BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    id_comment BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    id_comment_on_comment BIGINT REFERENCES comments_on_comments(id) ON DELETE CASCADE,
    reaction_type TEXT NOT NULL CHECK (reaction_type IN ('like', 'dislike')),
    created_at TIMESTAMPTZ DEFAULT now(),
    CHECK (((id_post IS NOT NULL)::integer + (id_comment IS NOT NULL)::integer + (id_comment_on_comment IS NOT NULL)::integer) = 1)
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    id_post BIGINT REFERENCES posts(id) ON DELETE SET NULL,
    id_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text_show TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subjects_year ON subjects (id_year);
CREATE INDEX IF NOT EXISTS idx_users_year ON users (id_year);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_subject_created ON posts (id_subject, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_user_created ON posts (id_user, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments (id_post, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments (id_user);
CREATE INDEX IF NOT EXISTS idx_comments_on_comments_comment_created ON comments_on_comments (id_comment, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_on_comments_user ON comments_on_comments (id_user);
CREATE INDEX IF NOT EXISTS idx_reactions_post ON reactions (id_post) WHERE id_post IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reactions_comment ON reactions (id_comment) WHERE id_comment IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reactions_comment_on_comment ON reactions (id_comment_on_comment) WHERE id_comment_on_comment IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications (id_user, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread_created ON notifications (id_user, created_at DESC, id DESC) WHERE is_read = FALSE;
