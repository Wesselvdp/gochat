CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    name TEXT,
    email TEXT NOT NULL UNIQUE,
    account TEXT NOT NULL,
    externalId TEXT UNIQUE,
    createdAt TEXT NOT NULL DEFAULT (datetime('now')),
    updatedAt TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (account) REFERENCES account(id)
);

CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_user_account ON user(account);