CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    name TEXT,
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

CREATE TABLE IF NOT EXISTS event (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     event TEXT NOT NULL,
     timestamp TEXT NOT NULL DEFAULT (datetime('now')),
     user TEXT NOT NULL,
     FOREIGN KEY (user) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS account_domain (
    account TEXT NOT NULL,
    domain TEXT NOT NULL,
    PRIMARY KEY (account, domain),
    FOREIGN KEY (account) REFERENCES account(id)
);

