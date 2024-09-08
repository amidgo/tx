package mocks_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

type mockTestReporter struct {
	t      *testing.T
	called bool
}

func newMockTestReporter(t *testing.T, expectCalled bool) *mockTestReporter {
	r := &mockTestReporter{t: t}

	t.Cleanup(
		func() {
			require.Equal(t, expectCalled, r.called)
		},
	)

	return r
}

func (r *mockTestReporter) Fatal(...any) {
	r.called = true
}

func (r *mockTestReporter) Fatalf(string, ...any) {
	r.called = true
}

func (r *mockTestReporter) Cleanup(f func()) {
	r.t.Cleanup(f)
}

func Test_Provider_Begin_UnexpectedCall(t *testing.T) {
	provider := mocks.NewProvider(newMockTestReporter(t, true))

	tx, err := provider.Begin(context.Background())
	require.Nil(t, tx)
	require.NoError(t, err)
}

func Test_Provider_BeginTx_UnexpectedCall(t *testing.T) {
	provider := mocks.NewProvider(newMockTestReporter(t, true))

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{})
	require.Nil(t, tx)
	require.NoError(t, err)
}

func Test_Provider_ExpectBeginAndReturnError_Valid(t *testing.T) {
	beginError := errors.New("begin error")

	provider := mocks.NewProvider(newMockTestReporter(t, false))

	provider.ExpectBeginAndReturnError(beginError)

	tx, err := provider.Begin(context.Background())
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginError)
}

func Test_Provider_ExpectBeginAndReturnError_CalledTwice(t *testing.T) {
	beginError := errors.New("begin error")

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginAndReturnError(beginError)

	tx, err := provider.Begin(context.Background())
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginError)

	tx, err = provider.Begin(context.Background())
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginError)
}

func Test_Provider_ExpectBeginAndReturnError_Expect_But_Not_Called(t *testing.T) {
	beginError := errors.New("begin error")

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginAndReturnError(beginError)
}

func Test_Provider_ExpectBeginAndReturnError_CallBeginTx(t *testing.T) {
	beginError := errors.New("begin error")

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginAndReturnError(beginError)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{})
	require.Nil(t, tx)
	require.NoError(t, err)
}

func Test_Provider_ExpectBeginTxAndReturnError_Valid(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.NewProvider(newMockTestReporter(t, false))

	provider.ExpectBeginTxAndReturnError(beginTxError, opts)

	tx, err := provider.BeginTx(context.Background(), opts)
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_CalledTwice(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnError(beginTxError, opts)

	tx, err := provider.BeginTx(context.Background(), opts)
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginTxError)

	tx, err = provider.BeginTx(context.Background(), opts)
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_Call_With_Unexpected_Opts(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnError(beginTxError, opts)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{Isolation: sql.LevelDefault})
	require.Nil(t, tx)
	require.ErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_Expect_But_Not_Called(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnError(beginTxError, opts)
}

func Test_Provider_ExpectBeginTxAndReturnError_CalledBegin(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnError(beginTxError, opts)

	tx, err := provider.Begin(context.Background())
	require.Nil(t, tx)
	require.NoError(t, err)
}

func Test_Provider_ExpectBeginAndReturnTx_Valid(t *testing.T) {
	expectedTx := &mocks.Transaction{}

	provider := mocks.NewProvider(newMockTestReporter(t, false))

	provider.ExpectBeginAndReturnTx(expectedTx)

	tx, err := provider.Begin(context.Background())
	require.NoError(t, err)
	require.Equal(t, expectedTx, tx)
}

func Test_Provider_ExpectBeginAndReturnTx_CalledTwice(t *testing.T) {
	expectedTx := &mocks.Transaction{}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginAndReturnTx(expectedTx)

	tx, err := provider.Begin(context.Background())
	require.NoError(t, err)
	require.Equal(t, expectedTx, tx)

	tx, err = provider.Begin(context.Background())
	require.NoError(t, err)
	require.Equal(t, expectedTx, tx)
}

func Test_Provider_ExpectBeginAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	expectedTx := &mocks.Transaction{}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginAndReturnTx(expectedTx)
}

func Test_Provider_ExpectBeginAndReturnTx_CalledBeginTx(t *testing.T) {
	expectedTx := &mocks.Transaction{}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginAndReturnTx(expectedTx)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{})
	require.NoError(t, err)
	require.Nil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Valid(t *testing.T) {
	expectedTx := &mocks.Transaction{}
	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.NewProvider(newMockTestReporter(t, false))

	provider.ExpectBeginTxAndReturnTx(expectedTx, opts)

	tx, err := provider.BeginTx(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, expectedTx, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_CalledTwice(t *testing.T) {
	expectedTx := &mocks.Transaction{}
	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnTx(expectedTx, opts)

	tx, err := provider.BeginTx(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, expectedTx, tx)

	tx, err = provider.BeginTx(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, expectedTx, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Call_With_Unexpected_Opts(t *testing.T) {
	expectedTx := &mocks.Transaction{}
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnTx(expectedTx, opts)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{Isolation: sql.LevelDefault})
	require.Equal(t, expectedTx, tx)
	require.NoError(t, err)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	expectedTx := &mocks.Transaction{}
	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnTx(expectedTx, opts)
}

func Test_Provider_ExpectBeginTxAndReturnTx_CalledBegin(t *testing.T) {
	expectedTx := &mocks.Transaction{}
	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.NewProvider(newMockTestReporter(t, true))

	provider.ExpectBeginTxAndReturnTx(expectedTx, opts)

	tx, err := provider.Begin(context.Background())
	require.Nil(t, tx)
	require.NoError(t, err)
}
