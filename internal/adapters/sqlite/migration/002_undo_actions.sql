CREATE TABLE IF NOT EXISTS undo_actions
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id     TEXT    NOT NULL,
    item_id     TEXT    NOT NULL,
    was_checked INTEGER NOT NULL,
    checked_at  TEXT,
    position    INTEGER NOT NULL
);