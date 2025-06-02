package tx

import (
	"context"
	"database/sql"
	"errors"
)

type options struct {
	serializationRetryCount int
}

type Option func(*options)

func RetrySerialization(times int) Option {
	return func(o *options) {
		o.serializationRetryCount = times
	}
}

func txPipelineExec(
	ctx context.Context,
	beginner Beginner,
	withTx func(txContext context.Context) error,
	txOpts *sql.TxOptions,
	opts ...Option,
) func() error {
	pipeline := makeTxPipeline(ctx, beginner, withTx, txOpts)

	driver, _ := getDriver(beginner)

	if driver != nil {
		pipeline = useDriverToTxPipeline(pipeline, driver)
	}

	exec := pipeline.exec()

	options := &options{}

	for _, op := range opts {
		op(options)
	}

	if options.serializationRetryCount != 0 {
		exec = retrySerializationExec(exec, options.serializationRetryCount)
	}

	return exec
}

func retrySerializationExec(exec func() error, serializationRetryCount int) func() error {
	return func() error {
		err := exec()
		if !errors.Is(err, ErrSerialization) {
			return err
		}

		for i := serializationRetryCount; i != 0; i-- {
			err = exec()

			if errors.Is(err, ErrSerialization) {
				continue
			}

			return err
		}

		return errors.Join(errSerializationRepeatTimesExcedeed, err)
	}
}

func useDriverToTxPipeline(pipeline txPipeline, driver Driver) txPipeline {
	return txPipeline{
		begin: func() (Tx, error) {
			tx, err := pipeline.begin()
			err = driverError(driver, err)

			return tx, err
		},
		withTx: func(txContext context.Context) error {
			err := pipeline.withTx(txContext)

			return driverError(driver, err)
		},
		commit: func(tx Tx) error {
			err := pipeline.commit(tx)

			return driverError(driver, err)
		},
		rollback: pipeline.rollback,
	}
}

func makeTxPipeline(
	ctx context.Context,
	beginner Beginner,
	withTx func(txContext context.Context) error,
	txOpts *sql.TxOptions,
) txPipeline {
	return txPipeline{
		begin: func() (Tx, error) {
			return beginner.BeginTx(ctx, txOpts)
		},
		withTx: withTx,
		commit: func(tx Tx) error {
			return tx.Commit()
		},
		rollback: func(tx Tx) {
			_ = tx.Rollback()
		},
	}
}

type txPipeline struct {
	begin    func() (Tx, error)
	withTx   func(txContext context.Context) error
	commit   func(tx Tx) error
	rollback func(tx Tx)
}

func (t txPipeline) exec() func() error {
	return func() error {
		tx, err := t.begin()
		if err != nil {
			return errors.Join(ErrBeginTx, err)
		}

		committed := false

		defer func() {
			if committed {
				return
			}

			t.rollback(tx)
		}()

		err = t.withTx(tx.Context())
		if err != nil {
			return err
		}

		err = t.commit(tx)
		if err != nil {
			return errors.Join(ErrCommit, err)
		}

		committed = true

		return nil
	}
}

func Run(
	ctx context.Context,
	beginner Beginner,
	withTx func(txContext context.Context) error,
	txOpts *sql.TxOptions,
	opts ...Option,
) error {
	exec := txPipelineExec(
		ctx,
		beginner,
		withTx,
		txOpts,
		opts...,
	)

	return exec()
}

var errSerializationRepeatTimesExcedeed = errors.New("serialization repeat times exceeded")

func driverError(driver Driver, err error) error {
	if err == nil {
		return nil
	}

	if driver == nil {
		return err
	}

	return driver.Error(err)
}
