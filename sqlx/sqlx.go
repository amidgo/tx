package sqlxtransaction

import (
	"context"
	"database/sql"
	"sync"

	ttn "github.com/amidgo/transaction"
	"github.com/jmoiron/sqlx"
)

type txKey struct{}

var _ ttn.Transaction = (*transaction)(nil)

type transaction struct {
	tx   *sqlx.Tx
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

type Provider struct {
	db *sqlx.DB
}

func NewProvider(db *sqlx.DB) *Provider {
	return &Provider{
		db: db,
	}
}

func (s *Provider) Begin(ctx context.Context) (ttn.Transaction, error) {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Transaction, error) {
	tx, err := s.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *Provider) transactionContext(ctx context.Context, tx *sqlx.Tx) context.Context {
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
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
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
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
