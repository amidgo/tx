package stdlibtransaction

import (
	"context"
	"database/sql"
	"sync"

	ttn "github.com/amidgo/transaction"
)

type txKey struct{}

type transaction struct {
	tx   *sql.Tx
	ctx  context.Context
	once sync.Once
}

func (s *transaction) Context() context.Context {
	return s.ctx
}

func (s *transaction) Commit(ctx context.Context) error {
	s.clearTx()

	return s.tx.Commit()
}

func (s *transaction) Rollback(ctx context.Context) error {
	s.clearTx()

	return s.tx.Rollback()
}

func (s *transaction) clearTx() {
	s.once.Do(func() { s.ctx = ttn.ClearTx(s.ctx) })
}

type Provider struct {
	db *sql.DB
}

func NewProvider(db *sql.DB) *Provider {
	return &Provider{db: db}
}

func (s *Provider) Begin(ctx context.Context) (ttn.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *Provider) BeginTx(ctx context.Context, opts sql.TxOptions) (ttn.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *Provider) transactionContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ttn.StartTx(ctx), txKey{}, tx)
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
	if !ttn.TxEnabled(ctx) {
		return s.db, false
	}

	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
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
