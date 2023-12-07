package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"
)

type KV struct {
	Key       string
	Value     string
	UpdatedAt time.Time
	CreatedAt time.Time
}

func (d *Database) Read(ctx context.Context, name string, version *int64) (*ReadRes, error) {
	g, gCtx := errgroup.WithContext(ctx)

	res := ReadRes{}
	g.Go(func() error {
		return d.getData(gCtx, name, version, res.setValue)
	})

	if version != nil {
		g.Go(func() error {
			return d.getAllRemovedIds(gCtx, name, version, res.setRemovedIds)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &res, nil
}

func (d *Database) getAllRemovedIds(ctx context.Context, name string, version *int64, setFunc func(map[string]int64)) error {
	var rows *sql.Rows
	var err error
	if name == KVTable {
		rows, err = d.base.QueryContext(ctx, `select a.kv_key, a.updated_at from kv_access a where a.access IN ($1, $2) 
                                  AND strftime('%s', a.updated_at) * 1000 >= $3`, AccessBlocked, AccessDeleted, *version)
	} else {
		return errors.New(fmt.Sprintf("collection: %s does not exist", name))
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var id string
		var updatedAt time.Time
		if err := rows.Scan(&id, &updatedAt); err != nil {
			return err
		}
		res[id] = updatedAt.UnixMilli()
	}
	setFunc(res)
	return nil
}

func (d *Database) getData(ctx context.Context, name string, version *int64, setFunc func([]ReadVal)) error {
	switch name {
	case KVTable:
		return d.getKVData(ctx, version, setFunc)
	}
	return errors.New(fmt.Sprintf("collection: %s does not exist", name))
}

func (d *Database) getKVData(ctx context.Context, version *int64, setFunc func([]ReadVal)) error {
	var rows *sql.Rows
	var err error

	if version != nil {
		rows, err = d.base.QueryContext(ctx, `select k.key, k.value, k.created_at, k.updated_at from kv_access ka inner join kv k on k.key = ka.kv_key 
                                                 where ka.access == $1 and strftime('%s', k.updated_at) * 1000 > $2`, AccessAllowed, *version)
	} else {
		rows, err = d.base.QueryContext(ctx, `select k.key, k.value, k.created_at, k.updated_at from kv_access ka inner join kv k on k.key = ka.kv_key 
                                                 where ka.access == $1`, AccessAllowed)
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	values := make([]ReadVal, 0)
	for rows.Next() {
		var value KV
		if err := rows.Scan(&value.Key, &value.Value, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return err
		}
		values = append(values, ReadVal{
			Id:        value.Key,
			CreatedAt: value.CreatedAt,
			UpdatedAt: value.UpdatedAt,
			Value:     value.Value,
		})
	}
	setFunc(values)
	return nil
}

func (d *Database) ReadKV(key string) (*KV, error) {
	row := d.base.QueryRow(`select key, value, created_at, updated_at from kv where key = $1`, key)

	var c KV
	if err := row.Scan(&c.Key, &c.Value, &c.UpdatedAt, &c.CreatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}
