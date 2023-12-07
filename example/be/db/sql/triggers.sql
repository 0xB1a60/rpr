-- kv
CREATE TRIGGER kv_insert
    AFTER INSERT
    ON kv
BEGIN
INSERT INTO kv_changes (operation, kv_key, after, originated_at)
VALUES ('INSERT', NEW.key,
        json_object('value', NEW.value, 'created_at', CAST(strftime('%s', NEW.created_at) AS INTEGER) * 1000,
                    'updated_at', CAST(strftime('%s', NEW.updated_at) AS INTEGER) * 1000), NEW.updated_at);
END;

CREATE TRIGGER kv_update
    AFTER UPDATE OF value
    ON kv
    FOR EACH ROW
BEGIN
INSERT INTO kv_changes (operation, kv_key, before, after, originated_at)
VALUES ('UPDATE', NEW.key,
        json_object('value', OLD.value, 'created_at', CAST(strftime('%s', OLD.created_at) AS INTEGER) * 1000,
                    'updated_at', CAST(strftime('%s', OLD.updated_at) AS INTEGER) * 1000),
        json_object('value', NEW.value, 'created_at', CAST(strftime('%s', NEW.created_at) AS INTEGER) * 1000,
                    'updated_at', CAST(strftime('%s', NEW.updated_at) AS INTEGER) * 1000), NEW.updated_at);

END;

CREATE TRIGGER kv_delete
    AFTER DELETE
    ON kv
    FOR EACH ROW
BEGIN
INSERT INTO kv_changes (operation, kv_key, before, originated_at)
VALUES ('DELETE', OLD.key,
        json_object('value', OLD.value, 'created_at', CAST(strftime('%s', OLD.created_at) AS INTEGER) * 1000,
                    'updated_at', CAST(strftime('%s', OLD.updated_at) AS INTEGER) * 1000), CURRENT_TIMESTAMP);

UPDATE kv_access
SET access = 'DELETED'
WHERE kv_key = OLD.key;
END;

-- kv access
CREATE TRIGGER kv_access_insert
    AFTER INSERT
    ON kv_access
BEGIN
INSERT INTO kv_access_changes (operation, kv_key, after, originated_at)
VALUES ('INSERT', NEW.kv_key,
        json_object('access', NEW.access, 'updated_at', CAST(strftime('%s', NEW.updated_at) AS INTEGER) * 1000),
        NEW.updated_at);
END;

CREATE TRIGGER kv_access_update
    AFTER UPDATE OF access
    ON kv_access
    FOR EACH ROW
BEGIN
INSERT INTO kv_access_changes (operation, kv_key, before, after, originated_at)
VALUES ('UPDATE', OLD.kv_key,
        json_object('access', OLD.access, 'updated_at', CAST(strftime('%s', OLD.updated_at) AS INTEGER) * 1000),
        json_object('access', NEW.access, 'updated_at', CAST(strftime('%s', NEW.updated_at) AS INTEGER) * 1000),
        NEW.updated_at);

END;

CREATE TRIGGER kv_access_delete
    AFTER DELETE
    ON kv_access
    FOR EACH ROW
BEGIN
SELECT RAISE(ABORT, 'Deletion is not allowed as it is needed for state propagation');
END;