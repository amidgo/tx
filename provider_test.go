package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/amidgo/tester"
	"github.com/amidgo/transaction"
	mocks "github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
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
		if err != nil {
			require.ErrorIs(t, w.ExpectedError, err.(error))
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
	require.ErrorIs(t, err, w.ExpectedError)
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

	tester.RunNamedTesters(t,
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
				require.True(t, mocks.TxEnabled().Matches(ctx))

				return errWithTx
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName: "with tx paniced",
			Provider: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectRollback(nil), opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				require.True(t, mocks.TxEnabled().Matches(ctx))

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
				require.True(t, mocks.TxEnabled().Matches(ctx))

				return nil
			},
			Opts:          opts,
			ExpectedError: errCommit,
		},
		&WithProviderTest{
			CaseName: "commit success",
			Provider: mocks.ExpectBeginTxAndReturnTx(mocks.ExpectCommit, opts),
			WithTx: func(t *testing.T, ctx context.Context) error {
				require.True(t, mocks.TxEnabled().Matches(ctx))

				return nil
			},
			Opts:          opts,
			ExpectedError: nil,
		},
	)
}
