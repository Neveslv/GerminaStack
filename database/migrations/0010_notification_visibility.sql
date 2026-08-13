ALTER TABLE notifications ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_notifications_user_visible_created
    ON notifications (id_user, created_at DESC, id DESC)
    WHERE is_hidden = FALSE;
