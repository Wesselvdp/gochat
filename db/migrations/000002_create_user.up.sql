CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    account INTEGER NOT NULL,
    externalId TEXT UNIQUE,
    createdAt TEXT NOT NULL DEFAULT (datetime('now')),
    updatedAt TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (account) REFERENCES account(id)
);

CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_user_account ON user(account);