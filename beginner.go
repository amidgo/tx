package tx

import (
	"context"
	"database/sql"
	"errors"
)

var ErrSerialization = errors.New("serialization error")

type Tx interface {
	Context() context.Context
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

func getDriver(provider Beginner) (Driver, bool) {
	driver, ok := provider.(interface{ Driver() Driver })
	if !ok {
		return nil, false
	}

	return driver.Driver(), true
}

type driverBeginner struct {
	Beginner
	driver Driver
}

func (d *driverBeginner) Driver() Driver {
	return d.driver
}

func BeginnerWithDriver(provider Beginner, driver Driver) Beginner {
	return &driverBeginner{
		Beginner: provider,
		driver:   driver,
	}
}
