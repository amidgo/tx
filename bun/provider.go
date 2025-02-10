package buntx

import (
	"context"
	"database/sql"
	"sync"

	ttn "github.com/amidgo/tx"
	"github.com/uptrace/bun"
)

type txKey struct{}

var _ ttn.Tx = (*tx)(nil)

type tx struct {
	bunTx bun.Tx

	ctx  context.Context
	once sync.Once
}

func (s *tx) Context() context.Context {
	return s.ctx
}

func (s *tx) Commit() error {
	s.clearTx()

	return s.bunTx.Commit()
}

func (s *tx) Rollback() error {
	s.clearTx()

	return s.bunTx.Rollback()
}

func (s *tx) clearTx() {
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

func (s *Provider) Begin(ctx context.Context) (ttn.Tx, error) {
	bunTx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &tx{
		bunTx: bunTx,
		ctx:   s.txContext(ctx, bunTx),
	}, nil
}

func (s *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Tx, error) {
	bunTx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &tx{
		bunTx: bunTx,
		ctx:   s.txContext(ctx, bunTx),
	}, nil
}

func (s *Provider) txContext(ctx context.Context, bunTx bun.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, bunTx)
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
	tx, ok := ctx.Value(txKey{}).(bun.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

func (s *Provider) WithTx(
	ctx context.Context,
	f func(ctx context.Context, exec Executor) error,
	txOpts *sql.TxOptions,
	opts ...ttn.Option,
) error {
	return ttn.WithTx(ctx, s,
		func(txContext context.Context) error {
			exec := s.Executor(txContext)

			return f(txContext, exec)
		},
		txOpts,
		opts...,
	)
}

type Executor interface {
	bun.IDB
}
