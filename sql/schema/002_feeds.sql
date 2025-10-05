-- +goose Up
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users (id)
);

-- +goose Down
DROP TABLE feeds;