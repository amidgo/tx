package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/amidgo/transaction"
	mocks "github.com/amidgo/transaction/mocks"
)

type WithProviderTest struct {
	CaseName      string
	Provider      mocks.ProviderMock
	WithTx        func(t *testing.T, ctx context.Context) error
	Opts          *sql.TxOptions
	ExpectedError error
}

func (w *WithProviderTest) Name() string {
	return w.CaseName
}

func (w *WithProviderTest) Test(t *testing.T) {
	provider := w.Provider(t)
	defer func() {
		err := recover()
		if err == nil {
			return
		}

		if !errors.Is(err.(error), w.ExpectedError) {
			t.Fatalf(
				"unexpected error from panic recover, expected %+v, actual %+v",
				w.ExpectedError,
				err,
			)
		}
	}()

	withTx := func(context.Context) error { return nil }
	if w.WithTx != nil {
		withTx = func(ctx context.Context) error {
			return w.WithTx(t, ctx)
		}
	}

	err := transaction.WithProvider(
		context.Background(),
		provider,
		withTx,
		w.Opts,
	)
	if !errors.Is(err, w.ExpectedError) {
		t.Fatalf(
			"unexpected error from transaction.WithProvider, expected %+v, actual %+v",
			w.ExpectedError,
			err,
		)
	}
}

func runWithProviderTests(t *testing.T, tests ...*WithProviderTest) {
	for _, tst := range tests {
		t.Run(tst.Name(), tst.Test)
	}
}

func Test_WithProvider(t *testing.T) {
	var (
		errBeginTx = errors.New("begin tx")
		errWithTx  = errors.New("with tx")
		errCommit  = errors.New("commit")
	)

	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	runWithProviderTests(t,
		&WithProviderTest{
			CaseName:      "failed begin tx",
			Provider:      mocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			Opts:          opts,
			ExpectedError: errBeginTx,
		},
		&WithProviderTest{
			CaseName: "with tx returned error",
			Provider: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectRollback(nil), opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				return errWithTx
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName: "with tx paniced",
			Provider: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectRollback(nil), opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				panic(errWithTx)
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName: "commit returned error",
			Provider: mocks.ExpectBeginTxAndReturnTx(
				mocks.ExpectRollbackAfterFailedCommit(errCommit),
				opts,
			),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				return nil
			},
			Opts:          opts,
			ExpectedError: errCommit,
		},
		&WithProviderTest{
			CaseName: "commit success",
			Provider: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectCommit, opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				checkTxEnabled(t, ctx)

				return nil
			},
			Opts:          opts,
			ExpectedError: nil,
		},
	)
}

func checkTxEnabled(t *testing.T, ctx context.Context) {
	if !mocks.TxEnabled().Matches(ctx) {
		t.Fatalf("check tx fail, mocks.TxEnabled not matches ctx, %s", mocks.TxEnabled().String())
	}

	if mocks.TxDisabled().Matches(ctx) {
		t.Fatal("check tx fail, mocks.TxDisabled matches ctx")
	}
}
