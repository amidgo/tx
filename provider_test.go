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
	CaseName      string
	Provider      func(t *testing.T) transaction.Provider
	WithTx        func(ctx context.Context) error
	Opts          sql.TxOptions
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
			CaseName: "failed begin tx",
			Provider: func(t *testing.T) transaction.Provider {
				provider := mocks.NewProvider(t)

				opts := sql.TxOptions{
					Isolation: sql.LevelReadCommitted,
					ReadOnly:  true,
				}

				provider.ExpectBeginTxAndReturnError(errBeginTx, opts)

				return provider
			},
			Opts: sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
				ReadOnly:  true,
			},
			ExpectedError: errBeginTx,
		},
		&WithProviderTest{
			CaseName: "with tx returned error",
			Provider: func(t *testing.T) transaction.Provider {
				provider := mocks.NewProvider(t)
				tx := mocks.NewTransaction(t)

				provider.ExpectBeginTxAndReturnTx(tx, opts)
				tx.ExpectRollback()

				return provider
			},
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				return errWithTx
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName: "with tx paniced",
			Provider: func(t *testing.T) transaction.Provider {
				provider := mocks.NewProvider(t)
				tx := mocks.NewTransaction(t)

				provider.ExpectBeginTxAndReturnTx(tx, opts)
				tx.ExpectRollback()

				return provider
			},
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				panic(errWithTx)
			},
			Opts:          opts,
			ExpectedError: errWithTx,
		},
		&WithProviderTest{
			CaseName: "commit returned error",
			Provider: func(t *testing.T) transaction.Provider {
				provider := mocks.NewProvider(t)
				tx := mocks.NewTransaction(t)

				opts := sql.TxOptions{
					Isolation: sql.LevelReadCommitted,
					ReadOnly:  true,
				}

				provider.ExpectBeginTxAndReturnTx(tx, opts)

				tx.ExpectRollbackAfterFailedCommit(errCommit)

				return provider
			},
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				return nil
			},
			Opts:          opts,
			ExpectedError: errCommit,
		},
		&WithProviderTest{
			CaseName: "commit success",
			Provider: func(t *testing.T) transaction.Provider {
				provider := mocks.NewProvider(t)
				tx := mocks.NewTransaction(t)

				opts := sql.TxOptions{
					Isolation: sql.LevelReadCommitted,
					ReadOnly:  true,
				}

				provider.ExpectBeginTxAndReturnTx(tx, opts)

				tx.ExpectCommit()

				return provider
			},
			WithTx: func(ctx context.Context) error {
				require.True(t, transaction.TxEnabled(ctx))

				return nil
			},
			Opts:          opts,
			ExpectedError: nil,
		},
	)
}
