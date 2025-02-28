package sqlxtx

import (
	"context"
	"database/sql"
	"sync"

	ttn "github.com/amidgo/tx"
	"github.com/jmoiron/sqlx"
)

type txKey struct{}

var _ ttn.Tx = (*tx)(nil)

type tx struct {
	sqlxTx *sqlx.Tx

	ctx  context.Context
	once sync.Once
}

func (s *tx) Context() context.Context {
	return s.ctx
}

func (s *tx) Commit() error {
	s.clearTx()

	return s.sqlxTx.Commit()
}

func (s *tx) Rollback() error {
	s.clearTx()

	return s.sqlxTx.Rollback()
}

func (s *tx) clearTx() {
	s.once.Do(func() {
		s.ctx = context.WithValue(s.ctx, txKey{}, nil)
	})
}

type Beginner struct {
	db *sqlx.DB
}

func NewBeginner(db *sqlx.DB) *Beginner {
	return &Beginner{
		db: db,
	}
}

func (s *Beginner) Begin(ctx context.Context) (ttn.Tx, error) {
	sqlxTx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	if err != nil {
		return nil, err
	}

	return &tx{
		sqlxTx: sqlxTx,
		ctx:    s.txContext(ctx, sqlxTx),
	}, nil
}

func (s *Beginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Tx, error) {
	sqlxTx, err := s.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &tx{
		sqlxTx: sqlxTx,
		ctx:    s.txContext(ctx, sqlxTx),
	}, nil
}

func (s *Beginner) txContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
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
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
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
