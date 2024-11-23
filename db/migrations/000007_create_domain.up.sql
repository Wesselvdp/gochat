CREATE TABLE IF NOT EXISTS account_domain (
    account TEXT NOT NULL,
    domain TEXT NOT NULL,
    PRIMARY KEY (account, domain),
    FOREIGN KEY (account) REFERENCES account(id)
 );

