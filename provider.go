package tx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrSerialization = errors.New("serialization error")

type Tx interface {
	Context() context.Context
	Commit() error
	Rollback() error
}

type Provider interface {
	Begin(ctx context.Context) (Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	TxEnabled(ctx context.Context) bool
}

type Options struct {
	serializationRetryCount int
}

type Option func(*Options)

func RetrySerialization(times int) Option {
	return func(o *Options) {
		o.serializationRetryCount = times
	}
}

func WithTx(
	ctx context.Context,
	provider Provider,
	withTx func(txContext context.Context) error,
	txOpts *sql.TxOptions,
	opts ...Option,
) error {
	options := &Options{}

	for _, op := range opts {
		op(options)
	}

	driver, _ := getDriver(provider)

	tx, err := provider.BeginTx(ctx, txOpts)

	err = driverError(driver, err)
	if err != nil {
		return fmt.Errorf("begin tx, %w", err)
	}

	finished := false

	defer func() {
		if finished {
			return
		}

		_ = tx.Rollback()
	}()

	err = withTx(tx.Context())
	err = driverError(driver, err)

	switch {
	case options.serializationRetryCount != 0 && errors.Is(err, ErrSerialization):
		finished = true
		_ = tx.Rollback()

		retryErr := retry(ctx, driver, provider, withTx, txOpts, options.serializationRetryCount)
		if retryErr != nil {
			return fmt.Errorf("retry %w, %w, after %d retries", err, retryErr, options.serializationRetryCount)
		}

		return nil
	case err != nil:
		return err
	}

	err = tx.Commit()
	err = driverError(driver, err)

	switch {
	case options.serializationRetryCount != 0 && errors.Is(err, ErrSerialization):
		finished = true
		_ = tx.Rollback()

		retryErr := retry(ctx, driver, provider, withTx, txOpts, options.serializationRetryCount)
		if retryErr != nil {
			return fmt.Errorf("retry %w, %w, after %d retries", err, retryErr, options.serializationRetryCount)
		}

		return nil
	case err != nil:
		return err
	}

	finished = true

	return nil
}

var errRepeatTimesExcedeed = errors.New("repeat times exceeded")

func retry(
	ctx context.Context,
	driver Driver,
	provider Provider,
	withTx func(ctx context.Context) error,
	txOpts *sql.TxOptions,
	repeatTimes int,
) error {
	if repeatTimes == 0 {
		return errRepeatTimesExcedeed
	}

	tx, err := provider.BeginTx(ctx, txOpts)

	err = driverError(driver, err)
	if err != nil {
		return err
	}

	finished := false

	defer func() {
		if finished {
			return
		}

		_ = tx.Rollback()
	}()

	err = withTx(tx.Context())
	err = driverError(driver, err)

	switch {
	case errors.Is(err, ErrSerialization):
		finished = true
		_ = tx.Rollback()
		repeatTimes--

		retryErr := retry(ctx, driver, provider, withTx, txOpts, repeatTimes)
		if retryErr != nil {
			return retryErr
		}

		return nil
	case err != nil:
		return err
	}

	err = tx.Commit()
	err = driverError(driver, err)

	switch {
	case errors.Is(err, ErrSerialization):
		finished = true
		_ = tx.Rollback()
		repeatTimes--

		retryErr := retry(ctx, driver, provider, withTx, txOpts, repeatTimes)
		if retryErr != nil {
			return retryErr
		}

		return nil
	case err != nil:
		return err
	}

	finished = true

	return nil
}

func driverError(driver Driver, err error) error {
	if err == nil {
		return nil
	}

	if driver == nil {
		return err
	}

	return driver.Error(err)
}
