package transaction

import (
	"context"
	"database/sql"
	"fmt"
)

type Transaction interface {
	Context() context.Context
	Commit() error
	Rollback() error
}

type Provider interface {
	Begin(ctx context.Context) (Transaction, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	TxEnabled(ctx context.Context) bool
}

func WithProvider(
	ctx context.Context,
	provider Provider,
	withTx func(txContext context.Context) error,
	opts *sql.TxOptions,
) error {
	tx, err := provider.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx, %w", err)
	}

	committed := false

	defer func() {
		if committed {
			return
		}

		_ = tx.Rollback()
	}()

	err = withTx(tx.Context())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx, %w", err)
	}

	committed = true

	return nil
}
