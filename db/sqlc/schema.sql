CREATE TABLE user (
                      id   INTEGER PRIMARY KEY,
                      name text,
                      email text,
                      account integer NOT NULL,
                      externalId text,
                      createdAt text,
                      updatedAt text
);

CREATE TABLE account (
                         id    INTEGER PRIMARY KEY,
                         name text,
                         createdAt text,
                         updatedAt text
);


