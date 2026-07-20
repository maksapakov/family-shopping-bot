CREATE TABLE IF NOT EXISTS shopping_lists
(
    chat_id     TEXT PRIMARY KEY,
    list_id     TEXT    NOT NULL,
    active_tab  INTEGER NOT NULL,
    message_ref TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS items
(
    id         TEXT PRIMARY KEY,
    chat_id    TEXT    NOT NULL REFERENCES shopping_lists (chat_id) ON DELETE CASCADE,
    name       TEXT    NOT NULL,
    is_checked INTEGER NOT NULL DEFAULT 0,
    checked_at TEXT,
    added_by   TEXT    NOT NULL,
    location   INTEGER NOT NULL
);