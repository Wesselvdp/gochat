CREATE TABLE IF NOT EXISTS account (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   name TEXT NOT NULL,
   createdAt TEXT NOT NULL DEFAULT (datetime('now')),
   updatedAt TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_account_name ON account(name);