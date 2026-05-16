CREATE TABLE users_comics_saved (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comics_id INT NOT NULL REFERENCES comics(id) ON DELETE CASCADE,

    saved_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY(user_id, comics_id)
);