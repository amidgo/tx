package sqltx

import (
	"context"
	"database/sql"
	"sync"

	ttn "github.com/amidgo/tx"
)

type txKey struct{}

var _ ttn.Tx = (*tx)(nil)

type tx struct {
	sqlTx *sql.Tx

	ctx  context.Context
	once sync.Once
}

func (s *tx) Context() context.Context {
	return s.ctx
}

func (s *tx) Commit() error {
	s.clearTx()

	return s.sqlTx.Commit()
}

func (s *tx) Rollback() error {
	s.clearTx()

	return s.sqlTx.Rollback()
}

func (s *tx) clearTx() {
	s.once.Do(func() {
		s.ctx = context.WithValue(s.ctx, txKey{}, nil)
	})
}

type Beginner struct {
	db *sql.DB
}

func NewBeginner(db *sql.DB) *Beginner {
	return &Beginner{db: db}
}

func (s *Beginner) Begin(ctx context.Context) (ttn.Tx, error) {
	sqlTx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &tx{
		sqlTx: sqlTx,
		ctx:   s.txContext(ctx, sqlTx),
	}, nil
}

func (s *Beginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Tx, error) {
	sqlTx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &tx{
		sqlTx: sqlTx,
		ctx:   s.txContext(ctx, sqlTx),
	}, nil
}

func (s *Beginner) txContext(ctx context.Context, sqlTx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, sqlTx)
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
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

func (s *Beginner) WithTx(
	ctx context.Context,
	withTx func(ctx context.Context, exec Executor) error,
	txOpts *sql.TxOptions,
	opts ...ttn.Option,
) error {
	return ttn.Run(ctx, s,
		func(txContext context.Context) error {
			exec := s.Executor(txContext)

			return withTx(txContext, exec)
		},
		txOpts,
		opts...,
	)
}

type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
