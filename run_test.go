package tx_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"

	"github.com/amidgo/tx"
	txmocks "github.com/amidgo/tx/mocks"
)

type runTest struct {
	Name           string
	Beginner       txmocks.BeginnerMock
	WithTx         func(t *testing.T, ctx context.Context) error
	Opts           *sql.TxOptions
	ExpectedErrors []error
}

func (w *runTest) Test(t *testing.T) {
	beginner := w.Beginner(t)
	defer func() {
		recover := recover()
		if recover == nil {
			return
		}

		err := recover.(error)

		if len(w.ExpectedErrors) == 0 {
			if err != nil {
				t.Fatalf("unexpected error, %s", err)
			}
		}

		for _, expectedErr := range w.ExpectedErrors {
			if !errors.Is(err, expectedErr) {
				t.Fatalf(
					"unexpected error from panic recover, expected %+v, actual %+v",
					expectedErr,
					err,
				)
			}
		}
	}()

	withTx := func(context.Context) error { return nil }
	if w.WithTx != nil {
		withTx = func(ctx context.Context) error {
			return w.WithTx(t, ctx)
		}
	}

	err := tx.Run(
		context.Background(),
		beginner,
		withTx,
		w.Opts,
	)

	if len(w.ExpectedErrors) == 0 && err != nil {
		t.Fatalf("unexpected error, %s", err)
	}

	for _, expectedErr := range w.ExpectedErrors {
		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"unexpected error from tx.Run, expected %+v, actual %+v",
				expectedErr,
				err,
			)
		}
	}

}

func Test_Run(t *testing.T) {
	var (
		errBeginTx = errors.New("begin tx")
		errWithTx  = errors.New("with tx")
		errCommit  = errors.New("commit")
	)

	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	tests := []*runTest{
		{
			Name:           "failed begin tx",
			Beginner:       txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			Opts:           opts,
			ExpectedErrors: []error{tx.ErrBeginTx, errBeginTx},
		},
		{
			Name:     "with tx returned error",
			Beginner: txmocks.ExpectBeginTxAndReturnTx(txmocks.ExpectRollback(nil), opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				return errWithTx
			},
			Opts:           opts,
			ExpectedErrors: []error{errWithTx},
		},
		{
			Name:     "with tx paniced",
			Beginner: txmocks.ExpectBeginTxAndReturnTx(txmocks.ExpectRollback(nil), opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				panic(errWithTx)
			},
			Opts:           opts,
			ExpectedErrors: []error{errWithTx},
		},
		{
			Name: "commit returned error",
			Beginner: txmocks.ExpectBeginTxAndReturnTx(
				txmocks.ExpectRollbackAfterFailedCommit(errCommit),
				opts,
			),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				return nil
			},
			Opts:           opts,
			ExpectedErrors: []error{tx.ErrCommit, errCommit},
		},
		{
			Name:     "commit success",
			Beginner: txmocks.ExpectBeginTxAndReturnTx(txmocks.ExpectCommit, opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				return nil
			},
			Opts: opts,
		},
	}

	for _, tst := range tests {
		t.Run(tst.Name, tst.Test)
	}
}

type runDriverTest struct {
	Name             string
	DriverMock       txmocks.DriverMock
	BeginnerMock     txmocks.BeginnerMock
	TxOpts           *sql.TxOptions
	Opts             []tx.Option
	WithTx           func(t *testing.T, ctx context.Context) error
	ExpectedErrors   []error
	UnexpectedErrors []error
}

func (w *runDriverTest) Test(t *testing.T) {
	beginner := tx.BeginnerWithDriver(
		w.BeginnerMock(t),
		w.DriverMock(t),
	)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	withTx := func(txContext context.Context) error {
		return w.WithTx(t, txContext)
	}

	err := tx.Run(ctx,
		beginner,
		withTx,
		w.TxOpts,
		w.Opts...,
	)

	if len(w.ExpectedErrors) == 0 && err != nil {
		t.Fatalf("expected no error, actual %+v", err)
	}

	for _, expectedErr := range w.ExpectedErrors {
		if !errors.Is(err, expectedErr) {
			t.Fatalf("unexpected error, expect %+v, actual %+v", expectedErr, err)
		}
	}

	for _, unexpectedErr := range w.UnexpectedErrors {
		if errors.Is(err, unexpectedErr) {
			t.Fatalf("unexpected error, unexpect %+v, actual %+v", unexpectedErr, err)
		}
	}
}

func Test_Run_Driver(t *testing.T) {
	withTx := func(count int, err error) func(t *testing.T, ctx context.Context) error {
		called := 0

		return func(t *testing.T, ctx context.Context) error {
			checkTxEnabled(t, ctx)

			if called == count {
				return nil
			}

			called++

			return err
		}
	}

	txOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}

	tests := []*runDriverTest{
		{
			Name:       "success, no driver calls occured",
			DriverMock: txmocks.NilDriver,
			BeginnerMock: txmocks.ExpectBeginTxAndReturnTx(
				txmocks.ExpectCommit,
				nil,
			),
			WithTx:         withTx(0, nil),
			ExpectedErrors: []error{},
		},
		{
			Name: "failed begin tx",
			DriverMock: txmocks.ExpectDriverError(
				errors.Is,
				io.ErrUnexpectedEOF,
				io.ErrUnexpectedEOF,
			),
			BeginnerMock: txmocks.ExpectBeginTxAndReturnError(
				io.ErrUnexpectedEOF,
				nil,
			),
			WithTx:           func(t *testing.T, ctx context.Context) error { return nil },
			ExpectedErrors:   []error{tx.ErrBeginTx, io.ErrUnexpectedEOF},
			UnexpectedErrors: []error{tx.ErrCommit},
		},
		{
			Name: "failed commit tx",
			DriverMock: txmocks.ExpectDriverError(
				errors.Is,
				io.ErrUnexpectedEOF,
				io.ErrUnexpectedEOF,
			),
			BeginnerMock: txmocks.ExpectBeginTxAndReturnTx(
				txmocks.ExpectRollbackAfterFailedCommit(
					io.ErrUnexpectedEOF,
				),
				nil,
			),
			WithTx:           func(t *testing.T, ctx context.Context) error { return nil },
			ExpectedErrors:   []error{tx.ErrCommit, io.ErrUnexpectedEOF},
			UnexpectedErrors: []error{tx.ErrBeginTx},
		},
		{
			Name: "failed withTx",
			DriverMock: txmocks.ExpectDriverError(
				errors.Is,
				io.ErrUnexpectedEOF,
				io.ErrUnexpectedEOF,
			),
			BeginnerMock: txmocks.ExpectBeginTxAndReturnTx(
				txmocks.ExpectRollback(nil),
				nil,
			),
			WithTx:           withTx(1, io.ErrUnexpectedEOF),
			ExpectedErrors:   []error{io.ErrUnexpectedEOF},
			UnexpectedErrors: []error{tx.ErrCommit, tx.ErrCommit},
		},
		{
			Name: "failed withTx, serialization error, but no opts provided",
			DriverMock: txmocks.JoinDrivers(
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
			),
			BeginnerMock: txmocks.JoinBeginners(
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					nil,
				),
			),
			WithTx:           withTx(1, io.ErrUnexpectedEOF),
			ExpectedErrors:   []error{tx.ErrSerialization},
			UnexpectedErrors: []error{tx.ErrCommit, tx.ErrCommit},
		},
		{
			Name: "failed commit, serialization error, but no opts provided",
			DriverMock: txmocks.ExpectDriverError(
				errors.Is,
				io.ErrUnexpectedEOF,
				tx.ErrSerialization,
			),
			BeginnerMock: txmocks.ExpectBeginTxAndReturnTx(
				txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
				nil,
			),
			WithTx:           func(t *testing.T, ctx context.Context) error { return nil },
			ExpectedErrors:   []error{tx.ErrCommit, tx.ErrSerialization},
			UnexpectedErrors: []error{tx.ErrBeginTx},
		},
		{
			Name: "serialization error, opts provided, endless retry",
			DriverMock: txmocks.JoinDrivers(
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
			),
			BeginnerMock: txmocks.JoinBeginners(
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectCommit,
					txOpts,
				),
			),
			WithTx: withTx(2, io.ErrUnexpectedEOF),
			TxOpts: txOpts,
			Opts: []tx.Option{
				tx.RetrySerialization(-1),
			},
		},
		{
			Name: "serialization error, opts provided, limited retry",
			DriverMock: txmocks.JoinDrivers(
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
			),
			BeginnerMock: txmocks.JoinBeginners(
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
					txOpts,
				),
			),
			WithTx: withTx(2, io.ErrUnexpectedEOF),
			TxOpts: txOpts,
			Opts: []tx.Option{
				tx.RetrySerialization(3),
			},
			ExpectedErrors:   []error{tx.ErrCommit, tx.ErrSerialization},
			UnexpectedErrors: []error{tx.ErrBeginTx},
		},
		{
			Name: "serialization error, opts provided, endless retry, beginTx failed",
			DriverMock: txmocks.JoinDrivers(
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					tx.ErrSerialization,
				),
				txmocks.ExpectDriverError(
					errors.Is,
					io.ErrUnexpectedEOF,
					errors.Join(io.ErrShortWrite, io.ErrUnexpectedEOF),
				),
			),
			BeginnerMock: txmocks.JoinBeginners(
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollback(nil),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnTx(
					txmocks.ExpectRollbackAfterFailedCommit(io.ErrUnexpectedEOF),
					txOpts,
				),
				txmocks.ExpectBeginTxAndReturnError(io.ErrUnexpectedEOF, txOpts),
			),
			WithTx: withTx(2, io.ErrUnexpectedEOF),
			TxOpts: txOpts,
			Opts: []tx.Option{
				tx.RetrySerialization(-1),
			},
			ExpectedErrors:   []error{tx.ErrBeginTx, io.ErrShortWrite, io.ErrUnexpectedEOF},
			UnexpectedErrors: []error{tx.ErrCommit},
		},
	}

	for _, tst := range tests {
		t.Run(tst.Name, tst.Test)
	}
}

func checkTxEnabled(t *testing.T, ctx context.Context) {
	if !txmocks.TxEnabled().Matches(ctx) {
		t.Fatalf("check tx fail, mocks.TxEnabled not matches ctx, %s", txmocks.TxEnabled().String())
	}

	if txmocks.TxDisabled().Matches(ctx) {
		t.Fatal("check tx fail, mocks.TxDisabled matches ctx")
	}
}
