package db

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"time"
)

const (
	driverName = "sqlite3_enhanced"
)

type cdcListener func(op int, table string, rowid int64)

var (
	cdcListeners = make(map[string]cdcListener)
)

func registerListener(id string, listener cdcListener) {
	cdcListeners[id] = listener
}

func unregisterListener(id string) {
	delete(cdcListeners, id)
}

func registerDriver() {
	for _, driver := range sql.Drivers() {
		if driver == driverName {
			return
		}
	}

	sql.Register(driverName,
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conn.RegisterUpdateHook(func(op int, db string, table string, rowid int64) {
					// run it in a goroutine as this function blocks the database
					go func() {
						// wait 5 milliseconds as the change event is fired before the insert
						time.Sleep(5 * time.Millisecond)
						for _, listener := range cdcListeners {
							listener(op, table, rowid)
						}
					}()
				})
				return nil
			},
		})
}
