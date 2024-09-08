package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

func Test_Transaction_Context(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	tx := mocks.ExpectNothing()(testReporter)

	ctx := tx.Context()
	require.True(t, mocks.TxEnabled.Matches(ctx))
	require.False(t, mocks.TxDisabled.Matches(ctx))
}

func Test_Transaction_Commit_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	tx := mocks.ExpectCommit()(testReporter)

	err := tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_Commit_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	tx := mocks.ExpectCommit()(testReporter)

	err := tx.Commit(context.Background())
	require.NoError(t, err)

	err = tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_Commit_CalledRollback(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	tx := mocks.ExpectCommit()(testReporter)

	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectCommit_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	mocks.ExpectCommit()(testReporter)
}

func Test_Transaction_ExpectRollback_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	errRollback := errors.New("rollback error")

	tx := mocks.ExpectRollback(errRollback)(testReporter)

	err := tx.Rollback(context.Background())
	require.ErrorIs(t, err, errRollback)
}

func Test_Transaction_ExpectRollback_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errRollback := errors.New("rollback error")

	tx := mocks.ExpectRollback(errRollback)(testReporter)

	err := tx.Rollback(context.Background())
	require.ErrorIs(t, err, errRollback)

	err = tx.Rollback(context.Background())
	require.ErrorIs(t, err, errRollback)
}

func Test_Transaction_ExpectRollback_CalledCommit(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errRollback := errors.New("rollback error")

	tx := mocks.ExpectRollback(errRollback)(testReporter)

	err := tx.Commit(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_ExpectRollback_Expected_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errRollback := errors.New("rollback error")

	mocks.ExpectRollback(errRollback)(testReporter)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	err = tx.Rollback(context.Background())
	require.NoError(t, err)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_RollbackFirst(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Rollback(context.Background())
	require.NoError(t, err)

	err = tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_OnlyCommit(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_OnlyRollback(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_CommitCalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	err = tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	tx.Rollback(context.Background())
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_RollbackCalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit(context.Background())
	require.ErrorIs(t, err, errCommit)

	tx.Rollback(context.Background())
	tx.Rollback(context.Background())
}
