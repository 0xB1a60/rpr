create table kv
(
    key        TEXT
        constraint kv_pk primary key,
    value      TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER kv_updated_at
    BEFORE UPDATE
    ON kv
BEGIN
UPDATE kv
SET updated_at=CURRENT_TIMESTAMP
WHERE key = NEW.key;
END;

create table kv_changes
(
    id            integer
        constraint kv_changes_pk
            primary key autoincrement,
    kv_key        TEXT,
    operation     TEXT,
    before        TEXT,
    after         TEXT,
    originated_at TIMESTAMP default CURRENT_TIMESTAMP
);

