package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ChangesEntry struct {
	Id           string
	Operation    string
	Before       *string
	After        *string
	OriginatedAt time.Time
}

func (d *Database) ReadKVChange(rowId int64) (*ChangesEntry, error) {
	row := d.base.QueryRow(`select kc.kv_key, kc.operation, kc.before, kc.after, kc.originated_at from kv_changes kc
   inner join kv_access ka on kc.kv_key = ka.kv_key where kc.rowid == $1`, rowId)

	var c ChangesEntry
	if err := row.Scan(&c.Id, &c.Operation, &c.Before, &c.After, &c.OriginatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, d.RemoveChange(KVChangesTable, rowId)
		}
		return nil, err
	}
	return &c, nil
}

func (d *Database) ReadKVAccessChange(rowId int64) (*ChangesEntry, error) {
	row := d.base.QueryRow(`select kc.kv_key, kc.operation, kc.before, kc.after, kc.originated_at from kv_access_changes kc  
                                                                            where kc.rowid == $1`, rowId)

	var c ChangesEntry
	if err := row.Scan(&c.Id, &c.Operation, &c.Before, &c.After, &c.OriginatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, d.RemoveChange(KVAccessChangesTable, rowId)
		}
		return nil, err
	}
	return &c, nil
}

func (d *Database) RemoveChange(table string, id int64) error {
	if table == KVChangesTable {
		_, err := d.base.Exec(`DELETE FROM kv_changes WHERE rowid=$1`, id)
		return err
	}
	if table == KVAccessChangesTable {
		_, err := d.base.Exec(`DELETE FROM kv_access_changes WHERE rowid=$1`, id)
		return err
	}
	return fmt.Errorf("table: %s is not supported", table)
}
