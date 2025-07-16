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

var _ ttn.Beginner = (*Beginner)(nil)

type Beginner struct {
	db *bun.DB
}

func NewBeginner(db *bun.DB) *Beginner {
	return &Beginner{
		db: db,
	}
}

func (s *Beginner) Begin(ctx context.Context) (ttn.Tx, error) {
	bunTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &tx{
		bunTx: bunTx,
		ctx:   s.txContext(ctx, bunTx),
	}, nil
}

func (s *Beginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Tx, error) {
	bunTx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &tx{
		bunTx: bunTx,
		ctx:   s.txContext(ctx, bunTx),
	}, nil
}

func (s *Beginner) txContext(ctx context.Context, bunTx bun.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, bunTx)
}

func (s *Beginner) Executor(ctx context.Context) Executor {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *Beginner) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *Beginner) executor(ctx context.Context) (Executor, bool) {
	tx, ok := ctx.Value(txKey{}).(bun.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

func (s *Beginner) WithTx(
	ctx context.Context,
	f func(ctx context.Context, exec Executor) error,
	txOpts *sql.TxOptions,
	opts ...ttn.Option,
) error {
	return ttn.Run(ctx, s,
		func(txContext context.Context) error {
			exec := s.Executor(txContext)

			// must be tx without executor
			return f(ctx, exec)
		},
		txOpts,
		opts...,
	)
}

type Executor interface {
	bun.IDB
}
