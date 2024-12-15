CREATE TABLE IF NOT EXISTS file (
   id TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   createdAt TEXT NOT NULL DEFAULT (datetime('now')),
   updatedAt TEXT NOT NULL DEFAULT (datetime('now')),
   owner TEXT NOT NULL,
   FOREIGN KEY (owner) REFERENCES user(id)
);
