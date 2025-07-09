package tx

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrSerialization = errors.New("serialization error")
	ErrCommit        = errors.New("commit error")
	ErrBeginTx       = errors.New("begin tx error")
)

type Tx interface {
	Context() context.Context
	CommitRollbacker
}

type contextWrapper struct {
	CommitRollbacker
}

func (contextWrapper) Context() context.Context {
	return context.Background()
}

type CommitRollbacker interface {
	Commit() error
	Rollback() error
}

type Beginner interface {
	Begin(ctx context.Context) (Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type Driver interface {
	Error(err error) error
}

func getDriver(x any) (Driver, bool) {
	driver, ok := x.(interface{ Driver() Driver })
	if !ok {
		return nil, false
	}

	return driver.Driver(), true
}

type driverBeginner struct {
	Beginner
	driver Driver
}

type driverTx struct {
	Tx
	driver Driver
}

func (d driverTx) Driver() Driver {
	return d.driver
}

func (d driverBeginner) Driver() Driver {
	return d.driver
}

func (d driverBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := d.Beginner.BeginTx(ctx, opts)

	return TxWithDriver(tx, d.driver), err
}

func (d driverBeginner) Begin(ctx context.Context) (Tx, error) {
	tx, err := d.Beginner.Begin(ctx)

	return TxWithDriver(tx, d.driver), err
}

func BeginnerWithDriver(beginner Beginner, driver Driver) Beginner {
	return driverBeginner{
		Beginner: beginner,
		driver:   driver,
	}
}

func TxWithDriver(tx Tx, driver Driver) Tx {
	return driverTx{
		Tx:     tx,
		driver: driver,
	}
}

func CommitRollbackerWithDriver(tx CommitRollbacker, driver Driver) CommitRollbacker {
	return driverTx{
		Tx:     contextWrapper{tx},
		driver: driver,
	}
}
