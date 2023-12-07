create table kv_access
(
    kv_key     TEXT,
    access     TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX kv_key_idx ON kv_access (kv_key);
CREATE INDEX kv_key_access ON kv_access (access);

CREATE TRIGGER kv_access_updated_at
    BEFORE UPDATE
    ON kv_access
BEGIN
UPDATE kv_access
SET updated_at=CURRENT_TIMESTAMP
WHERE kv_key = NEW.kv_key;
END;

create table kv_access_changes
(
    id            integer
        constraint kv_access_changes_pk
            primary key autoincrement,
    kv_key        TEXT,
    operation     TEXT,
    before        TEXT,
    after         TEXT,
    originated_at TIMESTAMP default CURRENT_TIMESTAMP
);