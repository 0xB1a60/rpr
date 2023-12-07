package db

import (
	"context"
	"database/sql"
	_ "embed"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/mattn/go-sqlite3"
)

//go:embed sql/kv.sql
var kvSql string

//go:embed sql/kv_access.sql
var kvAccessSql string

//go:embed sql/triggers.sql
var triggersSql string

type Change struct {
	Table string
	RowId int64
}

type Database struct {
	base *sql.DB

	ChangeCh chan Change
}

func Open(name string) (*Database, func() error, error) {
	changeCh := make(chan Change)

	listenerId, err := gonanoid.Nanoid()
	if err != nil {
		return nil, nil, err
	}

	registerDriver()
	registerListener(listenerId, func(op int, table string, rowId int64) {
		// listen only for these tables
		if table != KVChangesTable && table != KVAccessChangesTable {
			return
		}

		// ignore delete events in _changes tables as they are cleanup
		if op == sqlite3.SQLITE_DELETE {
			return
		}

		changeCh <- Change{
			Table: table,
			RowId: rowId,
		}
	})

	db, err := sql.Open(driverName, name)
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return &Database{
			base:     db,
			ChangeCh: changeCh,
		}, func() error {
			unregisterListener(listenerId)
			return db.Close()
		}, nil
}

// simple way to apply scripts, should not be used in real database
func (d *Database) ApplySqlScripts() error {
	row := d.base.QueryRow(`SELECT count(name) FROM sqlite_master WHERE type='table' AND (name='kv_changes' OR name='kv_access_changes')`)

	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count == 2 {
		return nil
	}

	tx, err := d.base.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(kvSql); err != nil {
		return err
	}

	if _, err := tx.Exec(kvAccessSql); err != nil {
		return err
	}

	if _, err := tx.Exec(triggersSql); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *Database) Cleanup() error {
	_, err := d.base.Exec(`DELETE FROM kv_changes`)
	if err != nil {
		return err
	}
	_, err = d.base.Exec(`DELETE FROM kv_access_changes`)
	return err
}

func (d *Database) Exec(sqlTxt string) error {
	_, err := d.base.Exec(sqlTxt)
	return err
}
