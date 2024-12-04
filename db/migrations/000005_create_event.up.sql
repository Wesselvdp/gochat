CREATE TABLE IF NOT EXISTS event (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event TEXT NOT NULL,
    timestamp TEXT NOT NULL DEFAULT (datetime('now')),
    metadata JSON,
    user TEXT NOT NULL,
    FOREIGN KEY (user) REFERENCES user(id)
);

CREATE INDEX idx_event_user ON event(user);
