# Transaction Management Library

![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)
[![Go Reference](https://pkg.go.dev/badge/github.com/amidgo/transaction.svg)](https://pkg.go.dev/github.com/amidgo/transaction)

A flexible transaction management library for Go applications supporting multiple database drivers and ORMs.

## **Installation**

```bash
go get github.com/amidgo/transaction@v0.0.6
```

## **Usage**

```go
import (
	"context"
	"sql"
	"fmt"

	"github.com/amidgo/transaction"
)

type Repository interface {
	Foo(ctx context.Context) error
	Bar(ctx context.Context) error
}

type service struct {
	txProvider transaction.Provider

	repo Repository
}

func (s *service) TODOSomething(ctx context.Context) error {
	err := transaction.WithProvider(ctx,
		s.txProvider,
		func(txContext context.Context) error {
			err := repo.Foo(txContext)
			if err != nil {
				return fmt.Errorf("call repo.Foo, %w", err)
			}

			err = repo.Bar(txContext)
			if err != nil {
				return fmt.Errorf("call repo.Bar, %w", err)
			}

			return nil
		},
		&sql.TxOptions{
			Isolation: sql.LevelReadCommited,
		}
	)
	if err != nil {
		return fmt.Errorf("try to do something, %w", err)
	}

	return nil
}


// in your repo

type sqlRepo struct {
	provider *stdlibtransaction.Provider
}


func (s *sqlRepo) Foo(ctx context.Context) error {
	// if ctx contains txKey{} returns *sql.Tx else returns *sql.DB
	exec := s.provider.Executor(ctx)

	return s.foo(ctx, exec)
}

func (s *sqlRepo) foo(ctx context.Context, exec stdlibtransaction.Executor) error {
	exec.QueryRowContext(ctx, "query")
	exec.ExecContext(ctx, "query")

	...
}

func (s *sqlRepo) Bar(ctx context.Context) error {
	// if ctx contains txKey{} returns *sql.Tx else returns *sql.DB
	exec := s.provider.Executor(ctx)

	return s.bar(ctx, exec)
}

func (s *sqlRepo) bar(ctx context.Context, exec stdlibtransaction.Executor) error {
	exec.QueryRowContext(ctx, "query")
	exec.ExecContext(ctx, "query")

	...
}

// If you want to guarantee that the context will contain the transaction, use provider.WithTx
// DANGER: WithTx does not affect external calls, use wisely

func (s *sqlRepo) FooBar(ctx context.Context) error {
	return s.provider.WithTx(ctx,
		s.fooBar,
		&sql.TxOptions{
			Isolation: sql.LevelReadCommited,
		},
	)
}

func (s *sqlRepo) fooBar(ctx context.Context, exec stdlibtransaciton.Executor) error {
	err := s.foo(ctx, exec)
	if err != nil {
		return err
	}

	err = s.bar(ctx, exec)
	if err != nil {
		return err
	}

	return nil
}


// WithTx stdlibtransaction code example

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

func (s *Provider) Executor(ctx context.Context) Executor {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *Provider) executor(ctx context.Context) (Executor, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}

```

## **Core concepts**

call of function transaction.WithProvider begin transaction with provided opts, get transaction context and exec withTx func with this context

transaction.WithProvider ensures that the transaction will be rolled back in case of and error

```go
type Transaction interface {
	Context() context.Context
	Commit() error
	Rollback() error
}

type Provider interface {
	Begin(ctx context.Context) (Transaction, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	TxEnabled(ctx context.Context) bool
}

func WithProvider(
	ctx context.Context,
	provider Provider,
	withTx func(txContext context.Context) error,
	opts *sql.TxOptions,
) error {
	tx, err := provider.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx, %w", err)
	}

	committed := false

	defer func() {
		if committed {
			return
		}

		_ = tx.Rollback()
	}()

	err = withTx(tx.Context())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx, %w", err)
	}

	committed = true

	return nil
}
```

### **Provider**

transaction.Provider interface includes 3 methods

```go
type Provider interface {
	Begin(ctx context.Context) (Transaction, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	TxEnabled(ctx context.Context) bool
}
```

#### ***Begin and BeginTx***

Create new Transaction which include parent context with provider specific tx instance:

[stdlib provider example](https://github.com/amidgo/transaction/blob/v0.0.6/stdlib/stdlib.go#L51)
```go
type txKey struct{}

func (s *Provider) Begin(ctx context.Context) (ttn.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (ttn.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx, ctx: s.transactionContext(ctx, tx)}, nil
}

func (s *Provider) transactionContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
```

#### ***TxEnabled***

Checks that the context.Context contains a provider specific transaction

[stdlib provider example]
```go
type txKey struct{}

func (s *Provider) Executor(ctx context.Context) Executor {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *Provider) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *Provider) executor(ctx context.Context) (Executor, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}
```
