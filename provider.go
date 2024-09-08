package transaction

import (
	"context"
	"database/sql"
	"fmt"
)

type Transaction interface {
	Context() context.Context
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Provider interface {
	Begin(ctx context.Context) (Transaction, error)
	BeginTx(ctx context.Context, opts sql.TxOptions) (Transaction, error)
}

func WithProvider(
	ctx context.Context,
	provider Provider,
	withTx func(txContext context.Context) error,
	opts sql.TxOptions,
) error {
	tx, err := provider.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx, %w", err)
	}

	tx = &notRollbackAfterCommit{
		tx: tx,
	}

	defer func() { _ = tx.Rollback(ctx) }()

	err = withTx(tx.Context())
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit tx, %w", err)
	}

	return nil
}

type notRollbackAfterCommit struct {
	tx       Transaction
	commited bool
}

func (r *notRollbackAfterCommit) Rollback(ctx context.Context) error {
	if r.commited {
		return nil
	}

	return r.tx.Rollback(ctx)
}

func (r *notRollbackAfterCommit) Commit(ctx context.Context) error {
	err := r.tx.Commit(ctx)

	r.commited = (err == nil)

	return err
}

func (r *notRollbackAfterCommit) Context() context.Context {
	return r.tx.Context()
}
