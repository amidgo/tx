package buntransaction

import (
	"context"
	"database/sql"
	"sync"

	ttn "github.com/amidgo/transaction"
	"github.com/uptrace/bun"
)

type txKey struct{}

var _ ttn.Transaction = (*transaction)(nil)

type transaction struct {
	tx   *bun.Tx
	ctx  context.Context
	once sync.Once
}

func (s *transaction) Context() context.Context {
	return s.ctx
}

func (s *transaction) Commit() error {
	s.clearTx()

	return s.tx.Commit()
}

func (s *transaction) Rollback() error {
	s.clearTx()

	return s.tx.Rollback()
}

func (s *transaction) clearTx() {
	s.once.Do(func() {
		s.ctx = context.WithValue(s.ctx, txKey{}, nil)
	})
}

var _ ttn.Provider = (*Provider)(nil)

type Provider struct {
	db *bun.DB
}

func NewProvider(db *bun.DB) *Provider {
	return &Provider{
		db: db,
	}
}

func (s *Provider) Begin(ctx context.Context) (ttn.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &transaction{tx: &tx, ctx: s.transactionContext(ctx, &tx)}, nil
}

func (s *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &transaction{tx: &tx, ctx: s.transactionContext(ctx, &tx)}, nil
}

func (s *Provider) transactionContext(ctx context.Context, tx *bun.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func (s *Provider) Executor(ctx context.Context) Executor {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *Provider) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *Provider) executor(ctx context.Context) (Executor, bool) {
	tx, ok := ctx.Value(txKey{}).(*bun.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

func (s *Provider) WithTx(ctx context.Context, f func(ctx context.Context, exec Executor) error, opts *sql.TxOptions) error {
	exec, enabled := s.executor(ctx)
	if enabled {
		return f(ctx, exec)
	}

	return ttn.WithProvider(ctx, s,
		func(txContext context.Context) error {
			exec := s.Executor(txContext)

			return f(txContext, exec)
		},
		opts,
	)
}

type Executor interface {
	bun.IDB
}
