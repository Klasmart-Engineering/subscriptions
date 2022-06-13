CREATE TABLE api_key (
    owner VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    PRIMARY KEY (owner),
    UNIQUE(api_key)
);

CREATE TABLE api_key_permission (
    owner VARCHAR(255) NOT NULL,
    permission VARCHAR(255) NOT NULL,
    PRIMARY KEY (owner, permission),
    FOREIGN KEY (owner) REFERENCES api_key(owner)
);
