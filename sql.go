package transaction

import (
	"context"
	"database/sql"
	"sync"
)

type sqlTxKey struct{}

type SQLTransaction struct {
	tx   *sql.Tx
	ctx  context.Context
	once sync.Once
}

func (s *SQLTransaction) Context() context.Context {
	return s.ctx
}

func (s *SQLTransaction) Commit(ctx context.Context) error {
	s.clearTx()

	return s.tx.Commit()
}

func (s *SQLTransaction) Rollback(ctx context.Context) error {
	s.clearTx()

	return s.tx.Rollback()
}

func (s *SQLTransaction) clearTx() {
	s.once.Do(func() { s.ctx = ClearTx(s.ctx) })
}

type SQLProvider struct {
	db *sql.DB
}

func NewSQLProvider(db *sql.DB) *SQLProvider {
	return &SQLProvider{db: db}
}

func (s *SQLProvider) Begin(ctx context.Context) (Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &SQLTransaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *SQLProvider) BeginTx(ctx context.Context, opts sql.TxOptions) (Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return &SQLTransaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *SQLProvider) transactionContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(StartTx(ctx), sqlTxKey{}, tx)
}

func (s *SQLProvider) Executor(ctx context.Context) SQLExecutor {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *SQLProvider) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *SQLProvider) executor(ctx context.Context) (SQLExecutor, bool) {
	if !TxEnabled(ctx) {
		return s.db, false
	}

	tx, ok := ctx.Value(sqlTxKey{}).(*sql.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

type SQLExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
