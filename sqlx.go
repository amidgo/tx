package transaction

import (
	context "context"
	sql "database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
)

type sqlxTxKey struct{}

type SqlxTransaction struct {
	tx   *sqlx.Tx
	ctx  context.Context
	once sync.Once
}

func (s *SqlxTransaction) Context() context.Context {
	return s.ctx
}

func (s *SqlxTransaction) Commit(ctx context.Context) error {
	s.clearTx()

	return s.tx.Commit()
}

func (s *SqlxTransaction) Rollback(ctx context.Context) error {
	s.clearTx()

	return s.tx.Rollback()
}

func (s *SqlxTransaction) clearTx() {
	s.once.Do(func() { s.ctx = ClearTx(s.ctx) })
}

type SqlxProvider struct {
	db *sqlx.DB
}

func NewSqlxProvider(db *sqlx.DB) *SqlxProvider {
	return &SqlxProvider{
		db: db,
	}
}

func (s *SqlxProvider) Begin(ctx context.Context) (Transaction, error) {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	if err != nil {
		return nil, err
	}

	return &SqlxTransaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *SqlxProvider) BeginTx(ctx context.Context, opts sql.TxOptions) (Transaction, error) {
	tx, err := s.db.BeginTxx(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return &SqlxTransaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *SqlxProvider) transactionContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(StartTx(ctx), sqlxTxKey{}, tx)
}

func (s *SqlxProvider) Executor(ctx context.Context) SQLXExecutor {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *SqlxProvider) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *SqlxProvider) executor(ctx context.Context) (SQLXExecutor, bool) {
	if !TxEnabled(ctx) {
		return s.db, false
	}

	tx, ok := ctx.Value(sqlxTxKey{}).(*sqlx.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

type SQLXExecutor interface {
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
