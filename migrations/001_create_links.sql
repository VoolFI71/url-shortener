CREATE TABLE links (
    code         varchar(10) PRIMARY KEY,
    original_url text        NOT NULL UNIQUE,

    CONSTRAINT links_code_format CHECK (code ~ '^[a-zA-Z0-9_]{10}$')
);
