package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/amidgo/tester"
	"github.com/amidgo/transaction"
	"github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

type WithProviderTest struct {
	CaseName         string
	ProviderTemplate mocks.ProviderMock
	WithTx           func(ctx context.Context) error
	Opts             sql.TxOptions
	ExpectedError    error
}

func (w *WithProviderTest) Name() string {
	return w.CaseName
}

func (w *WithProviderTest) Test(t *testing.T) {
	provider := w.ProviderTemplate(t)
	defer func() {
		err := recover()
		if err != nil {
			require.ErrorIs(t, w.ExpectedError, err.(error))
		}
	}()

	withTx := func(context.Context) error { return nil }
	if w.WithTx != nil {
		withTx = w.WithTx
	}

	err := transaction.WithProvider(
		context.Background(),
		provider,
		withTx,
		w.Opts,
	)
	require.ErrorIs(t, err, w.ExpectedError)
}

func Test_WithProvider(t *testing.T) {
	var (
		errBeginTx = errors.New("begin tx")
		errWithTx  = errors.New("with tx")
		errCommit  = errors.New("commit")
	)

	opts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	tester.RunNamedTesters(t,
		&WithProviderTest{
			CaseName:         "failed begin tx",
			ProviderTemplate: mocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			Opts:             opts,
			ExpectedError:    errBeginTx,
		},
		&WithProviderTest{
			CaseName:         "with tx returned error",
			ProviderTemplate: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectRollback(nil), opts),
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				return errWithTx
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName:         "with tx paniced",
			ProviderTemplate: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectRollback(nil), opts),
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				panic(errWithTx)
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName: "commit returned error",
			ProviderTemplate: mocks.ExpectBeginTxAndReturnTx(
				mocks.ExpectRollbackAfterFailedCommit(errCommit),
				opts,
			),
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				return nil
			},
			Opts:          opts,
			ExpectedError: errCommit,
		},
		&WithProviderTest{
			CaseName:         "commit success",
			ProviderTemplate: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectCommit(), opts),
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				return nil
			},
			Opts:          opts,
			ExpectedError: nil,
		},
	)
}
