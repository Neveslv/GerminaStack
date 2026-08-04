CREATE TABLE IF NOT EXISTS two_factor_challenges (
    id text PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    max_attempts integer NOT NULL CHECK (max_attempts > 0),
    used_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_two_factor_challenges_user_active
    ON two_factor_challenges (user_id, created_at DESC)
    WHERE used_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_two_factor_challenges_expires_at
    ON two_factor_challenges (expires_at);
