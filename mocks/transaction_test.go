package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

func Test_Transaction_Rollback_UnexpectedCall(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.Rollback(context.Background())
}

func Test_Transaction_Commit_UnexpectedCall(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	err := tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_Context(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, false))

	ctx := tx.Context()
	require.True(t, mocks.TxEnabled.Matches(ctx))
	require.False(t, mocks.TxDisabled.Matches(ctx))
}

func Test_Transaction_Commit_Valid(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, false))

	tx.ExpectCommit()

	err := tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_Commit_CalledTwice(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.ExpectCommit()

	err := tx.Commit(context.Background())
	require.NoError(t, err)

	err = tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_Commit_CalledRollback(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.ExpectCommit()

	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectCommit_Expect_But_Not_Called(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.ExpectCommit()
}

func Test_Transaction_ExpectRollback_Valid(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, false))

	tx.ExpectRollback()

	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectRollback_CalledTwice(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.ExpectRollback()

	tx.Rollback(context.Background())
	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectRollback_CalledCommit(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.ExpectRollback()

	err := tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_ExpectRollback_Expected_But_Not_Called(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))

	tx.ExpectRollback()
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_Valid(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, false))
	errCommit := errors.New("failed commit")

	tx.ExpectRollbackAfterFailedCommit(errCommit)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_RollbackFirst(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))
	errCommit := errors.New("failed commit")

	tx.ExpectRollbackAfterFailedCommit(errCommit)

	tx.Rollback(context.Background())

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_OnlyCommit(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))
	errCommit := errors.New("failed commit")

	tx.ExpectRollbackAfterFailedCommit(errCommit)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_OnlyRollback(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))
	errCommit := errors.New("failed commit")

	tx.ExpectRollbackAfterFailedCommit(errCommit)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_CommitCalledTwice(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))
	errCommit := errors.New("failed commit")

	tx.ExpectRollbackAfterFailedCommit(errCommit)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	err = tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_RollbackCalledTwice(t *testing.T) {
	tx := mocks.NewTransaction(newMockTestReporter(t, true))
	errCommit := errors.New("failed commit")

	tx.ExpectRollbackAfterFailedCommit(errCommit)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	tx.Rollback(context.Background())
	tx.Rollback(context.Background())
}
