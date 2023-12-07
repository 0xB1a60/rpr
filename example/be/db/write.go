package db

import (
	"context"
)

func (d *Database) AddKV(ctx context.Context, key string, value string) error {
	tx, err := d.base.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO kv (key, value)VALUES ($1, $2)`, key, value)
	if err != nil {
		return err
	}

	_, accessErr := tx.ExecContext(ctx, `INSERT INTO kv_access (kv_key, access)VALUES ($1, $2)`, key, AccessAllowed)
	if accessErr != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) EditKV(ctx context.Context, key string, newValue string) (*int64, error) {
	res, err := d.base.ExecContext(ctx, `UPDATE kv SET value=$1 WHERE key = $2`, newValue, key)
	if err != nil {
		return nil, err
	}

	changes, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &changes, nil
}

func (d *Database) RemoveKV(ctx context.Context, key string) error {
	_, err := d.base.ExecContext(ctx, `DELETE FROM kv WHERE key=$1`, key)
	return err
}

func (d *Database) RemoveKVAccess(ctx context.Context, key string) error {
	_, err := d.base.ExecContext(ctx, `UPDATE kv_access SET access = $1 where kv_key = $2`, AccessBlocked, key)
	return err
}
