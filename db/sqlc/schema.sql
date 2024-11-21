CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    account TEXT PRIMARY KEY,
    externalId TEXT UNIQUE,
    createdAt TEXT NOT NULL DEFAULT (datetime('now')),
    updatedAt TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (account) REFERENCES account(id)
);
CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_user_account ON user(account);

CREATE TABLE IF NOT EXISTS account (
   id TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   createdAt TEXT NOT NULL DEFAULT (datetime('now')),
   updatedAt TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_account_name ON account(name);