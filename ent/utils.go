package ent

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
)

func (c *Client) DoTransaction(ctx context.Context, opts *sql.TxOptions, f func(ctx context.Context, tx *Tx) error) error {
	tx, err := c.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	ctx = NewTxContext(ctx, tx)

	if err := f(ctx, tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: %v", err, rerr)
		}
		return err
	}

	return tx.Commit()
}
